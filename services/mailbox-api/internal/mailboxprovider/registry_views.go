package mailboxprovider

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
