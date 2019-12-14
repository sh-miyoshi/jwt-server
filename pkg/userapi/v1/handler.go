package userapi

import (
	"github.com/gorilla/mux"
	"github.com/sh-miyoshi/jwt-server/pkg/db"
	"github.com/sh-miyoshi/jwt-server/pkg/db/model"
	jwthttp "github.com/sh-miyoshi/jwt-server/pkg/http"
	"github.com/sh-miyoshi/jwt-server/pkg/logger"
	"github.com/sh-miyoshi/jwt-server/pkg/role"
	"net/http"
)

// AllUserGetHandler ...
//   require role: project-read
func AllUserGetHandler(w http.ResponseWriter, r *http.Request) {
	// Authorize API Request
	if err := jwthttp.AuthHeader(r.Header, role.ResProject, role.TypeRead); err != nil {
		logger.Info("Failed to authorize header: %v", err)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	projectName := vars["projectName"]

	users, err := db.GetInst().User.GetList(projectName)
	if err != nil {
		if err == model.ErrNoSuchProject {
			logger.Info("No such project: %s", projectName)
			http.Error(w, "Project Not Found", http.StatusNotFound)
		} else {
			logger.Error("Failed to get project: %+v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	jwthttp.ResponseWrite(w, "UserGetAllUserGetHandlerHandler", &users)
}

// UserCreateHandler ...
func UserCreateHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Not implemented yet")
	http.Error(w, "Not Implemented yet", http.StatusInternalServerError)
}

// UserDeleteHandler ...
func UserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Not implemented yet")
	http.Error(w, "Not Implemented yet", http.StatusInternalServerError)
}

// UserGetHandler ...
//   require role: user-read
func UserGetHandler(w http.ResponseWriter, r *http.Request) {
	// Authorize API Request
	if err := jwthttp.AuthHeader(r.Header, role.ResUser, role.TypeRead); err != nil {
		logger.Info("Failed to authorize header: %v", err)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	projectName := vars["projectName"]
	userID := vars["userID"]

	user, err := db.GetInst().User.Get(projectName, userID)
	if err != nil {
		if err == model.ErrNoSuchProject {
			logger.Info("No such project: %s", projectName)
			http.Error(w, "Project Not Found", http.StatusNotFound)
		} else {
			logger.Error("Failed to get project: %+v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	res := UserGetResponse{
		ID:           user.ID,
		Name:         user.Name,
		Enabled:      user.Enabled,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt.String(),
		Roles:        user.Roles,
	}

	for _, s := range user.Sessions {
		res.Sessions = append(res.Sessions, s.SessionID)
	}

	jwthttp.ResponseWrite(w, "UserGetHandler", &res)
}

// UserUpdateHandler ...
func UserUpdateHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Not implemented yet")
	http.Error(w, "Not Implemented yet", http.StatusInternalServerError)
}

// UserRoleAddHandler ...
func UserRoleAddHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Not implemented yet")
	http.Error(w, "Not Implemented yet", http.StatusInternalServerError)
}

// UserRoleDeleteHandler ...
func UserRoleDeleteHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Not implemented yet")
	http.Error(w, "Not Implemented yet", http.StatusInternalServerError)
}
