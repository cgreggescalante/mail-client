package main

import (
	"bytes"
	"html/template"
)

type MailboxTreeData struct {
	Accounts []Account
}

func (a *App) GetMailboxTree() string {
	if !a.loadedMailboxes {
		a.LoadMailboxes()
	}

	tmpl := template.Must(template.ParseFiles("./templates/mailbox-tree.gohtml"))
	var rendered bytes.Buffer
	err := tmpl.Execute(&rendered, MailboxTreeData{Accounts: a.accounts})
	if err != nil {
		return "Could not render mailbox tree"
	}
	return rendered.String()
}

//func GetMessageList() string {
//
//}
//
//func GetMessage(uid uint32) string {
//
//}
