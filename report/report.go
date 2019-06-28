package report

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jws"
)

type oauthTokenResponse struct {
	TokenType        string `json:"token_type"`
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// AccessToken swaps the given JWT for an access token.
func AccessToken(jwt string) (*oauth2.Token, error) {
	v := make(url.Values)
	v.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	v.Set("assertion", jwt)

	resp, err := http.PostForm("https://accounts.google.com/o/oauth2/token", v)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out oauthTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return nil, err
	}

	if out.Error != "" {
		return nil, fmt.Errorf("oauth error %v: %v", out.Error, out.ErrorDescription)
	}

	return &oauth2.Token{
		AccessToken: out.AccessToken,
		TokenType:   out.TokenType,
		Expiry:      time.Now().Add(time.Second * time.Duration(out.ExpiresIn)),
	}, nil
}

// GenerateJWT creates a signed JSON Web Token using a Google API Service Account.
func GenerateJWT(raw []byte, expiry time.Duration) (string, error) {
	audience := "https://accounts.google.com/o/oauth2/token"
	scope := "https://www.googleapis.com/auth/homegraph"

	// Extract the RSA private key from the service account keyfile bytes.
	conf, err := google.JWTConfigFromJSON(raw)
	if err != nil {
		return "", fmt.Errorf("Could not parse service account JSON: %v", err)
	}

	now := time.Now().Unix()
	expiryLength := int64(expiry.Seconds())

	jwt := &jws.ClaimSet{
		Scope: scope,
		Iat:   now,
		Exp:   now + expiryLength,
		Aud:   audience,
		Iss:   conf.Email,
	}
	jwsHeader := &jws.Header{
		Algorithm: "RS256",
		Typ:       "JWT",
	}

	block, _ := pem.Decode(conf.PrivateKey)
	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("private key parse error: %v", err)
	}
	rsaKey, ok := parsedKey.(*rsa.PrivateKey)

	// Sign the JWT with the service account's private key.
	if !ok {
		return "", errors.New("private key failed rsa.PrivateKey type assertion")
	}
	return jws.Encode(jwsHeader, jwt, rsaKey)
}
