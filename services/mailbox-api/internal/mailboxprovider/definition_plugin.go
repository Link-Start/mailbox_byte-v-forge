package mailboxprovider

type definitionPlugin struct {
	definition Definition
}

func NewDefinitionPlugin(definition Definition) Plugin {
	definition.ProviderKey = NormalizeKey(definition.ProviderKey)
	for index, alias := range definition.AliasKeys {
		definition.AliasKeys[index] = NormalizeKey(alias)
	}
	return definitionPlugin{definition: definition}
}

func (p definitionPlugin) Key() string { return p.definition.ProviderKey }

func (p definitionPlugin) Aliases() []string {
	return append([]string{}, p.definition.AliasKeys...)
}

func (p definitionPlugin) DisplayName() string { return p.definition.DisplayNameValue }

func (p definitionPlugin) StoredInboxOnly() bool { return p.definition.StoredInboxOnlyValue }
