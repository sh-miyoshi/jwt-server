package userapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sh-miyoshi/hekate/pkg/audit"
	"github.com/sh-miyoshi/hekate/pkg/db"
	"github.com/sh-miyoshi/hekate/pkg/db/model"
	"github.com/sh-miyoshi/hekate/pkg/errors"
	jwthttp "github.com/sh-miyoshi/hekate/pkg/http"
	"github.com/sh-miyoshi/hekate/pkg/logger"
	"github.com/sh-miyoshi/hekate/pkg/role"
	"github.com/sh-miyoshi/hekate/pkg/secret"
	"github.com/sh-miyoshi/hekate/pkg/util"
)

// AllUserGetHandler ...
//   require role: read-project
func AllUserGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]

	// Authorize API Request
	if err := jwthttp.Authorize(r, projectName, role.ResProject, role.TypeRead); err != nil {
		errors.PrintAsInfo(errors.Append(err, "Failed to authorize header"))
		errors.WriteToHTTP(w, errors.ErrUnpermitted, 0, "")
		return
	}

	queries := r.URL.Query()
	logger.Debug("Query: %v", queries)

	filter := &model.UserFilter{
		Name: queries.Get("name"),
	}

	users, err := db.GetInst().UserGetList(projectName, filter)
	if err != nil {
		if errors.Contains(err, model.ErrUserValidateFailed) {
			errors.PrintAsInfo(errors.Append(err, "User Get List Failed"))
			errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
		} else {
			errors.Print(errors.Append(err, "Failed to get user"))
			errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		}
		return
	}

	// Get all custom roles due to check all users
	customRoles, err := db.GetInst().CustomRoleGetList(projectName, nil)
	if err != nil {
		errors.Print(errors.Append(err, "Failed to get custom role list"))
		errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		return
	}

	res := []*UserGetResponse{}
	for _, user := range users {
		roles := []CustomRole{}
		for _, rid := range user.CustomRoles {
			for _, r := range customRoles {
				if rid == r.ID {
					roles = append(roles, CustomRole{
						r.ID,
						r.Name,
					})
					break
				}
			}
		}

		tmp := &UserGetResponse{
			ID:          user.ID,
			Name:        user.Name,
			CreatedAt:   user.CreatedAt.Format(time.RFC3339),
			SystemRoles: user.SystemRoles,
			CustomRoles: roles,
			Locked:      user.LockState.Locked,
		}
		sessions, err := db.GetInst().SessionGetList(projectName, &model.SessionFilter{UserID: user.ID})
		if err != nil {
			errors.Print(errors.Append(err, "Failed to get session list"))
			errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
			return
		}

		for _, s := range sessions {
			tmp.Sessions = append(tmp.Sessions, s.SessionID)
		}

		res = append(res, tmp)
	}

	jwthttp.ResponseWrite(w, "UserGetAllUserGetHandlerHandler", &res)
}

// UserCreateHandler ...
//   require role: write-project
func UserCreateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]

	var err *errors.Error
	defer func() {
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		if err = audit.GetInst().Save(projectName, time.Now(), "USER", r.Method, r.URL.String(), msg); err != nil {
			errors.Print(errors.Append(err, "Failed to save audit event"))
		}
	}()

	// Authorize API Request
	if err = jwthttp.Authorize(r, projectName, role.ResProject, role.TypeWrite); err != nil {
		errors.PrintAsInfo(errors.Append(err, "Failed to authorize header"))
		errors.WriteToHTTP(w, errors.ErrUnpermitted, 0, "")
		return
	}

	// Parse Request
	var request UserCreateRequest
	if e := json.NewDecoder(r.Body).Decode(&request); e != nil {
		err = errors.New("Invalid request", "Failed to decode user create request: %v", e)
		errors.PrintAsInfo(err)
		errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
		return
	}

	// validate password
	project, err := db.GetInst().ProjectGet(projectName)
	if err != nil {
		errors.Print(errors.Append(err, "Failed to get project"))
		errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		return
	}

	if err = secret.CheckPassword(request.Name, request.Password, project.PasswordPolicy); err != nil {
		errors.PrintAsInfo(errors.Append(err, "The password does not much the policy"))
		errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
		return
	}

	// Create User Entry
	user := model.UserInfo{
		ID:           uuid.New().String(),
		ProjectName:  projectName,
		Name:         request.Name,
		CreatedAt:    time.Now(),
		PasswordHash: util.CreateHash(request.Password),
		SystemRoles:  request.SystemRoles,
		CustomRoles:  request.CustomRoles,
	}

	if err = db.GetInst().UserAdd(projectName, &user); err != nil {
		if errors.Contains(err, model.ErrUserValidateFailed) {
			errors.PrintAsInfo(errors.Append(err, "user validation failed"))
			errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
		} else if errors.Contains(err, model.ErrUserAlreadyExists) {
			errors.PrintAsInfo(errors.Append(err, "User %s is already exists", user.Name))
			errors.WriteToHTTP(w, err, http.StatusConflict, "")
		} else {
			errors.Print(errors.Append(err, "Failed to create user"))
			errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		}
		return
	}

	roles := []CustomRole{}
	for _, rid := range user.CustomRoles {
		r, err := db.GetInst().CustomRoleGet(projectName, rid)
		if err != nil {
			errors.Print(errors.Append(err, "Failed to get user %s custom role %s info", user.ID, r.ID))
			errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
			return
		}
		roles = append(roles, CustomRole{
			ID:   r.ID,
			Name: r.Name,
		})
	}

	// Return Response
	res := UserGetResponse{
		ID:          user.ID,
		Name:        user.Name,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		SystemRoles: user.SystemRoles,
		CustomRoles: roles,
		Locked:      user.LockState.Locked,
	}

	jwthttp.ResponseWrite(w, "UserGetAllUserGetHandlerHandler", &res)
}

// UserDeleteHandler ...
//   require role: write-project
func UserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	userID := vars["userID"]

	var err *errors.Error
	defer func() {
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		if err = audit.GetInst().Save(projectName, time.Now(), "USER", r.Method, r.URL.String(), msg); err != nil {
			errors.Print(errors.Append(err, "Failed to save audit event"))
		}
	}()

	// Authorize API Request
	if err = jwthttp.Authorize(r, projectName, role.ResProject, role.TypeWrite); err != nil {
		errors.PrintAsInfo(errors.Append(err, "Failed to authorize header"))
		errors.WriteToHTTP(w, errors.ErrUnpermitted, 0, "")
		return
	}

	// Delete User
	if err = db.GetInst().UserDelete(projectName, userID); err != nil {
		if errors.Contains(err, model.ErrNoSuchUser) || errors.Contains(err, model.ErrUserValidateFailed) {
			errors.PrintAsInfo(errors.Append(err, "User %s is not found", userID))
			errors.WriteToHTTP(w, err, http.StatusNotFound, "")
		} else {
			errors.Print(errors.Append(err, "Failed to delete user"))
			errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		}
		return
	}

	// Return 204 (No content) for success
	w.WriteHeader(http.StatusNoContent)
	logger.Info("UserDeleteHandler method successfully finished")
}

// UserGetHandler ...
//   require role: read-project
func UserGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	userID := vars["userID"]

	// Authorize API Request
	if err := jwthttp.Authorize(r, projectName, role.ResProject, role.TypeRead); err != nil {
		errors.PrintAsInfo(errors.Append(err, "Failed to authorize header"))
		errors.WriteToHTTP(w, errors.ErrUnpermitted, 0, "")
		return
	}

	user, err := db.GetInst().UserGet(projectName, userID)
	if err != nil {
		if errors.Contains(err, model.ErrNoSuchUser) || errors.Contains(err, model.ErrUserValidateFailed) {
			errors.PrintAsInfo(errors.Append(err, "User %s is not found", userID))
			errors.WriteToHTTP(w, err, http.StatusNotFound, "")
		} else {
			errors.Print(errors.Append(err, "Failed to get user"))
			errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		}
		return
	}

	roles := []CustomRole{}
	for _, rid := range user.CustomRoles {
		r, err := db.GetInst().CustomRoleGet(projectName, rid)
		if err != nil {
			errors.Print(errors.Append(err, "Failed to get user %s custom role %s info", user.ID, r.ID))
			errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
			return
		}
		roles = append(roles, CustomRole{
			ID:   r.ID,
			Name: r.Name,
		})
	}

	res := UserGetResponse{
		ID:          user.ID,
		Name:        user.Name,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		SystemRoles: user.SystemRoles,
		CustomRoles: roles,
		Locked:      user.LockState.Locked,
	}

	sessions, err := db.GetInst().SessionGetList(projectName, &model.SessionFilter{UserID: user.ID})
	if err != nil {
		errors.Print(errors.Append(err, "Failed to get session list"))
		errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		return
	}

	for _, s := range sessions {
		res.Sessions = append(res.Sessions, s.SessionID)
	}

	jwthttp.ResponseWrite(w, "UserGetHandler", &res)
}

// UserUpdateHandler ...
//   require role: write-project
func UserUpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	userID := vars["userID"]

	var err *errors.Error
	defer func() {
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		if err = audit.GetInst().Save(projectName, time.Now(), "USER", r.Method, r.URL.String(), msg); err != nil {
			errors.Print(errors.Append(err, "Failed to save audit event"))
		}
	}()

	// Authorize API Request
	if err = jwthttp.Authorize(r, projectName, role.ResProject, role.TypeWrite); err != nil {
		errors.PrintAsInfo(errors.Append(err, "Failed to authorize header"))
		errors.WriteToHTTP(w, errors.ErrUnpermitted, 0, "")
		return
	}

	// Parse Request
	var request UserPutRequest
	if e := json.NewDecoder(r.Body).Decode(&request); e != nil {
		err = errors.Append(errors.ErrInvalidRequest, "Failed to decode user update request: %v", e)
		errors.PrintAsInfo(err)
		errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
		return
	}

	// Get Previous User Info
	user, err := db.GetInst().UserGet(projectName, userID)
	if err != nil {
		if errors.Contains(err, model.ErrNoSuchUser) {
			errors.PrintAsInfo(errors.Append(err, "User %s is not found", userID))
			errors.WriteToHTTP(w, err, http.StatusNotFound, "")
		} else if errors.Contains(err, model.ErrUserValidateFailed) {
			errors.PrintAsInfo(errors.Append(err, "Invalid user ID format"))
			errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
		} else {
			errors.Print(errors.Append(err, "Failed to update user"))
			errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		}
		return
	}

	// Update Parameters
	// name, roles
	user.Name = request.Name
	user.SystemRoles = request.SystemRoles
	user.CustomRoles = request.CustomRoles

	// Update DB
	if err = db.GetInst().UserUpdate(projectName, user); err != nil {
		if errors.Contains(err, model.ErrUserValidateFailed) || errors.Contains(err, model.ErrUserAlreadyExists) {
			errors.PrintAsInfo(errors.Append(err, "Invalid user request format"))
			errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
		} else {
			errors.Print(errors.Append(err, "Failed to update user"))
			errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.Info("UserUpdateHandler method successfully finished")
}

// UserRoleAddHandler ...
//   require role: write-project
func UserRoleAddHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	userID := vars["userID"]
	roleID := vars["roleID"]

	var err *errors.Error
	defer func() {
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		if err = audit.GetInst().Save(projectName, time.Now(), "USER", r.Method, r.URL.String(), msg); err != nil {
			errors.Print(errors.Append(err, "Failed to save audit event"))
		}
	}()

	// Authorize API Request
	if err = jwthttp.Authorize(r, projectName, role.ResProject, role.TypeWrite); err != nil {
		errors.PrintAsInfo(errors.Append(err, "Failed to authorize header"))
		errors.WriteToHTTP(w, errors.ErrUnpermitted, 0, "")
		return
	}

	roleType := model.RoleCustom
	if _, _, ok := role.GetInst().Parse(roleID); ok {
		roleType = model.RoleSystem
	}

	if err = db.GetInst().UserAddRole(projectName, userID, roleType, roleID); err != nil {
		if errors.Contains(err, model.ErrNoSuchUser) {
			logger.Info("No such user: %s", userID)
			errors.WriteToHTTP(w, err, http.StatusNotFound, "")
		} else if errors.Contains(err, model.ErrRoleAlreadyAppended) {
			logger.Info("Role %s is already appended", roleID)
			errors.WriteToHTTP(w, err, http.StatusConflict, "")
		} else if errors.Contains(err, model.ErrUserValidateFailed) {
			if !model.ValidateUserID(userID) {
				logger.Info("UserID %s is invalid id format", userID)
				errors.WriteToHTTP(w, err, http.StatusNotFound, "")
			} else {
				// Includes role not found
				errors.PrintAsInfo(errors.Append(err, "Invalid role was specified"))
				errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
			}
		} else {
			errors.Print(errors.Append(err, "Failed to add role to user"))
			errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.Info("UserRoleAddHandler method successfully finished")
}

// UserRoleDeleteHandler ...
//   require role: write-project
func UserRoleDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	userID := vars["userID"]
	roleID := vars["roleID"]

	var err *errors.Error
	defer func() {
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		if err = audit.GetInst().Save(projectName, time.Now(), "USER", r.Method, r.URL.String(), msg); err != nil {
			errors.Print(errors.Append(err, "Failed to save audit event"))
		}
	}()

	// Authorize API Request
	if err = jwthttp.Authorize(r, projectName, role.ResProject, role.TypeWrite); err != nil {
		errors.PrintAsInfo(errors.Append(err, "Failed to authorize header"))
		errors.WriteToHTTP(w, errors.ErrUnpermitted, 0, "")
		return
	}

	// Get Previous User Info
	if err = db.GetInst().UserDeleteRole(projectName, userID, roleID); err != nil {
		if errors.Contains(err, model.ErrNoSuchUser) {
			logger.Info("No such user: %s", userID)
			errors.WriteToHTTP(w, err, http.StatusNotFound, "")
		} else if errors.Contains(err, model.ErrNoSuchRoleInUser) {
			logger.Info("User %s do not have Role %s", userID, roleID)
			errors.WriteToHTTP(w, err, http.StatusNotFound, "")
		} else if errors.Contains(err, model.ErrUserValidateFailed) {
			if !model.ValidateUserID(userID) {
				logger.Info("UserID %s is invalid id format", userID)
				errors.WriteToHTTP(w, err, http.StatusNotFound, "")
			} else {
				errors.PrintAsInfo(errors.Append(err, "Invalid ID was specified"))
				errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
			}
		} else {
			errors.Print(errors.Append(err, "Failed to delete role from user"))
			errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.Info("UserRoleDeleteHandler method successfully finished")
}

// UserResetPasswordHandler ...
//   require role: write-project
func UserResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	userID := vars["userID"]

	var err *errors.Error
	defer func() {
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		if err = audit.GetInst().Save(projectName, time.Now(), "USER", r.Method, r.URL.String(), msg); err != nil {
			errors.Print(errors.Append(err, "Failed to save audit event"))
		}
	}()

	// Authorize API Request
	if err = jwthttp.Authorize(r, projectName, role.ResProject, role.TypeWrite); err != nil {
		errors.PrintAsInfo(errors.Append(err, "Failed to authorize header"))
		errors.WriteToHTTP(w, errors.ErrUnpermitted, 0, "")
		return
	}

	var req UserResetPasswordRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		err = errors.Append(errors.ErrInvalidRequest, "Failed to decode user reset password request: %v", e)
		errors.PrintAsInfo(err)
		errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
		return
	}

	if err = db.GetInst().UserChangePassword(projectName, userID, req.Password); err != nil {
		if errors.Contains(err, model.ErrNoSuchUser) {
			logger.Info("No such user: %s", userID)
			errors.WriteToHTTP(w, err, http.StatusNotFound, "")
		} else if errors.Contains(err, model.ErrUserValidateFailed) {
			if !model.ValidateUserID(userID) {
				logger.Info("UserID %s is invalid id format", userID)
				errors.WriteToHTTP(w, err, http.StatusNotFound, "")
			} else {
				errors.PrintAsInfo(errors.Append(err, "Invalid password was specified"))
				errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
			}
		} else if errors.Contains(err, secret.ErrPasswordPolicyFailed) {
			errors.PrintAsInfo(errors.Append(err, "Invalid password was specified"))
			errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
		} else {
			errors.Print(errors.Append(err, "Failed to reset user password"))
			errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	logger.Info("UserResetPasswordHandler method successfully finished")
}

// UserUnlockHandler ...
//   require role: write-project
func UserUnlockHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]
	userID := vars["userID"]

	var err *errors.Error
	defer func() {
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		if err = audit.GetInst().Save(projectName, time.Now(), "USER", r.Method, r.URL.String(), msg); err != nil {
			errors.Print(errors.Append(err, "Failed to save audit event"))
		}
	}()

	// Authorize API Request
	if err = jwthttp.Authorize(r, projectName, role.ResProject, role.TypeWrite); err != nil {
		errors.PrintAsInfo(errors.Append(err, "Failed to authorize header"))
		errors.WriteToHTTP(w, errors.ErrUnpermitted, 0, "")
		return
	}

	user, err := db.GetInst().UserGet(projectName, userID)
	if err != nil {
		if errors.Contains(err, model.ErrNoSuchUser) {
			errors.PrintAsInfo(errors.Append(err, "User %s is not found", userID))
			errors.WriteToHTTP(w, err, http.StatusNotFound, "")
		} else if errors.Contains(err, model.ErrUserValidateFailed) {
			errors.PrintAsInfo(errors.Append(err, "Invalid user ID format"))
			errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
		} else {
			errors.Print(errors.Append(err, "Failed to get user"))
			errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		}
		return
	}

	// clear user lock state
	user.LockState = model.LockState{}
	if err = db.GetInst().UserUpdate(projectName, user); err != nil {
		errors.Print(errors.Append(err, "Failed to update user lock state"))
		errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.Info("UserUnlockHandler method successfully finished")
}
