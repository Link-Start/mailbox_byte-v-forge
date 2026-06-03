package mailboxprovider

import (
	"fmt"
	"strings"

	"mailboxapi/internal/mailboxmodel"
)

func (r *Registry) NormalizeProviderInput(provider string) string {
	value := NormalizeKey(provider)
	if value == "" {
		return ""
	}
	if definition := r.ByKey(value); definition != nil {
		return definition.Key()
	}
	return value
}

func (r *Registry) ValidatePoll(row MailboxRecord) error {
	if strings.TrimSpace(row.Email) == "" {
		return fmt.Errorf("mailbox is required")
	}
	definition := r.StorageByKey(row.Provider)
	if definition == nil || !definition.CanValidatePoll() {
		return fmt.Errorf("mailbox provider cannot poll inbox: %s", row.Provider)
	}
	return definition.ValidatePoll(row)
}

func (r *Registry) PrepareProjection(mailbox *mailboxmodel.Record) {
	if mailbox == nil {
		return
	}
	if definition := r.CapabilityByKey(mailbox.GetProviderKey()); definition != nil {
		definition.PrepareProjection(mailbox)
	}
}

func (r *Registry) StorageByKey(provider string) StorageExtension {
	if plugin := r.ByKey(provider); plugin != nil {
		return plugin
	}
	return nil
}

func (r *Registry) RetentionByKey(provider string) InboxRetentionPolicy {
	if plugin := r.ByKey(provider); plugin != nil {
		return plugin
	}
	return nil
}

func (r *Registry) CapabilityByKey(provider string) CapabilityPlugin {
	if plugin := r.ByKey(provider); plugin != nil {
		return plugin
	}
	return nil
}
