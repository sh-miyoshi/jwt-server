package memory

import (
	"github.com/sh-miyoshi/jwt-server/pkg/db/model"
	"sync"
)

// CustomRoleHandler implement db.CustomRoleHandler
type CustomRoleHandler struct {
	// roleList[roleID] = CustomRole
	roleList map[string]*model.CustomRole
	mu       sync.Mutex
}

// NewCustomRoleHandler ...
func NewCustomRoleHandler() (*CustomRoleHandler, error) {
	res := &CustomRoleHandler{
		roleList: make(map[string]*model.CustomRole),
	}
	return res, nil
}

// Add ...
func (h *CustomRoleHandler) Add(ent *model.CustomRole) error {
	h.roleList[ent.ID] = ent
	return nil
}

// Delete ...
func (h *CustomRoleHandler) Delete(roleID string) error {
	if _, exists := h.roleList[roleID]; exists {
		delete(h.roleList, roleID)
		return nil
	}
	return model.ErrNoSuchCustomRole
}

// GetList ...
func (h *CustomRoleHandler) GetList(projectName string) ([]string, error) {
	res := []string{}

	for _, role := range h.roleList {
		if role.ProjectName == projectName {
			res = append(res, role.ID)
		}
	}

	return res, nil
}

// Get ...
func (h *CustomRoleHandler) Get(roleID string) (*model.CustomRole, error) {
	res, exists := h.roleList[roleID]
	if !exists {
		return nil, model.ErrNoSuchCustomRole
	}

	return res, nil
}

// Update ...
func (h *CustomRoleHandler) Update(ent *model.CustomRole) error {
	if _, exists := h.roleList[ent.ID]; !exists {
		return model.ErrNoSuchCustomRole
	}

	h.roleList[ent.ID] = ent

	return nil
}

// DeleteAll ...
func (h *CustomRoleHandler) DeleteAll(projectName string) error {
	for _, role := range h.roleList {
		if role.ProjectName == projectName {
			delete(h.roleList, role.ID)
		}
	}
	return nil
}

// BeginTx ...
func (h *CustomRoleHandler) BeginTx() error {
	h.mu.Lock()
	return nil
}

// CommitTx ...
func (h *CustomRoleHandler) CommitTx() error {
	h.mu.Unlock()
	return nil
}

// AbortTx ...
func (h *CustomRoleHandler) AbortTx() error {
	h.mu.Unlock()
	return nil
}