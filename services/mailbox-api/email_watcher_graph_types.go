package main

type graphMessage struct {
	ID                     string
	Subject                string
	From                   graphRecipient
	BodyPreview            string
	Body                   graphBody
	ToRecipients           []graphRecipient
	CcRecipients           []graphRecipient
	BccRecipients          []graphRecipient
	InternetMessageHeaders []graphHeader
	ReceivedDateTime       string
}

type graphBody struct {
	Content     string
	ContentType string
}

type graphRecipient struct {
	EmailAddress graphEmailAddress
}

type graphEmailAddress struct {
	Address string
}

type graphHeader struct {
	Name  string
	Value string
}
