package mailboxprovider

func (r *Registry) StorageByKey(provider string) StorageExtension {
	if plugin := r.ByKey(provider); plugin != nil {
		return plugin
	}
	return nil
}

func (r *Registry) RetentionByKey(provider string) InboxRetentionPolicy {
	if plugin := r.ByKey(provider); plugin != nil {
		return plugin
	}
	return nil
}

func (r *Registry) CapabilityByKey(provider string) CapabilityPlugin {
	if plugin := r.ByKey(provider); plugin != nil {
		return plugin
	}
	return nil
}
