package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

// NOTE: JUST ABOUT EVERYTHING IN THIS FILE WAS RIPPED FROM
// https://github.com/googleapis/google-api-go-client/tree/master/idtoken
// Long story short, I didn't know how to validate the jwt token google
// provided, SOOO! I asked for help as that looked to be the module I needed
// to use. Turns out what I need is not something they offer, but I was able to
// make a feature request to get that done. Hopefully, if they actually do the thing,
// I can nuke this whole file and use the module. I feel gross, having ripped all this
// code, but I guess it's the same as pulling in an import.
//
// AAAANNNYYYY WAY, that's why I'm not going to add tests for this file

const es256KeySize int = 32

// Information used to get data for validation
const (
	Issuer        = "chat@system.gserviceaccount.com"
	CertURLPrefix = "https://www.googleapis.com/service_accounts/v1/metadata/jwk/"
)

// JWT contains the encoded segments of the consumed
// JWT
type JWT struct {
	header    string
	payload   string
	signature string
}

// Header represents the decoded header object passed within
// the jwt
type Header struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	KeyID     string `json:"kid"`
}

// Payload represents the decoded payload object passed within
// the jwt
type Payload struct {
	Issuer   string                 `json:"iss"`
	Audience string                 `json:"aud"`
	Expires  int64                  `json:"exp"`
	IssuedAt int64                  `json:"iat"`
	Subject  string                 `json:"sub,omitempty"`
	Claims   map[string]interface{} `json:"-"`
}

// JWK represents the consumed jwk object
type JWK struct {
	Alg string `json:"alg"`
	Crv string `json:"crv"`
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	E   string `json:"e"`
	N   string `json:"n"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

type certResponse struct {
	Keys []JWK `json:"keys"`
}

// Validator is used to communicate with the given url
// to retrieve the cert data.
type Validator struct {
	client *http.Client
}

func validateJWT(authString string) (bool, error) {
	segments := strings.Split(authString, ".")

	if len(segments) != 3 {
		return false, errors.New("invalid JWT, expected 3 segments")
	}

	jwt := JWT{
		header:    segments[0],
		payload:   segments[1],
		signature: segments[2],
	}

	v := Validator{client: http.DefaultClient}

	header := jwt.parseHeader()
	sig := jwt.parseSignature()

	switch header.Algorithm {
	case "RS256":
		if err := v.validateRS256(header.KeyID, jwt.hashedContent(), []byte(sig)); err != nil {
			return false, err
		}
	case "ES256":
		if err := v.validateES256(header.KeyID, jwt.hashedContent(), []byte(sig)); err != nil {
			return false, err
		}
	default:
		return false, fmt.Errorf("idtoken: expected JWT signed with RS256 or ES256 but found %q", header.Algorithm)
	}

	return true, nil
}

func (j JWT) parseHeader() Header {
	var header Header

	h, err := jwt.DecodeSegment(j.header)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(h, &header); err != nil {
		panic(err)
	}

	return header
}

func (j JWT) parsePayload() Payload {
	var payload Payload

	p, err := jwt.DecodeSegment(j.payload)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(p, &payload); err != nil {
		panic(err)
	}

	return payload
}

func (j JWT) parseSignature() string {
	signature, err := jwt.DecodeSegment(j.signature)
	if err != nil {
		panic(err)
	}

	return string(signature)
}

func (j JWT) hashedContent() []byte {
	signedContent := j.header + "." + j.payload
	hashed := sha256.Sum256([]byte(signedContent))
	return hashed[:]
}

func (v *Validator) validateRS256(keyID string, hashedContent []byte, sig []byte) error {
	certResp, err := v.getCert(CertURLPrefix + Issuer)
	if err != nil {
		return err
	}

	j, err := findMatchingKey(certResp, keyID)
	if err != nil {
		return err
	}
	dn, err := decode(j.N)
	if err != nil {
		return err
	}
	de, err := decode(j.E)
	if err != nil {
		return err
	}

	pk := &rsa.PublicKey{
		N: new(big.Int).SetBytes(dn),
		E: int(new(big.Int).SetBytes(de).Int64()),
	}
	return rsa.VerifyPKCS1v15(pk, crypto.SHA256, hashedContent, sig)
}

func (v *Validator) validateES256(keyID string, hashedContent []byte, sig []byte) error {
	certResp, err := v.getCert(CertURLPrefix + Issuer)
	if err != nil {
		return err
	}
	j, err := findMatchingKey(certResp, keyID)
	if err != nil {
		return err
	}
	dx, err := decode(j.X)
	if err != nil {
		return err
	}
	dy, err := decode(j.Y)
	if err != nil {
		return err
	}

	pk := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(dx),
		Y:     new(big.Int).SetBytes(dy),
	}
	r := big.NewInt(0).SetBytes(sig[:es256KeySize])
	s := big.NewInt(0).SetBytes(sig[es256KeySize:])
	if valid := ecdsa.Verify(pk, hashedContent, r, s); !valid {
		return fmt.Errorf("idtoken: ES256 signature not valid")
	}
	return nil
}

func findMatchingKey(response *certResponse, keyID string) (*JWK, error) {
	if response == nil {
		return nil, fmt.Errorf("idtoken: cert response is nil")
	}
	for _, v := range response.Keys {
		if v.Kid == keyID {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("idtoken: could not find matching cert keyId for the token provided")
}

func (v Validator) getCert(url string) (*certResponse, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("idtoken: unable to retrieve cert, got status code %d", resp.StatusCode)
	}

	certResp := &certResponse{}
	if err := json.NewDecoder(resp.Body).Decode(certResp); err != nil {
		return nil, err

	}

	return certResp, nil
}

func decode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
