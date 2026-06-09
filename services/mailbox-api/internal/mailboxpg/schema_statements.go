package mailboxpg

func baseSchemaStatements(defaultProvider string) []string {
	statements := mailboxSchemaStatements(defaultProvider)
	statements = append(statements, inboxSchemaStatements(defaultProvider)...)
	statements = append(statements, legacySchemaStatements()...)
	statements = append(statements, indexSchemaStatements()...)
	return statements
}
