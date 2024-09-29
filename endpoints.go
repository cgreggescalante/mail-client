package main

import (
	"bytes"
	"html/template"
	"log"
)

type MailboxTreeData struct {
	Accounts []MailboxTreeAccount
}

type MailboxTreeAccount struct {
	Email     string
	Mailboxes []string
}

func (a *App) GetMailboxTree() string {
	if !a.loadedMailboxes {
		a.LoadMailboxes()
	}

	tmpl := template.Must(template.ParseFiles("./templates/mailbox-tree.gohtml"))
	var rendered bytes.Buffer
	err := tmpl.Execute(&rendered, MailboxTreeData{Accounts: []MailboxTreeAccount{{Email: a.accounts[0].Email, Mailboxes: []string{"INBOX"}}}})
	if err != nil {
		return "Could not render templates"
	}
	return rendered.String()
}

type MessageListData struct {
	Mailbox  string
	Messages []Message
}

func (a *App) GetMessageList(mailbox string) string {
	if !a.loadedMessages {
		a.LoadMessages()
	}

	tmpl := template.Must(template.ParseFiles("./templates/message-list.gohtml"))
	var rendered bytes.Buffer
	err := tmpl.Execute(&rendered, MessageListData{Mailbox: mailbox, Messages: a.messages})
	if err != nil {
		return "Could not render template"
	}
	return rendered.String()
}

type MessageData struct {
	Message
	HtmlBody template.HTML
}

type GetMessageData struct {
	Message string
	Body    string
}

func (a *App) GetMessage(uid uint32) GetMessageData {
	selectedMessage := MessageData{}
	found := false

	for _, msg := range a.messages {
		if msg.Uid == uid {
			selectedMessage = MessageData{msg, ""}
			found = true
		}
	}

	if !found {
		return GetMessageData{"Could not find message", ""}
	}

	tmpl := template.Must(template.ParseFiles("./templates/message.gohtml"))
	var rendered bytes.Buffer
	err := tmpl.Execute(&rendered, selectedMessage)
	if err != nil {
		log.Println(err)
		return GetMessageData{"Could not render template", ""}
	}
	return GetMessageData{rendered.String(), selectedMessage.Body}
}
