package tokenapi

import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sh-miyoshi/jwt-server/pkg/logger"
	// "github.com/sh-miyoshi/jwt-server/pkg/db"
	// "github.com/sh-miyoshi/jwt-server/pkg/db/model"
)

// TokenCreateHandler method create JWT token
func TokenCreateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["projectID"]

	// TODO(Validate project ID)

	// Parse Request
	var request TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Info(projectID, "Failed to decode token create request: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// TODO(Validate Request)

	// project, err := db.GetInst().Project.Get(projectID)
	// if err != nil {
	// 	if err == model.ErrNoSuchProject {
	// 		http.Error(w, "Project Not Found", http.StatusNotFound)
	// 	}else{
	// 		logger.Error("Failed to get project info: %v", err)
	// 		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	// 	}
	// 	return
	// }

	// // Password Authenticate
	// // TODO(hash password)

	// switch request.AuthType {
	// case: "password"
	// default:
	// 	logger.Error("Invalid Authentication Type: %s", request.AuthType)
	// 	http.Error(w, "Bad Request", http.StatusBadRequest)
	// 	return
	// }

	// Return JWT Token
	logger.Info("TokenCreateHandler method is not implemented yet")
	http.Error(w, "Not Implemented yet", http.StatusInternalServerError)
}
