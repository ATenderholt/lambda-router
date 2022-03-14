package settings

import "context"

const configContextKey = contextKey("config")

type contextKey string

func (config *Config) NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, configContextKey, config)
}

func FromContext(ctx context.Context) *Config {
	cfg, ok := ctx.Value(configContextKey).(*Config)
	if ok {
		return cfg
	}

	return DefaultConfig()
}
