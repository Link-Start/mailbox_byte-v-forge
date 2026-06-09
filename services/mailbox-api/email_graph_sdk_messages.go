package main

import graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"

func graphMessagesFromSDK(messages []graphmodels.Messageable) []graphMessage {
	out := make([]graphMessage, 0, len(messages))
	for _, message := range messages {
		msg, ok := graphMessageFromSDK(message)
		if !ok {
			continue
		}
		out = append(out, msg)
	}
	return out
}

func graphMessageFromSDK(message graphmodels.Messageable) (graphMessage, bool) {
	if message == nil {
		return graphMessage{}, false
	}
	return graphMessage{
		ID:                     stringValueFromPtr(message.GetId()),
		Subject:                stringValueFromPtr(message.GetSubject()),
		From:                   graphRecipientFromSDK(message.GetFrom()),
		BodyPreview:            stringValueFromPtr(message.GetBodyPreview()),
		Body:                   graphBodyFromSDK(message.GetBody()),
		ToRecipients:           graphRecipientsFromSDK(message.GetToRecipients()),
		CcRecipients:           graphRecipientsFromSDK(message.GetCcRecipients()),
		BccRecipients:          graphRecipientsFromSDK(message.GetBccRecipients()),
		InternetMessageHeaders: graphHeadersFromSDK(message.GetInternetMessageHeaders()),
		ReceivedDateTime:       graphTimeFromSDK(message.GetReceivedDateTime()),
	}, true
}

func graphBodyFromSDK(body graphmodels.ItemBodyable) graphBody {
	if body == nil {
		return graphBody{}
	}
	contentType := ""
	if body.GetContentType() != nil {
		contentType = body.GetContentType().String()
	}
	return graphBody{Content: stringValueFromPtr(body.GetContent()), ContentType: contentType}
}

func graphRecipientFromSDK(recipient graphmodels.Recipientable) graphRecipient {
	if recipient == nil || recipient.GetEmailAddress() == nil {
		return graphRecipient{}
	}
	return graphRecipient{EmailAddress: graphEmailAddress{Address: stringValueFromPtr(recipient.GetEmailAddress().GetAddress())}}
}

func graphRecipientsFromSDK(recipients []graphmodels.Recipientable) []graphRecipient {
	out := make([]graphRecipient, 0, len(recipients))
	for _, recipient := range recipients {
		out = append(out, graphRecipientFromSDK(recipient))
	}
	return out
}

func graphHeadersFromSDK(headers []graphmodels.InternetMessageHeaderable) []graphHeader {
	out := make([]graphHeader, 0, len(headers))
	for _, header := range headers {
		if header == nil {
			continue
		}
		out = append(out, graphHeader{
			Name:  stringValueFromPtr(header.GetName()),
			Value: stringValueFromPtr(header.GetValue()),
		})
	}
	return out
}
