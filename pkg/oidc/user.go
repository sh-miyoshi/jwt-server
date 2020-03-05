package oidc

import (
	"github.com/sh-miyoshi/hekate/pkg/db"
	"github.com/sh-miyoshi/hekate/pkg/db/model"
	"html/template"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sh-miyoshi/hekate/pkg/logger"
)

// WriteUserLoginPage ...
func WriteUserLoginPage(code string, state string, projectName string, w http.ResponseWriter) {
	tpl, err := template.ParseFiles(userLoginHTML)
	if err != nil {
		logger.Error("Failed to parse template: %v", err)
		http.Error(w, "User Login Page maybe broken", http.StatusInternalServerError)
		return
	}

	url := "/api/v1/project/" + projectName + "/openid-connect/login?login_verify_code=" + code
	if state != "" {
		url += "&state=" + state
	}

	d := map[string]string{
		"URL":             url,
		"CSSResourcePath": userLoginResPath + "/css",
		"IMGResourcePath": userLoginResPath + "/img",
	}

	w.Header().Add("Content-Type", "text/html; charset=UTF-8")
	tpl.Execute(w, d)
}

// RegisterUserLoginSession ...
func RegisterUserLoginSession(req *AuthRequest) (string, error) {
	code := uuid.New().String()

	s := &model.LoginSessionInfo{
		VerifyCode:  code,
		ExpiresIn:   time.Now().Add(time.Second * time.Duration(expiresTimeSec)),
		ClientID:    req.ClientID,
		RedirectURI: req.RedirectURI,
	}

	if err := db.GetInst().LoginSessionAdd(s); err != nil {
		return "", errors.Wrap(err, "add user login session failed")
	}
	return code, nil
}

// UserLoginVerify ...
func UserLoginVerify(code string) (*UserLoginInfo, error) {
	s, err := db.GetInst().LoginSessionGet(code)
	if err != nil {
		return nil, errors.Wrap(err, "user login session get failed")
	}

	if err := db.GetInst().LoginSessionDelete(code); err != nil {
		return nil, errors.Wrap(err, "user login sessiond delete failed")
	}

	now := time.Now().Unix()
	if now > s.ExpiresIn.Unix() {
		return nil, errors.New("Session already expired")
	}
	return &UserLoginInfo{
		ClientID:    s.ClientID,
		RedirectURI: s.RedirectURI,
	}, nil
}
