package httpsign

// Config contains additional configuration for the signer.
type Config struct {
	signAlg         bool
	signCreated     bool
	fakeCreated     int64
	requestResponse struct{ name, signature string }
}

// NewConfig generates a default configuration.
func NewConfig() Config {
	return Config{
		signAlg:     true,
		signCreated: true,
		fakeCreated: 0,
	}
}

// SignAlg indicates that an "alg" signature parameters must be generated and signed (default: true).
func (c Config) SignAlg(b bool) Config {
	c.signAlg = b
	return c
}

// SignCreated indicates that a "created" signature parameters must be generated and signed (default: true).
func (c Config) SignCreated(b bool) Config {
	c.signCreated = b
	return c
}

// setFakeCreated indicates that the specified Unix timestamp must be used instead of the current time
// (default: 0, meaning use current time). Only used for testing.
func (c Config) setFakeCreated(ts int64) Config {
	c.fakeCreated = ts
	return c
}

// SetRequestResponse allows the server to indicate signature name and signature that
// it had received from a client and include it in the signature input
func (c Config) SetRequestResponse(name, signature string) Config {
	c.requestResponse = struct{ name, signature string }{name, signature}
	return c
}