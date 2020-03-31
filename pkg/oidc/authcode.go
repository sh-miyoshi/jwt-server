package oidc

import (
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sh-miyoshi/hekate/pkg/db"
	"github.com/sh-miyoshi/hekate/pkg/db/model"
	"github.com/sh-miyoshi/hekate/pkg/logger"
)

// NewAuthRequest ...
func NewAuthRequest(values url.Values) *AuthRequest {
	return &AuthRequest{
		Scope:        values.Get("scope"),
		ResponseType: values.Get("response_type"),
		ClientID:     values.Get("client_id"),
		RedirectURI:  values.Get("redirect_uri"),
		State:        values.Get("state"),
		Nonce:        values.Get("nonce"),
	}
}

// GenerateAuthCode ...
func GenerateAuthCode(clientID string, redirectURL string, userID string, nonce string) (string, error) {
	code := &model.AuthCode{
		CodeID:      uuid.New().String(),
		ClientID:    clientID,
		RedirectURL: redirectURL,
		ExpiresIn:   time.Now().Add(time.Second * time.Duration(expiresTimeSec)),
		UserID:      userID,
		Nonce:       nonce,
	}

	err := db.GetInst().AuthCodeAdd(code)

	return code.CodeID, err
}

func verifyAuthCode(codeID string) (*model.AuthCode, error) {
	code, err := db.GetInst().AuthCodeGet(codeID)
	if err != nil {
		if errors.Cause(err) == model.ErrNoSuchCode {
			// TODO(revoke all token in code.UserID) <- SHOULD
			return nil, errors.Wrap(ErrInvalidRequest, "no such code")
		}
		return nil, err
	}
	logger.Debug("Code: %v", code)

	if time.Now().Unix() >= code.ExpiresIn.Unix() {
		return nil, errors.Wrap(ErrInvalidRequest, "code is already expired")
	}

	return code, nil
}
