package mailboxpg

import "strings"

func insertSchemaStatementsAfter(statements []string, prefix string, additions []string) []string {
	if len(additions) == 0 {
		return statements
	}
	for i, statement := range statements {
		if strings.HasPrefix(strings.TrimSpace(statement), prefix) {
			out := make([]string, 0, len(statements)+len(additions))
			out = append(out, statements[:i+1]...)
			out = append(out, additions...)
			out = append(out, statements[i+1:]...)
			return out
		}
	}
	return append(statements, additions...)
}
