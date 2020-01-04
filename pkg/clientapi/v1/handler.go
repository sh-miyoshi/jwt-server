package clientapi

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/sh-miyoshi/jwt-server/pkg/db"
	"github.com/sh-miyoshi/jwt-server/pkg/db/model"
	jwthttp "github.com/sh-miyoshi/jwt-server/pkg/http"
	"github.com/sh-miyoshi/jwt-server/pkg/logger"
	"github.com/sh-miyoshi/jwt-server/pkg/role"
	"net/http"
	"time"
)

// AllClientGetHandler ...
//   require role: project-read
func AllClientGetHandler(w http.ResponseWriter, r *http.Request) {
	// Authorize API Request
	if err := jwthttp.AuthHeader(r, role.ResProject, role.TypeRead); err != nil {
		logger.Info("Failed to authorize header: %v", err)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	projectName := vars["projectName"]

	clients, err := db.GetInst().ClientGetList(projectName)
	if err != nil {
		if err == model.ErrNoSuchProject {
			logger.Info("No such project: %s", projectName)
			http.Error(w, "Project Not Found", http.StatusNotFound)
		} else {
			logger.Error("Failed to get client: %+v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	jwthttp.ResponseWrite(w, "ClientGetAllClientGetHandlerHandler", &clients)
}

// ClientCreateHandler ...
//   require role: project-write
func ClientCreateHandler(w http.ResponseWriter, r *http.Request) {
	// Authorize API Request
	if err := jwthttp.AuthHeader(r, role.ResProject, role.TypeWrite); err != nil {
		logger.Info("Failed to authorize header: %v", err)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	projectName := vars["projectName"]

	// Parse Request
	var request ClientCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Info("Failed to decode client create request: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// TODO(Validate Request)

	// Create Client Entry
	client := model.ClientInfo{
		ID:          request.ID,
		ProjectName: projectName,
		Secret:      request.Secret,
		AccessType:  request.AccessType,
		CreatedAt:   time.Now(),
	}

	if err := db.GetInst().ClientAdd(&client); err != nil {
		if err == model.ErrNoSuchProject {
			logger.Info("No such project: %s", projectName)
			http.Error(w, "Project Not Found", http.StatusNotFound)
		} else if err == model.ErrClientAlreadyExists {
			logger.Info("Client %s is already exists", client.ID)
			http.Error(w, "Client already exists", http.StatusConflict)
		} else {
			logger.Error("Failed to create client: %+v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Return Response
	res := ClientGetResponse{
		ID:         client.ID,
		Secret:     client.Secret,
		AccessType: client.AccessType,
		CreatedAt:  client.CreatedAt.String(),
	}

	jwthttp.ResponseWrite(w, "ClientGetAllClientGetHandlerHandler", &res)
}

// ClientDeleteHandler ...
//   require role: project-write
func ClientDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// Authorize API Request
	if err := jwthttp.AuthHeader(r, role.ResProject, role.TypeWrite); err != nil {
		logger.Info("Failed to authorize header: %v", err)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	projectName := vars["projectName"]
	clientID := vars["clientID"]

	if err := db.GetInst().ClientDelete(clientID); err != nil {
		if err == model.ErrNoSuchProject {
			logger.Info("No such project: %s", projectName)
			http.Error(w, "Project Not Found", http.StatusNotFound)
		} else if err == model.ErrNoSuchClient {
			logger.Info("No such client: %s", clientID)
			http.Error(w, "Client Not Found", http.StatusNotFound)
		} else {
			logger.Error("Failed to delete client: %+v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Return 204 (No content) for success
	w.WriteHeader(http.StatusNoContent)
	logger.Info("ClientDeleteHandler method successfully finished")
}

// ClientGetHandler ...
//   require role: client-read
func ClientGetHandler(w http.ResponseWriter, r *http.Request) {
	// Authorize API Request
	if err := jwthttp.AuthHeader(r, role.ResClient, role.TypeRead); err != nil {
		logger.Info("Failed to authorize header: %v", err)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	projectName := vars["projectName"]
	clientID := vars["clientID"]

	client, err := db.GetInst().ClientGet(clientID)
	if err != nil {
		if err == model.ErrNoSuchProject {
			logger.Info("No such project: %s", projectName)
			http.Error(w, "Project Not Found", http.StatusNotFound)
		} else {
			logger.Error("Failed to get client: %+v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	res := ClientGetResponse{
		ID:         client.ID,
		Secret:     client.Secret,
		AccessType: client.AccessType,
		CreatedAt:  client.CreatedAt.String(),
	}

	jwthttp.ResponseWrite(w, "ClientGetHandler", &res)
}

// ClientUpdateHandler ...
//   require role: client-write
func ClientUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// Authorize API Request
	if err := jwthttp.AuthHeader(r, role.ResClient, role.TypeWrite); err != nil {
		logger.Info("Failed to authorize header: %v", err)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	projectName := vars["projectName"]
	clientID := vars["clientID"]

	// Parse Request
	var request ClientPutRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Info("Failed to decode client update request: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Get Previous Client Info
	client, err := db.GetInst().ClientGet(clientID)
	if err != nil {
		if err == model.ErrNoSuchProject {
			logger.Info("No such project: %s", projectName)
			http.Error(w, "Project Not Found", http.StatusNotFound)
		} else if err == model.ErrNoSuchClient {
			logger.Info("No such client: %s", clientID)
			http.Error(w, "Client Not Found", http.StatusNotFound)
		} else {
			logger.Error("Failed to update client: %+v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	// Update Parameters
	client.Secret = request.Secret
	client.AccessType = request.AccessType

	// Update DB
	if err := db.GetInst().ClientUpdate(client); err != nil {
		logger.Error("Failed to update client: %+v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.Info("ClientUpdateHandler method successfully finished")
}