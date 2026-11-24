package config

// ModelMetadataName returns the configured model name or a fixed marker for an unconfigured model.
func (config *Config) ModelMetadataName(model string) string {
	if _, configured := config.Models[model]; configured {
		return model
	}

	return "unconfigured"
}
