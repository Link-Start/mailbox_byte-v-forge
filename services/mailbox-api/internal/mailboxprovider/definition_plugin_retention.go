package mailboxprovider

func (p definitionPlugin) RetentionPolicy() (MessageRetention, bool) {
	policy := p.definition.RetentionPolicyValue
	return policy, policy.HasRetention()
}

func (p definitionPlugin) IncludeVirtual(authStatus string) bool {
	return p.definition.IncludeVirtualFunc == nil || p.definition.IncludeVirtualFunc(authStatus)
}
