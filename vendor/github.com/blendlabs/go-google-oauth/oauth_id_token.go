package google

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	exception "github.com/blendlabs/go-exception"
)

// DeserializeJWT deserializes a jwt token.
func DeserializeJWT(corpus string) (*JWT, error) {
	parts := strings.Split(corpus, ".")
	if len(parts) < 3 {
		return nil, ErrInvalidJWT
	}

	headerContents, err := decodeJWTSegment(parts[0])
	if err != nil {
		return nil, exception.Wrap(err)
	}
	var header JWTHeader
	if err = json.Unmarshal(headerContents, &header); err != nil {
		return nil, exception.Wrap(err)
	}

	payloadContents, err := decodeJWTSegment(parts[1])
	if err != nil {
		return nil, exception.Wrap(err)
	}
	var payload JWTPayload
	if err = json.Unmarshal(payloadContents, &payload); err != nil {
		return nil, exception.Wrap(err)
	}

	signature, err := decodeJWTSegment(parts[2])
	if err != nil {
		return nil, exception.Wrap(err)
	}

	return &JWT{
		Header:    header,
		Payload:   payload,
		Signature: signature,
	}, nil
}

func decodeJWTSegment(corpus string) ([]byte, error) {
	if l := len(corpus) % 4; l > 0 {
		corpus += strings.Repeat("=", 4-l)
	}
	contents, err := base64.URLEncoding.DecodeString(corpus)
	if err != nil {
		return nil, exception.Wrap(err)
	}
	return contents, nil
}

// JWT is a full jwt.
type JWT struct {
	Header    JWTHeader
	Payload   JWTPayload
	Signature []byte
}

// JWTHeader is the header of a jwt.
type JWTHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

// JWTPayload is the oauth JWT token.
type JWTPayload struct {
	ISS           string  `json:"iss"`
	ATHash        string  `json:"at_hash"`
	EmailVerified bool    `json:"email_verified"`
	Sub           string  `json:"sub"` //actual user identifier
	AZP           string  `json:"azp"`
	Email         string  `json:"email"`
	AUD           string  `json:"aud"`
	IAT           float64 `json:"iat"`
	EXP           float64 `json:"exp"`
	Nonce         string  `json:"nonce"`
	HostedDomain  string  `json:"hd"`
}
