import './style.css';
import './app.css';

import {GetMailboxTree, GetMessage, GetMessageList} from '../wailsjs/go/main/App';

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
        <div id="messageHeader"></div>
        <div id="messageBody" style="background-color: white; color: black"></div>
    </div>
`;

let mailboxList = document.getElementById("mailboxes");
let messageList = document.getElementById("messages");
let messageHeader = document.getElementById("messageHeader");
let messageBody = document.getElementById("messageBody");
let shadowRoot = messageBody.attachShadow({mode: 'closed'});

window.onload = () => {
    GetMailboxTree().then((result) => mailboxList.innerHTML = result).catch((err) => console.error(err));
}
window.getMessage = (uid) => GetMessage(uid).then((result) => {
    messageHeader.innerHTML = result.Message
    shadowRoot.innerHTML = `<style>blockquote {margin-inline-end: 0; margin-inline-start: 5px; padding-inline-start: 3px; border-left: 1px solid black}</style>` + result.Body
}).catch((err) => console.error(err));

window.getMailboxContents = (mailbox) => GetMessageList(mailbox).then((result) => messageList.innerHTML = result).catch((err) => console.error(err));
