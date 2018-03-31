// Package jwt implements JWT verification.
// See "JSON Web Token (JWT)" RFC 7519
// and "JSON Web Signature (JWS)" RFC 7515.
package jwt

import (
	"crypto"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

// Algorithm Identification Tokens
const (
	HS256 = "HS256" // HMAC SHA-256
	HS384 = "HS384" // HMAC SHA-384
	HS512 = "HS512" // HMAC SHA-512
	RS256 = "RS256" // RSASSA-PKCS1-v1_5 with SHA-256
	RS384 = "RS384" // RSASSA-PKCS1-v1_5 with SHA-348
	RS512 = "RS512" // RSASSA-PKCS1-v1_5 with SHA-512
)

// HMACAlgs is the HMAC hash algorithm registration.
// When adding additional entries you also need to
// import the respective packages to link the hash
// function into the binary [crypto.Hash.Available].
var HMACAlgs = map[string]crypto.Hash{
	HS256: crypto.SHA256,
	HS384: crypto.SHA384,
	HS512: crypto.SHA512,
}

// RSAAlgs is the RSA hash algorithm registration.
// When adding additional entries you also need to
// import the respective packages to link the hash
// function into the binary [crypto.Hash.Available].
var RSAAlgs = map[string]crypto.Hash{
	RS256: crypto.SHA256,
	RS384: crypto.SHA384,
	RS512: crypto.SHA512,
}

// See crypto.Hash.Available.
var errHashLink = errors.New("jwt: hash function not linked into binary")

var encoding = base64.RawURLEncoding

// Registered are the IANA registered "JSON Web Token Claims".
type Registered struct {
	// Issuer claim identifies the principal that issued the JWT.
	Issuer string `json:"iss,omitempty"`

	// Subject claim identifies the principal that is the subject of the JWT.
	Subject string `json:"sub,omitempty"`

	// Audience claim identifies the recipients that the JWT is intended for.
	Audience string `json:"aud,omitempty"`

	// Expires claim identifies the expiration time on or after which the JWT
	// must not be accepted for processing.
	Expires *NumericTime `json:"exp,omitempty"`

	// NotBefore claim identifies the time before which the JWT must not be
	// accepted for processing.
	NotBefore *NumericTime `json:"nbf,omitempty"`

	// Issued identifies the time at which the JWT was issued.
	Issued *NumericTime `json:"iat,omitempty"`

	// ID claim provides a unique identifier for the JWT.
	ID string `json:"jti,omitempty"`
}

// Claims is claims set payload representation.
type Claims struct {
	Registered

	// Raw has the JSON payload. This field is read-only.
	Raw json.RawMessage

	// Set is the claims set mapped by name.
	Set map[string]interface{}
}

// Valid returns whether the claims sets may be accepted
// for processing at the given moment in time.
func (c *Claims) Valid(t time.Time) bool {
	n := NewNumericTime(t)
	if n == nil {
		// if there are time limits then can't be sure.
		return c.Registered.NotBefore == nil && c.Registered.Expires == nil
	}
	if c.Registered.Expires != nil && *c.Registered.Expires <= *n {
		return false
	}
	if c.Registered.NotBefore != nil && *c.Registered.NotBefore > *n {
		return false
	}
	return true
}

// String returns the claim when present and if the representation is a JSON string.
func (c *Claims) String(name string) (value string, ok bool) {
	v, ok := c.Set[name]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	if !ok {
		return "", false
	}
	return s, true
}

// Number returns the claim when present and if the representation is a JSON number.
func (c *Claims) Number(name string) (value float64, ok bool) {
	v, ok := c.Set[name]
	if !ok {
		return 0, false
	}
	f, ok := v.(float64)
	if !ok {
		return 0, false
	}
	return f, true
}

// NumericTime is a JSON numeric value representing the number
// of seconds from 1970-01-01T00:00:00Z UTC until the specified
// UTC date/time, ignoring leap seconds.
type NumericTime float64

// NewNumericTime returns the the corresponding
// representation with nil for the zero value.
func NewNumericTime(t time.Time) *NumericTime {
	if t.IsZero() {
		return nil
	}
	n := NumericTime(float64(t.UnixNano()) / 1E9)
	return &n
}

// Time returns the Go mapping with the zero value for nil.
func (n *NumericTime) Time() time.Time {
	if n == nil {
		return time.Time{}
	}
	return time.Unix(0, int64(float64(*n)*float64(time.Second))).UTC()
}

// String returs the ISO representation.
func (n *NumericTime) String() string {
	if n == nil {
		return ""
	}
	return n.Time().Format("2006-01-02T15:04:05.999999999Z")
}
