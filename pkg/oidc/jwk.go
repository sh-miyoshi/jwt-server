package oidc

import (
	"crypto/x509"

	"github.com/dvsekhvalnov/jose2go/base64url"
	"github.com/google/uuid"
	"github.com/sh-miyoshi/hekate/pkg/errors"
	"github.com/sh-miyoshi/hekate/pkg/util"
)

// GenerateJWKSet ...
func GenerateJWKSet(signAlg string, publicKey []byte) (*JWKSet, *errors.Error) {
	jwk := JWKInfo{
		KeyID:        uuid.New().String(),
		Algorithm:    signAlg,
		PublicKeyUse: "sig",
	}

	switch signAlg {
	case "RS256":
		jwk.KeyType = "RSA"
		key, err := x509.ParsePKCS1PublicKey(publicKey)
		if err != nil {
			return nil, errors.New("RSA key parse failed", "Failed to parse RSA public key: %v", err)
		}
		e := util.Int2bytes(uint64(key.E))
		jwk.E = base64url.Encode(e)
		jwk.N = base64url.Encode(key.N.Bytes())
	default:
		return nil, errors.New("Invalid request", "Now such signing algorithm")
	}

	res := &JWKSet{}
	res.Keys = append(res.Keys, jwk)

	return res, nil
}
