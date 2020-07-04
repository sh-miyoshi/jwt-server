package memory

import (
	"github.com/sh-miyoshi/hekate/pkg/db/model"
	"github.com/sh-miyoshi/hekate/pkg/errors"
)

// CustomRoleHandler implement db.CustomRoleHandler
type CustomRoleHandler struct {
	// roleList[roleID] = CustomRole
	roleList map[string]*model.CustomRole
}

// NewCustomRoleHandler ...
func NewCustomRoleHandler() *CustomRoleHandler {
	res := &CustomRoleHandler{
		roleList: make(map[string]*model.CustomRole),
	}
	return res
}

// Add ...
func (h *CustomRoleHandler) Add(projectName string, ent *model.CustomRole) *errors.Error {
	h.roleList[ent.ID] = ent
	return nil
}

// Delete ...
func (h *CustomRoleHandler) Delete(projectName string, roleID string) *errors.Error {
	if _, exists := h.roleList[roleID]; exists {
		if h.roleList[roleID].ProjectName == projectName {
			delete(h.roleList, roleID)
			return nil
		}
	}
	return model.ErrNoSuchCustomRole
}

// GetList ...
func (h *CustomRoleHandler) GetList(projectName string, filter *model.CustomRoleFilter) ([]*model.CustomRole, *errors.Error) {
	res := []*model.CustomRole{}

	for _, role := range h.roleList {
		if role.ProjectName == projectName {
			res = append(res, role)
		}
	}

	if filter != nil {
		res = filterRoleList(res, filter)
	}

	return res, nil
}

// Get ...
func (h *CustomRoleHandler) Get(projectName string, roleID string) (*model.CustomRole, *errors.Error) {
	res, exists := h.roleList[roleID]
	if !exists || res.ProjectName != projectName {
		return nil, model.ErrNoSuchCustomRole
	}

	return res, nil
}

// Update ...
func (h *CustomRoleHandler) Update(projectName string, ent *model.CustomRole) *errors.Error {
	if res, exists := h.roleList[ent.ID]; !exists || res.ProjectName != projectName {
		return model.ErrNoSuchCustomRole
	}

	h.roleList[ent.ID] = ent

	return nil
}

// DeleteAll ...
func (h *CustomRoleHandler) DeleteAll(projectName string) *errors.Error {
	for _, role := range h.roleList {
		if role.ProjectName == projectName {
			delete(h.roleList, role.ID)
		}
	}
	return nil
}

func filterRoleList(data []*model.CustomRole, filter *model.CustomRoleFilter) []*model.CustomRole {
	if filter == nil {
		return data
	}
	res := []*model.CustomRole{}

	for _, role := range data {
		if filter.Name != "" && role.Name != filter.Name {
			// missmatch name
			continue
		}
		// TODO(add other filter)
		res = append(res, role)
	}

	return res
}
