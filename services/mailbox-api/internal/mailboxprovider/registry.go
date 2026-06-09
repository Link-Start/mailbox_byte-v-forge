package mailboxprovider

import "fmt"

type Registry struct {
	ordered []Plugin
	byKey   map[string]Plugin
}

func NewRegistry(plugins ...Plugin) (*Registry, error) {
	registry := &Registry{byKey: map[string]Plugin{}}
	for _, plugin := range plugins {
		if err := registry.register(plugin); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func (r *Registry) register(plugin Plugin) error {
	if plugin == nil {
		return fmt.Errorf("mailbox provider plugin is required")
	}
	key := NormalizeKey(plugin.Key())
	if key == "" {
		return fmt.Errorf("mailbox provider plugin key is required")
	}
	if _, exists := r.byKey[key]; exists {
		return fmt.Errorf("duplicate mailbox provider plugin %q", key)
	}
	if fields, ok := plugin.TokenFields(); ok {
		if err := validateTokenFields(fields); err != nil {
			return fmt.Errorf("invalid mailbox provider %q token fields: %w", key, err)
		}
	}
	r.byKey[key] = plugin
	if err := r.registerAliases(plugin); err != nil {
		return err
	}
	r.ordered = append(r.ordered, plugin)
	return nil
}

func (r *Registry) registerAliases(plugin Plugin) error {
	for _, alias := range plugin.Aliases() {
		alias = NormalizeKey(alias)
		if alias == "" {
			continue
		}
		if _, exists := r.byKey[alias]; exists {
			return fmt.Errorf("duplicate mailbox provider alias %q", alias)
		}
		r.byKey[alias] = plugin
	}
	return nil
}
