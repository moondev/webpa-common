package key

import (
	"fmt"
	"math"

	jwt "github.com/dgrijalva/jwt-go"
)

//go:generate stringer -type=Type

// Type is the type of key, which determines the JWT algorithms
// that can use the key.
type Type int

const (
	HMAC Type = iota
	RSA
	EC
)

// Both returns true if keys of this type can be used for both sign and verify.
func (t Type) Both() bool {
	return t == HMAC
}

// ParseVerifyKey parses the given raw key data based on the Type.  For HMAC,
// the raw data is used as is as the key.  For other key types, the data is
// assumed to be a PEM-encoded public key.
func (t Type) ParseVerifyKey(data []byte) (interface{}, error) {
	switch t {
	case HMAC:
		return data, nil
	case RSA:
		return jwt.ParseRSAPublicKeyFromPEM(data)
	case EC:
		return jwt.ParseECPublicKeyFromPEM(data)
	default:
		return nil, fmt.Errorf("Invalid key type: %d", t)
	}
}

// ParseSignKey parses the given raw key data based on the Type.  For HMAC,
// the raw data is used as is as the key.  For other key types, the data is
// assumed to be a PEM-encoded private key.
func (t Type) ParseSignKey(data []byte) (interface{}, error) {
	switch t {
	case HMAC:
		return data, nil
	case RSA:
		return jwt.ParseRSAPrivateKeyFromPEM(data)
	case EC:
		return jwt.ParseECPrivateKeyFromPEM(data)
	default:
		return nil, fmt.Errorf("Invalid key type: %d", t)
	}
}

// TypeFromAlg determines the key Type from the JWT alg value
func TypeFromAlg(alg string) (Type, error) {
	switch alg[0:2] {
	case "HS":
		return HMAC, nil

	case "RS":
		return RSA, nil

	case "PS":
		return RSA, nil

	case "ES":
		return EC, nil

	default:
		return Type(math.MaxUint32), fmt.Errorf("Unsupported algorithm: %s", alg)
	}
}
