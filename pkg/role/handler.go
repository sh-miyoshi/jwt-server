package role

import (
	"fmt"

	"github.com/sh-miyoshi/hekate/pkg/errors"
	"github.com/sh-miyoshi/hekate/pkg/logger"
)

// Handler ...
type Handler struct {
	roleList []Info
}

var inst *Handler

// InitHandler ...
func InitHandler() *errors.Error {
	if inst != nil {
		return errors.New("Internal server error", "Default Role Handler is already initialized")
	}

	inst = &Handler{}

	// Create default role
	inst.createRole(ResCluster, TypeRead)
	inst.createRole(ResCluster, TypeWrite)
	inst.createRole(ResProject, TypeRead)
	inst.createRole(ResProject, TypeWrite)

	roles := []string{}
	for _, role := range inst.roleList {
		roles = append(roles, role.Name)
	}
	logger.Debug("All Default Role List: %v", roles)

	return nil
}

// Authorize ...
func Authorize(roles []string, targetResource Resource, roleType Type) bool {
	name := fmt.Sprintf("%s-%s", roleType.String(), targetResource.String())

	for _, role := range roles {
		if role == name {
			return true
		}
	}
	return false
}

// GetInst returns an instance of DB Manager
func GetInst() *Handler {
	return inst
}

// GetList ...
func (h *Handler) GetList() []string {
	res := []string{}
	for _, role := range h.roleList {
		res = append(res, role.ID)
	}
	return res
}

// Parse method parse role id string into resource and type
func (h *Handler) Parse(role string) (*Resource, *Type, bool) {
	for _, r := range h.roleList {
		if r.ID == role {
			return &r.TargetResource, &r.RoleType, true
		}
	}
	return nil, nil, false
}

func (h *Handler) createRole(targetResource Resource, roleType Type) {
	name := fmt.Sprintf("%s-%s", roleType.String(), targetResource.String())
	val := Info{
		ID:             name,
		Name:           name,
		TargetResource: targetResource,
		RoleType:       roleType,
	}
	h.roleList = append(h.roleList, val)
}
