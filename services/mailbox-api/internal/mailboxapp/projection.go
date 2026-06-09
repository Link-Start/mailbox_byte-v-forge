package mailboxapp

import (
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/mailboxmodel"
)

func PublicMailbox(mailbox *mailboxmodel.Record) *mailboxv1.EmailMailbox {
	if mailbox == nil {
		return nil
	}
	return &mailboxv1.EmailMailbox{
		EmailAddress:    mailbox.GetEmailAddress(),
		LastError:       mailbox.GetLastError(),
		CreatedAt:       mailbox.GetCreatedAt(),
		UpdatedAt:       mailbox.GetUpdatedAt(),
		AuthStatus:      mailboxmodel.PublicAuthStatus(mailbox.GetAuthStatus()),
		ProviderKey:     mailbox.GetProviderKey(),
		LatestSignal:    mailbox.GetLatestSignal(),
		Domain:          mailbox.GetDomain(),
		CredentialState: credentialState(mailbox),
	}
}

func PublicMailboxList(mailboxes []*mailboxmodel.Record) []*mailboxv1.EmailMailbox {
	out := make([]*mailboxv1.EmailMailbox, 0, len(mailboxes))
	for _, mailbox := range mailboxes {
		if public := PublicMailbox(mailbox); public != nil {
			out = append(out, public)
		}
	}
	return out
}
