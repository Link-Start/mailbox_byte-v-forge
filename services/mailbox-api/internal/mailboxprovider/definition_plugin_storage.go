package mailboxprovider

func (p definitionPlugin) SchemaStatements() []string {
	if p.definition.SchemaStatementsFunc == nil {
		return nil
	}
	return p.definition.SchemaStatementsFunc()
}

func (p definitionPlugin) CanValidatePoll() bool { return p.definition.ValidatePollFunc != nil }

func (p definitionPlugin) ValidatePoll(row MailboxRecord) error {
	return p.definition.ValidatePollFunc(row)
}

func (p definitionPlugin) TokenFields() (TokenFields, bool) {
	fields := p.definition.TokenFieldsValue
	return fields, fields.HasTokenStorage()
}

func (p definitionPlugin) PrepareLegacyData() []string {
	if p.definition.PrepareLegacyDataFunc == nil {
		return nil
	}
	return p.definition.PrepareLegacyDataFunc()
}
