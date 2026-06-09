package mailboxprovider

import (
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

	"mailboxapi/internal/mailboxmodel"
)

func (p definitionPlugin) Capabilities() *mailboxv1.MailboxProviderCapabilities {
	if p.definition.CapabilitiesFunc == nil {
		return nil
	}
	return p.definition.CapabilitiesFunc()
}

func (p definitionPlugin) LoadDomains() []string {
	if p.definition.LoadDomainsFunc == nil {
		return nil
	}
	return p.definition.LoadDomainsFunc()
}

func (p definitionPlugin) Domains(configured []string) []*mailboxv1.MailboxDomain {
	if p.definition.DomainsFunc == nil {
		return nil
	}
	return p.definition.DomainsFunc(configured)
}

func (p definitionPlugin) MatchesAddress(email string, cfg RuntimeContext) bool {
	return p.definition.MatchesAddressFunc != nil && p.definition.MatchesAddressFunc(email, cfg)
}

func (p definitionPlugin) PrepareProjection(mailbox *mailboxmodel.Record) {
	if p.definition.PrepareProjectionFunc != nil {
		p.definition.PrepareProjectionFunc(mailbox)
	}
}
