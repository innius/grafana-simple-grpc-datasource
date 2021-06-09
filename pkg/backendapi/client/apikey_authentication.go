package client

import "context"

const apiKeyHeader = "api-key"

type apikeyAuthentication struct {
	apiKey string
}

func (t apikeyAuthentication) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		apiKeyHeader: t.apiKey,
	}, nil
}

func (apikeyAuthentication) RequireTransportSecurity() bool {
	return true
}
