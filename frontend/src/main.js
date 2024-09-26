import './style.css';
import './app.css';

import {GetImapMessages, GetMailboxTree, GetMessage} from '../wailsjs/go/main/App';

document.querySelector('#app').innerHTML = `
    <div id="mailboxesColumn">
        <h4>Mailboxes</h4>
        <div id="mailboxes"></div>
    </div>
    <div id="messagesColumn">
        <h4>Messages</h4>
        <div id="messages"></div>
    </div>
    <div id="messageColumn">
        <h4>Message</h4>
    <div id="message"></div>
    </div>
`;

let mailboxList = document.getElementById("mailboxes");
let messageList = document.getElementById("messages");
let messageContent = document.getElementById("message");

window.onload = function () {
    GetMailboxTree().then((result) => {
        mailboxList.innerHTML = result
    }).catch((err) => console.error(err));
}

window.getMessage = function (uid) {
    console.log(uid);
    GetMessage(uid)
        .then((result) => {
            messageContent.innerHTML = `
                <h4>${result.subject}</h4>
                <p>${result.from}</p>
                <p>${result.body}</p>
            `;
        })
        .catch((err) => {
            console.error(err);
        });
}

window.getMailboxContents = function (mailbox) {
    GetImapMessages()
        .then((result) => {
            messageList.innerHTML = `${mailbox}` + result.map(
                (message) => `
                    <div onclick="getMessage(${message.uid})">${message.subject} - ${message.from}</div>
                `
            ).join('');
        })
        .catch((err) => {
            console.error(err);
        });
}

