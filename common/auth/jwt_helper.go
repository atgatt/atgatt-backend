package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type jwks struct {
	Keys []jsonWebKeys `json:"keys"`
}

type jsonWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

// GetAuth0PublicKey gets the appropriate public key from a given auth0 domain when supplied a JWT token
func GetAuth0PublicKey(auth0Domain string) (string, error) {
	if auth0Domain == "" {
		return "", errors.New("invalid domain")
	}
	cert := ""
	resp, err := http.Get(fmt.Sprintf("https://%s/.well-known/jwks.json", auth0Domain))
	if err != nil {
		return cert, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return cert, errors.New("got an unexpected http status code from auth0 when fetching jwks.json")
	}
	defer resp.Body.Close()

	var jwks = jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	for k := range jwks.Keys {
		// Find the RSA signing key
		if jwks.Keys[k].Kty == "RSA" && jwks.Keys[k].Use == "sig" {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
			break
		}
	}

	if cert == "" {
		err := errors.New("unable to find appropriate key")
		return cert, err
	}

	return cert, nil
}
