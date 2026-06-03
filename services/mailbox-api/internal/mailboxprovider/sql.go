package mailboxprovider

func (r *Registry) SchemaStatements() []string {
	statements := []string{}
	for _, provider := range r.StorageExtensions() {
		statements = append(statements, provider.SchemaStatements()...)
	}
	return statements
}

func (r *Registry) LegacyStatements() []string {
	statements := []string{}
	for _, provider := range r.StorageExtensions() {
		statements = append(statements, provider.PrepareLegacyData()...)
	}
	return statements
}
