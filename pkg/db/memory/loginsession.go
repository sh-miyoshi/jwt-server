package memory

import (
	"github.com/sh-miyoshi/hekate/pkg/db/model"
)

// LoginSessionHandler implement db.LoginSessionHandler
type LoginSessionHandler struct {
	// sessionList[verifyCode] = LoginSessionInfo
	sessionList map[string]*model.LoginSessionInfo
}

// NewLoginSessionHandler ...
func NewLoginSessionHandler() *LoginSessionHandler {
	res := &LoginSessionHandler{
		sessionList: make(map[string]*model.LoginSessionInfo),
	}
	return res
}

// Add ...
func (h *LoginSessionHandler) Add(info *model.LoginSessionInfo) error {
	if _, exists := h.sessionList[info.VerifyCode]; exists {
		return model.ErrLoginSessionAlreadyExists
	}

	h.sessionList[info.VerifyCode] = info
	return nil
}

// Delete ...
func (h *LoginSessionHandler) Delete(verifyCode string) error {
	if _, exists := h.sessionList[verifyCode]; !exists {
		return model.ErrNoSuchLoginSession
	}

	delete(h.sessionList, verifyCode)
	return nil
}

// Get ...
func (h *LoginSessionHandler) Get(verifyCode string) (*model.LoginSessionInfo, error) {
	if _, exists := h.sessionList[verifyCode]; !exists {
		return nil, model.ErrNoSuchLoginSession
	}

	return h.sessionList[verifyCode], nil
}
