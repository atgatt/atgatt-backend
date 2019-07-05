package helpers

import (
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/lestrrat-go/jwx/jwk"
)

// GetAuth0PublicKey gets the appropriate RSA public key used to sign JWTs when supplied with an auth0 domain name. This is needed to verify JWTs that are signed with the RS256 signing method.
func GetAuth0PublicKey(auth0Domain string) (*rsa.PublicKey, error) {
	var publicKey *rsa.PublicKey
	if auth0Domain == "" {
		return publicKey, errors.New("invalid domain")
	}

	jwks, err := jwk.FetchHTTP(fmt.Sprintf("https://%s/.well-known/jwks.json", auth0Domain))
	if err != nil {
		return nil, err
	}

	var rawPublicKey interface{}
	for _, k := range jwks.Keys {
		if k.KeyType().String() == "RSA" && k.KeyUsage() == "sig" {
			rawPublicKey, err = k.Materialize()
			if err != nil {
				return nil, err
			}
		}
	}

	if rawPublicKey == nil {
		err := errors.New("unable to find appropriate key")
		return publicKey, err
	}

	return rawPublicKey.(*rsa.PublicKey), nil
}
