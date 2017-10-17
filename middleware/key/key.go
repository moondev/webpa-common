package key

import "time"

// Interface is implemented by objects which represents keys that participate in the JWT workflow
type Interface interface {
	// KID returns the key identifier of this key
	KID() string

	// Key returns the key itself.  This can be a raw []byte or a parsed data type, like *rsa.PublicKey,
	// depending on the type of key.  Pass this value to other infrastructure that verifies or signs JWTs.
	Key() interface{}

	// Sign indicates whether this key can be used for signing.  It is possible for a key to be used for
	// both sign and verify (e.g. HMAC).
	Sign() bool

	// Verify indicates whether this key can be used for verification. It is possible for a key to be used
	// for both sign and verify (e.g. HMAC).
	Verify() bool

	// Type is the type of key, which describes both how the key should be parsed and which
	// algorithms it can be used for.
	Type() Type

	// Expires returns the system time at which this key should no longer be used.
	// If this method returns zero time, e.g. time.IsZero returns true, then this key
	// does not expire.
	Expires() time.Time
}

// key is the internal implementation of Interface.  Several convenient factory methods
// existing for creating instances of this type.
type key struct {
	kid          string
	parsedKey    interface{}
	keyType      Type
	sign, verify bool
	expires      time.Time
}

func (k *key) KID() string {
	return k.kid
}

func (k *key) Key() interface{} {
	return k.parsedKey
}

func (k *key) Sign() bool {
	return k.sign
}

func (k *key) Verify() bool {
	return k.verify
}

func (k *key) Type() Type {
	return k.keyType
}

func (k *key) Expires() time.Time {
	return k.expires
}

// NewVerifyKey returns a key useful for verifying signatures
func NewVerifyKey(kid, alg string, data []byte, expires time.Time) (Interface, error) {
	keyType, err := TypeFromAlg(alg)
	if err != nil {
		return nil, err
	}

	parsedKey, err := keyType.ParseVerifyKey(data)
	if err != nil {
		return nil, err
	}

	return &key{
		kid:       kid,
		parsedKey: parsedKey,
		keyType:   keyType,
		verify:    true,
		sign:      keyType.Both(),
		expires:   expires,
	}, nil
}

// NewSignKey returns a key useful for signing
func NewSignKey(kid, alg string, data []byte, expires time.Time) (Interface, error) {
	keyType, err := TypeFromAlg(alg)
	if err != nil {
		return nil, err
	}

	parsedKey, err := keyType.ParseSignKey(data)
	if err != nil {
		return nil, err
	}

	return &key{
		kid:       kid,
		parsedKey: parsedKey,
		keyType:   keyType,
		verify:    keyType.Both(),
		sign:      true,
		expires:   expires,
	}, nil
}

// RefreshKey produces a new key of the same type as an original, but with new data and expires
func RefreshKey(original Interface, data []byte, expires time.Time) (Interface, error) {
	var (
		parsedKey interface{}
		err       error
	)

	// key types that are used for both verify and sign are still parseable as verify keys
	if original.Verify() {
		if parsedKey, err = original.Type().ParseVerifyKey(data); err != nil {
			return nil, err
		}
	} else {
		if parsedKey, err = original.Type().ParseSignKey(data); err != nil {
			return nil, err
		}
	}

	return &key{
		kid:       original.KID(),
		parsedKey: parsedKey,
		keyType:   original.Type(),
		verify:    original.Verify(),
		sign:      original.Sign(),
		expires:   expires,
	}, nil
}
