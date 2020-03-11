package memory

import (
	"github.com/sh-miyoshi/hekate/pkg/db/model"
)

// SessionHandler implement db.SessionHandler
type SessionHandler struct {
	// sessionList[sessionID] = Session
	sessionList map[string]*model.Session
}

// NewSessionHandler ...
func NewSessionHandler() *SessionHandler {
	res := &SessionHandler{
		sessionList: make(map[string]*model.Session),
	}
	return res
}

// New ...
func (h *SessionHandler) New(session *model.Session) error {
	h.sessionList[session.SessionID] = session
	return nil
}

// Revoke ...
func (h *SessionHandler) Revoke(sessionID string) error {
	if _, exists := h.sessionList[sessionID]; exists {
		delete(h.sessionList, sessionID)
		return nil
	}

	return model.ErrNoSuchSession
}

// RevokeAll ...
func (h *SessionHandler) RevokeAll(userID string) error {
	newList := make(map[string]*model.Session)
	for _, s := range h.sessionList {
		if s.UserID != userID {
			newList[s.SessionID] = s
		}
	}
	h.sessionList = newList
	return nil
}

// Get ...
func (h *SessionHandler) Get(sessionID string) (*model.Session, error) {
	if _, exists := h.sessionList[sessionID]; !exists {
		return nil, model.ErrNoSuchSession
	}
	return h.sessionList[sessionID], nil
}

// GetList ...
func (h *SessionHandler) GetList(userID string) ([]string, error) {
	res := []string{}

	for id, s := range h.sessionList {
		if s.UserID == userID {
			res = append(res, id)
		}
	}

	return res, nil
}
