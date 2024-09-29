package main

import (
	"bytes"
	"context"
	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"io"
	"log"
	"mime"
	"strings"
	"time"
)

type App struct {
	store           *Store
	config          Config
	ctx             context.Context
	accounts        []Account
	messages        []Message
	loadedMailboxes bool
	loadedMessages  bool
}

type Account struct {
	Email     string `json:"email"`
	Mailboxes []imap.ListData
}

type Message struct {
	Uid     uint32
	Date    time.Time
	From    []*mail.Address
	To      []*mail.Address
	ReplyTo []*mail.Address
	Cc      []*mail.Address
	Bcc     []*mail.Address
	Subject string
	Body    string
}

func NewApp(store *Store, config Config) *App {
	creds := store.GetAccount("conor@johngregg.org")

	return &App{
		store:  store,
		config: config,
		accounts: []Account{{
			creds.Email,
			[]imap.ListData{},
		}},
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) LoadMailboxes() {
	c, err := imapclient.DialTLS("mail.hover.com:993", nil)
	if err != nil {
		log.Printf("Error connecting to mail server: %v", err)
		return
	}

	defer c.Logout()

	creds := a.store.GetAccount("conor@johngregg.org")

	if err := c.Login(creds.Email, creds.Password).Wait(); err != nil {
		log.Printf("Error logging in: %v", err)
		return
	}

	listCmd := c.List("", "*", &imap.ListOptions{
		ReturnStatus: &imap.StatusOptions{
			NumMessages: true,
			NumUnseen:   true,
		},
	})

	mailboxes, err := listCmd.Collect()
	if err != nil {
		log.Printf("Error listing mailboxes: %v", err)
	}

	for _, m := range mailboxes {
		a.accounts[0].Mailboxes = append(a.accounts[0].Mailboxes, *m)
	}

	a.loadedMailboxes = true
}

func (a *App) LoadMessages() {
	c, err := imapclient.DialTLS("mail.hover.com:993", &imapclient.Options{
		WordDecoder: &mime.WordDecoder{CharsetReader: charset.Reader},
	})
	if err != nil {
		log.Printf("Error connecting to mail server: %v", err)
		return
	}

	defer c.Logout()

	creds := a.store.GetAccount("conor@johngregg.org")

	if err := c.Login(creds.Email, creds.Password).Wait(); err != nil {
		log.Printf("Error logging in: %v", err)
		return
	}

	mailbox := imap.ListData{}
	ok := false
	for _, m := range a.accounts[0].Mailboxes {
		if m.Mailbox == "INBOX" {
			mailbox = m
			ok = true
			break
		}
	}

	if !ok {
		log.Println("INBOX mailbox not found")
		return
	}

	c.Select(mailbox.Mailbox, nil).Wait()

	seqSet := imap.SeqSetNum(1, 2, 3, 4, 5)

	fetchOptions := &imap.FetchOptions{
		UID:         true,
		BodySection: []*imap.FetchItemBodySection{{}},
	}
	fetchCmd := c.Fetch(seqSet, fetchOptions)
	defer func(fetchCmd *imapclient.FetchCommand) {
		err := fetchCmd.Close()
		if err != nil {
			log.Printf("Error closing FETCH command: %v", err)
		}
	}(fetchCmd)

	messageBuffer, err := fetchCmd.Collect()
	if err != nil {
		log.Printf("Error fetching messages: %v", err)
		return
	}

	for _, msg := range messageBuffer {
		log.Printf("UID: %v", msg.UID)
		message := Message{Uid: uint32(msg.UID)}

		bodySection := imap.FetchItemBodySection{}
		bodyBytes := []byte{}
		ok := false

		for section, sectionBytes := range msg.BodySection {
			bodySection = *section
			bodyBytes = sectionBytes
			ok = true
		}

		if !ok {
			log.Printf("FETCH command did not return body section")
			continue
		}

		log.Printf("Body bytes: %v", len(bodyBytes))

		for i, field := range bodySection.HeaderFields {
			log.Printf("Header field %d: %v", i, field)
		}
		mr, err := mail.CreateReader(bytes.NewReader(bodyBytes))
		if err != nil {
			log.Printf("failed to create mail reader: %v", err)
		}

		h := mr.Header
		if date, err := h.Date(); err != nil {
			log.Printf("failed to parse Date header field: %v", err)
		} else {
			message.Date = date
		}

		if to, err := h.AddressList("To"); err != nil {
			log.Printf("failed to parse To header field: %v", err)
		} else {
			message.To = to
		}
		if from, err := h.AddressList("From"); err != nil {
			log.Printf("failed to parse From header field: %v", err)
		} else {
			message.From = from
		}
		if replyTo, err := h.AddressList("Reply-To"); err != nil {
			log.Printf("failed to parse Reply-To header field: %v", err)
		} else {
			message.ReplyTo = replyTo
		}
		if cc, err := h.AddressList("Cc"); err != nil {
			log.Printf("failed to parse Cc header field: %v", err)
		} else {
			message.Cc = cc
		}
		if bcc, err := h.AddressList("Bcc"); err != nil {
			log.Printf("failed to parse Bcc header field: %v", err)
		} else {
			message.Bcc = bcc
		}

		if subject, err := h.Text("Subject"); err != nil {
			log.Printf("failed to parse Subject header field: %v", err)
		} else {
			message.Subject = subject
		}

		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatalf("failed to read message part: %v", err)
			}

			switch h := p.Header.(type) {
			case *mail.InlineHeader:
				if message.Body != "" || strings.Contains(h.Get("Content-Type"), "text/plain") {
					continue
				}
				b, _ := io.ReadAll(p.Body)
				message.Body = string(b)
			case *mail.AttachmentHeader:
				filename, _ := h.Filename()
				log.Printf("Attachment: %v", filename)
			}
		}

		a.messages = append(a.messages, message)
	}

	//for {
	//	msg := fetchCmd.Next()
	//	if msg == nil {
	//		break
	//	}
	//
	//	log.Println(msg.SeqNum)
	//
	//	message := Message{}
	//
	//	// Find the body section in the response
	//	var bodySection imapclient.FetchItemDataBodySection
	//	ok = false
	//	for {
	//		item := msg.Next()
	//		if item == nil {
	//			break
	//		}
	//		switch item := item.(type) {
	//		case *imapclient.FetchItemDataBodySection:
	//			bodySection = *item
	//			ok = true
	//		case imapclient.FetchItemDataUID:
	//			message.Uid = uint32(item.UID)
	//		}
	//		bodySection, ok = item.(imapclient.FetchItemDataBodySection)
	//		if ok {
	//			break
	//		}
	//	}
	//	if !ok {
	//		log.Fatalf("FETCH command did not return body section")
	//	}
	//
	//	// Read the message via the go-message library
	//	mr, err := mail.CreateReader(bodySection.Literal)
	//	if err != nil {
	//		log.Fatalf("failed to create mail reader: %v", err)
	//	}
	//
	//	// Print a few header fields
	//	h := mr.Header
	//	if date, err := h.Date(); err != nil {
	//		log.Printf("failed to parse Date header field: %v", err)
	//	} else {
	//		message.Date = date
	//	}
	//
	//	if to, err := h.AddressList("To"); err != nil {
	//		log.Printf("failed to parse To header field: %v", err)
	//	} else {
	//		message.To = to
	//	}
	//	if from, err := h.AddressList("From"); err != nil {
	//		log.Printf("failed to parse From header field: %v", err)
	//	} else {
	//		message.From = from
	//	}
	//	if replyTo, err := h.AddressList("Reply-To"); err != nil {
	//		log.Printf("failed to parse Reply-To header field: %v", err)
	//	} else {
	//		message.ReplyTo = replyTo
	//	}
	//	if cc, err := h.AddressList("Cc"); err != nil {
	//		log.Printf("failed to parse Cc header field: %v", err)
	//	} else {
	//		message.Cc = cc
	//	}
	//	if bcc, err := h.AddressList("Bcc"); err != nil {
	//		log.Printf("failed to parse Bcc header field: %v", err)
	//	} else {
	//		message.Bcc = bcc
	//	}
	//
	//	if subject, err := h.Text("Subject"); err != nil {
	//		log.Printf("failed to parse Subject header field: %v", err)
	//	} else {
	//		message.Subject = subject
	//	}
	//
	//	// Process the message's parts
	//	for {
	//		p, err := mr.NextPart()
	//		if err == io.EOF {
	//			break
	//		} else if err != nil {
	//			log.Fatalf("failed to read message part: %v", err)
	//		}
	//
	//		switch h := p.Header.(type) {
	//		case *mail.InlineHeader:
	//			if message.Body != "" || strings.Contains(h.Get("Content-Type"), "text/plain") {
	//				continue
	//			}
	//			b, _ := io.ReadAll(p.Body)
	//			message.Body = string(b)
	//		case *mail.AttachmentHeader:
	//			filename, _ := h.Filename()
	//			log.Printf("Attachment: %v", filename)
	//		}
	//	}
	//
	//	if err := fetchCmd.Close(); err != nil {
	//		log.Fatalf("FETCH command failed: %v", err)
	//	}
	//
	//	a.messages = append(a.messages, message)
	//}

	a.loadedMessages = true
}
