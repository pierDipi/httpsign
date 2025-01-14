package httpsign

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type requestResponse struct{ name, signature string }

// SignConfig contains additional configuration for the signer.
type SignConfig struct {
	signAlg         bool
	signCreated     bool
	fakeCreated     int64
	expires         int64
	nonce           string
	requestResponse *requestResponse
}

// NewSignConfig generates a default configuration.
func NewSignConfig() *SignConfig {
	return &SignConfig{
		signAlg:     true,
		signCreated: true,
		fakeCreated: 0,
		expires:     0,
		nonce:       "",
	}
}

// SignAlg indicates that an "alg" signature parameters must be generated and signed (default: true).
func (c *SignConfig) SignAlg(b bool) *SignConfig {
	c.signAlg = b
	return c
}

// SignCreated indicates that a "created" signature parameters must be generated and signed (default: true).
func (c *SignConfig) SignCreated(b bool) *SignConfig {
	c.signCreated = b
	return c
}

// setFakeCreated indicates that the specified Unix timestamp must be used instead of the current time
// (default: 0, meaning use current time). Only used for testing.
func (c *SignConfig) setFakeCreated(ts int64) *SignConfig {
	c.fakeCreated = ts
	return c
}

// SetExpires adds an "expires" parameter containing an expiration deadline, as Unix time.
// Default: 0 (do not add the parameter).
func (c *SignConfig) SetExpires(expires int64) *SignConfig {
	c.expires = expires
	return c
}

// SetNonce adds a "nonce" string parameter whose content should be unique per signed message.
// Default: empty string (do not add the parameter).
func (c *SignConfig) SetNonce(nonce string) *SignConfig {
	c.nonce = nonce
	return c
}

// SetRequestResponse allows the server to indicate the signature name and signature that
// it had received in a client's request and include them in the signature input of the response.
func (c *SignConfig) SetRequestResponse(name, signature string) *SignConfig {
	c.requestResponse = &requestResponse{name, signature}
	return c
}

// VerifyConfig contains additional configuration for the verifier.
type VerifyConfig struct {
	verifyCreated   bool
	notNewerThan    time.Duration
	notOlderThan    time.Duration
	allowedAlgs     []string
	rejectExpired   bool
	requestResponse *requestResponse
	verifyKeyID     bool
	dateWithin      time.Duration
}

// SetNotNewerThan sets the window for messages that appear to be newer than the current time,
// which can only happen if clocks are out of sync. Default: 1,000 ms.
func (v *VerifyConfig) SetNotNewerThan(notNewerThan time.Duration) *VerifyConfig {
	v.notNewerThan = notNewerThan
	return v
}

// SetNotOlderThan sets the window for messages that are older than the current time,
// because of network latency. Default: 10,000 ms.
func (v *VerifyConfig) SetNotOlderThan(notOlderThan time.Duration) *VerifyConfig {
	v.notOlderThan = notOlderThan
	return v
}

// SetVerifyCreated indicates that the "created" parameter must be within some time window,
// defined by NotNewerThan and NotOlderThan. Default: true.
func (v *VerifyConfig) SetVerifyCreated(verifyCreated bool) *VerifyConfig {
	v.verifyCreated = verifyCreated
	return v
}

// SetRejectExpired indicates that expired messages (according to the "expires" parameter) must fail verification.
// Default: true.
func (v *VerifyConfig) SetRejectExpired(rejectExpired bool) *VerifyConfig {
	v.rejectExpired = rejectExpired
	return v
}

// SetAllowedAlgs defines the allowed values of the "alg" parameter.
// This is useful if the actual algorithm used in verification is taken from the message - not a recommended practice.
// Default: an empty list, signifying all values are accepted.
func (v *VerifyConfig) SetAllowedAlgs(allowedAlgs []string) *VerifyConfig {
	v.allowedAlgs = allowedAlgs
	return v
}

// SetRequestResponse allows the server to indicate signature name and signature that
// it had received from a client and include it in the signature input. Here this is configured
// on the client side when verifying the response.
func (v *VerifyConfig) SetRequestResponse(name, signature string) *VerifyConfig {
	v.requestResponse = &requestResponse{
		name:      name,
		signature: signature,
	}
	return v
}

// SetVerifyKeyID defines how to verify the keyid parameter, if one exists. If this value is set,
// the signature verifies only if the value is the same as was specified in the Verifier structure.
// Default: true.
func (v *VerifyConfig) SetVerifyKeyID(verify bool) *VerifyConfig {
	v.verifyKeyID = verify
	return v
}

// SetVerifyDateWithin indicates that the Date header should be verified if it exists, and its value
// must be within a certain time duration (positive or negative) of the Created signature parameter.
// This verification is only available if the Created field itself is verified.
// Default: 0, meaning no verification of the Date header.
func (v *VerifyConfig) SetVerifyDateWithin(d time.Duration) *VerifyConfig {
	v.dateWithin = d
	return v
}

// NewVerifyConfig generates a default configuration.
func NewVerifyConfig() *VerifyConfig {
	return &VerifyConfig{
		verifyCreated: true,
		notNewerThan:  2 * time.Second,
		notOlderThan:  10 * time.Second,
		rejectExpired: true,
		allowedAlgs:   []string{},
		verifyKeyID:   true,
		dateWithin:    0, // meaning no constraint
	}
}

// HandlerConfig contains additional configuration for the HTTP message handler wrapper.
// Either or both of fetchVerifier and fetchSigner may be nil for the corresponding operation
// to be skipped.
type HandlerConfig struct {
	reqNotVerified func(w http.ResponseWriter,
		r *http.Request, err error)
	fetchVerifier func(r *http.Request) (sigName string, verifier *Verifier)
	fetchSigner   func(res http.Response, r *http.Request) (sigName string, signer *Signer)
}

// NewHandlerConfig generates a default configuration. When verification or respectively,
// signing is required, the respective "fetch" callback must be supplied.
func NewHandlerConfig() *HandlerConfig {
	return &HandlerConfig{
		reqNotVerified: defaultReqNotVerified,
		fetchVerifier:  nil,
		fetchSigner:    nil,
	}
}

func defaultReqNotVerified(w http.ResponseWriter, _ *http.Request, err error) {
	w.WriteHeader(http.StatusUnauthorized)
	if err == nil { // should not happen
		_, _ = fmt.Fprintf(w, "Unknown error")
	} else {
		log.Println("Could not verify request signature: " + err.Error())
		_, _ = fmt.Fprintln(w, "Could not verify request signature") // For security reasons, do not print error
	}
}

// SetReqNotVerified defines a callback to be called when a request fails to verify. The default
// callback sends an unsigned 401 status code with a generic error message. For production, you
// probably need to sign it.
func (h *HandlerConfig) SetReqNotVerified(f func(w http.ResponseWriter, r *http.Request,
	err error)) *HandlerConfig {
	h.reqNotVerified = f
	return h
}

// SetFetchVerifier defines a callback that looks at the incoming request and provides
// a Verifier structure. In the simplest case, the signature name is a constant, and the key ID
// and key value are fetched based on the sender's identity, which in turn is gleaned
// from a header or query parameter. If a Verifier cannot be determined, the function should return Verifier as nil.
func (h *HandlerConfig) SetFetchVerifier(f func(r *http.Request) (sigName string, verifier *Verifier)) *HandlerConfig {
	h.fetchVerifier = f
	return h
}

// SetFetchSigner defines a callback that looks at the incoming request and the response, just before it is sent,
// and provides
// a Signer structure. In the simplest case, the signature name is a constant, and the key ID
// and key value are fetched based on the sender's identity. To simplify this logic,
// it is recommended to use the request's ctx (Context) member
// to store this information. If a Signer cannot be determined, the function should return Signer as nil.
func (h *HandlerConfig) SetFetchSigner(f func(res http.Response, r *http.Request) (sigName string, signer *Signer)) *HandlerConfig {
	h.fetchSigner = f
	return h
}
