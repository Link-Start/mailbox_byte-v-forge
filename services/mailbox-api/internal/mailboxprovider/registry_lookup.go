package mailboxprovider

import "strings"

func (r *Registry) DefaultKey() string {
	if r == nil || len(r.ordered) == 0 {
		return ""
	}
	return r.ordered[0].Key()
}

func (r *Registry) ByKey(provider string) Plugin {
	if r == nil {
		return nil
	}
	return r.byKey[NormalizeKey(provider)]
}

func NormalizeKey(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}
