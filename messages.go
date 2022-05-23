package main

// oxNotification is the message sent by dovecot as a put request
type oxNotification struct {
	User            string `json:"user"`
	ImapUidvalidity int    `json:"imap-uidvalidity"`
	ImapUID         int    `json:"imap-uid"`
	Folder          string `json:"folder"`
	Event           string `json:"event"`
	From            string `json:"from"`
	Subject         string `json:"subject"`
	Snippet         string `json:"snippet"`
	Unseen          int    `json:"unseen"`
}

// rcResponse is the response our plugin sends if there is a valid session
type rcResponse struct {
	Username string `json:"username"`
}
