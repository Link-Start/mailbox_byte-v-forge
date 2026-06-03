package mailboxprovider

import (
	"fmt"
	"strings"
)

type Registry struct {
	ordered []Plugin
	byKey   map[string]Plugin
}

func NewRegistry(plugins ...Plugin) (*Registry, error) {
	registry := &Registry{byKey: map[string]Plugin{}}
	for _, plugin := range plugins {
		if plugin == nil {
			return nil, fmt.Errorf("mailbox provider plugin is required")
		}
		key := NormalizeKey(plugin.Key())
		if key == "" {
			return nil, fmt.Errorf("mailbox provider plugin key is required")
		}
		if _, exists := registry.byKey[key]; exists {
			return nil, fmt.Errorf("duplicate mailbox provider plugin %q", key)
		}
		registry.byKey[key] = plugin
		for _, alias := range plugin.Aliases() {
			alias = NormalizeKey(alias)
			if alias == "" {
				continue
			}
			if _, exists := registry.byKey[alias]; exists {
				return nil, fmt.Errorf("duplicate mailbox provider alias %q", alias)
			}
			registry.byKey[alias] = plugin
		}
		registry.ordered = append(registry.ordered, plugin)
	}
	return registry, nil
}

func (r *Registry) All() []Plugin {
	if r == nil {
		return nil
	}
	return append([]Plugin{}, r.ordered...)
}

func (r *Registry) CapabilityPlugins() []CapabilityPlugin {
	if r == nil {
		return nil
	}
	plugins := make([]CapabilityPlugin, 0, len(r.ordered))
	for _, plugin := range r.ordered {
		plugins = append(plugins, plugin)
	}
	return plugins
}

func (r *Registry) StorageExtensions() []StorageExtension {
	if r == nil {
		return nil
	}
	plugins := make([]StorageExtension, 0, len(r.ordered))
	for _, plugin := range r.ordered {
		plugins = append(plugins, plugin)
	}
	return plugins
}

func (r *Registry) InboxRetentionPolicies() []InboxRetentionPolicy {
	if r == nil {
		return nil
	}
	plugins := make([]InboxRetentionPolicy, 0, len(r.ordered))
	for _, plugin := range r.ordered {
		plugins = append(plugins, plugin)
	}
	return plugins
}

func (r *Registry) VirtualMailboxSources() []VirtualMailboxSource {
	if r == nil {
		return nil
	}
	plugins := make([]VirtualMailboxSource, 0, len(r.ordered))
	for _, plugin := range r.ordered {
		plugins = append(plugins, plugin)
	}
	return plugins
}

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
