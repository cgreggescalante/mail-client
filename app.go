package main

import (
	"bytes"
	"context"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message"
	"log"
	"strings"
)

const EMAIL = "conor@johngregg.org"
const PASSWORD = "HY2tn2B8Bnp2aXn_"

type App struct {
	ctx             context.Context
	accounts        []Account
	messages        []imap.Message
	loadedMailboxes bool
}

func NewApp() *App {
	return &App{
		accounts: []Account{{
			EMAIL,
			[]imap.MailboxInfo{},
		}},
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

type Account struct {
	Email     string `json:"email"`
	Mailboxes []imap.MailboxInfo
}

type Mailbox struct {
	Name string `json:"name"`
}

type Message struct {
	Subject string `json:"subject"`
	From    string `json:"from"`
	Uid     uint32 `json:"uid"`
	Body    string `json:"body"`
}

func (a *App) LoadMailboxes() {
	c, err := client.DialTLS("mail.hover.com:993", nil)
	if err != nil {
		log.Printf("Error connecting to mail server: %v", err)
		return
	}

	defer c.Logout()

	if err := c.Login(EMAIL, PASSWORD); err != nil {
		log.Printf("Error logging in: %v", err)
		return
	}

	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	for m := range mailboxes {
		a.accounts[0].Mailboxes = append(a.accounts[0].Mailboxes, *m)
	}

	if err := <-done; err != nil {
		log.Printf("Error listing mailboxes: %v", err)
	}

	a.loadedMailboxes = true
}

func (a *App) GetMessage(Uid uint32) Message {
	for _, message := range a.messages {
		if message.Uid == Uid {
			return Message{
				Subject: message.Envelope.Subject,
				From:    message.Envelope.From[0].Address(),
			}
		}
	}

	return Message{
		Subject: "Not",
		From:    "Found",
	}
}

func (a *App) GetImapMessages() []Message {
	c, err := client.DialTLS("mail.hover.com:993", nil)
	if err != nil {
		return []Message{}
	}

	defer c.Logout()

	if err := c.Login(EMAIL, PASSWORD); err != nil {
		return []Message{}
	}

	mbox, err := c.Select("INBOX", false)
	if err != nil {
		return []Message{}
	}
	log.Printf("Flags %v", mbox.Flags)

	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 1 {
		from = mbox.Messages - 1
	}

	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)
	section := &imap.BodySectionName{}
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{section.FetchItem(), imap.FetchBody, imap.FetchUid, imap.FetchEnvelope, imap.FetchBodyStructure}, messages)
	}()

	for msg := range messages {
		a.messages = append(a.messages, *msg)
	}

	return ImapToSimpleMessages(a.messages)
}

func ImapToSimpleMessage(msg imap.Message) Message {
	r := msg.GetBody(&imap.BodySectionName{})
	if r == nil {
		log.Println("Error getting body")
	}

	read, err := message.Read(r)
	if err != nil {
		log.Println("Error reading message")
	}

	body := ""

	if read.MultipartReader() != nil {
		for part, err := read.MultipartReader().NextPart(); err == nil; part, err = read.MultipartReader().NextPart() {
			if strings.Contains(part.Header.Get("Content-Type"), "text/plain") {
				buf := new(bytes.Buffer)
				buf.ReadFrom(part.Body)
				body = buf.String()
				break
			}
		}
	} else {
		log.Println("Single part message")
		buf := new(bytes.Buffer)
		buf.ReadFrom(read.Body)
		body = buf.String()
	}

	return Message{
		Subject: msg.Envelope.Subject,
		From:    msg.Envelope.From[0].Address(),
		Uid:     msg.Uid,
		Body:    body,
	}
}

func ImapToSimpleMessages(messages []imap.Message) []Message {
	result := make([]Message, len(messages))

	for i, msg := range messages {
		result[i] = ImapToSimpleMessage(msg)
	}

	return result
}

func ImapToSimpleMailboxes(mailboxes []imap.MailboxInfo) []Mailbox {
	result := make([]Mailbox, len(mailboxes))
	for i, mbox := range mailboxes {
		result[i] = Mailbox{
			Name: mbox.Name,
		}
	}

	return result
}
