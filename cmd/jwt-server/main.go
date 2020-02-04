package main

import (
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sh-miyoshi/jwt-server/cmd/jwt-server/config"
	clientapiv1 "github.com/sh-miyoshi/jwt-server/pkg/apihandler/v1/client"
	roleapiv1 "github.com/sh-miyoshi/jwt-server/pkg/apihandler/v1/customrole"
	oidcapiv1 "github.com/sh-miyoshi/jwt-server/pkg/apihandler/v1/oidc"
	projectapiv1 "github.com/sh-miyoshi/jwt-server/pkg/apihandler/v1/project"
	userapiv1 "github.com/sh-miyoshi/jwt-server/pkg/apihandler/v1/user"
	"github.com/sh-miyoshi/jwt-server/pkg/db"
	"github.com/sh-miyoshi/jwt-server/pkg/db/model"
	"github.com/sh-miyoshi/jwt-server/pkg/logger"
	"github.com/sh-miyoshi/jwt-server/pkg/oidc"
	defaultrole "github.com/sh-miyoshi/jwt-server/pkg/role"
	"github.com/sh-miyoshi/jwt-server/pkg/token"
	"github.com/sh-miyoshi/jwt-server/pkg/util"
	"net/http"
	"os"
	"time"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("%s: %s called", r.Method, r.URL.String())
		next.ServeHTTP(w, r)
	})
}

func setAPI(r *mux.Router) {
	const basePath = "/api/v1"

	// OpenID Connect API
	r.HandleFunc(basePath+"/project/{projectName}/.well-known/openid-configuration", oidcapiv1.ConfigGetHandler).Methods("GET")
	r.HandleFunc(basePath+"/project/{projectName}/openid-connect/token", oidcapiv1.TokenHandler).Methods("POST")
	r.HandleFunc(basePath+"/project/{projectName}/openid-connect/certs", oidcapiv1.CertsHandler).Methods("GET")
	r.HandleFunc(basePath+"/project/{projectName}/openid-connect/auth", oidcapiv1.AuthGETHandler).Methods("GET")
	r.HandleFunc(basePath+"/project/{projectName}/openid-connect/auth", oidcapiv1.AuthPOSTHandler).Methods("POST")
	r.HandleFunc(basePath+"/project/{projectName}/openid-connect/userinfo", oidcapiv1.UserInfoHandler).Methods("GET", "POST")
	r.HandleFunc(basePath+"/project/{projectName}/openid-connect/login", oidcapiv1.UserLoginHandler).Methods("POST")

	// Project API
	r.HandleFunc(basePath+"/project", projectapiv1.AllProjectGetHandler).Methods("GET")
	r.HandleFunc(basePath+"/project", projectapiv1.ProjectCreateHandler).Methods("POST")
	r.HandleFunc(basePath+"/project/{projectName}", projectapiv1.ProjectDeleteHandler).Methods("DELETE")
	r.HandleFunc(basePath+"/project/{projectName}", projectapiv1.ProjectGetHandler).Methods("GET")
	r.HandleFunc(basePath+"/project/{projectName}", projectapiv1.ProjectUpdateHandler).Methods("PUT")

	// User API
	r.HandleFunc(basePath+"/project/{projectName}/user", userapiv1.AllUserGetHandler).Methods("GET")
	r.HandleFunc(basePath+"/project/{projectName}/user", userapiv1.UserCreateHandler).Methods("POST")
	r.HandleFunc(basePath+"/project/{projectName}/user/{userID}", userapiv1.UserDeleteHandler).Methods("DELETE")
	r.HandleFunc(basePath+"/project/{projectName}/user/{userID}", userapiv1.UserGetHandler).Methods("GET")
	r.HandleFunc(basePath+"/project/{projectName}/user/{userID}", userapiv1.UserUpdateHandler).Methods("PUT")
	r.HandleFunc(basePath+"/project/{projectName}/user/{userID}/role/{roleID}", userapiv1.UserRoleAddHandler).Methods("POST")
	r.HandleFunc(basePath+"/project/{projectName}/user/{userID}/role/{roleID}", userapiv1.UserRoleDeleteHandler).Methods("DELETE")

	// Client API
	r.HandleFunc(basePath+"/project/{projectName}/client", clientapiv1.AllClientGetHandler).Methods("GET")
	r.HandleFunc(basePath+"/project/{projectName}/client", clientapiv1.ClientCreateHandler).Methods("POST")
	r.HandleFunc(basePath+"/project/{projectName}/client/{clientID}", clientapiv1.ClientDeleteHandler).Methods("DELETE")
	r.HandleFunc(basePath+"/project/{projectName}/client/{clientID}", clientapiv1.ClientGetHandler).Methods("GET")
	r.HandleFunc(basePath+"/project/{projectName}/client/{clientID}", clientapiv1.ClientUpdateHandler).Methods("PUT")

	// Custom Role API
	r.HandleFunc(basePath+"/project/{projectName}/role", roleapiv1.AllRoleGetHandler).Methods("GET")
	r.HandleFunc(basePath+"/project/{projectName}/role", roleapiv1.RoleCreateHandler).Methods("POST")
	r.HandleFunc(basePath+"/project/{projectName}/role/{roleID}", roleapiv1.RoleDeleteHandler).Methods("DELETE")
	r.HandleFunc(basePath+"/project/{projectName}/role/{roleID}", roleapiv1.RoleGetHandler).Methods("GET")
	r.HandleFunc(basePath+"/project/{projectName}/role/{roleID}", roleapiv1.RoleUpdateHandler).Methods("PUT")

	// Health Check
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}).Methods("GET")

	r.Use(loggingMiddleware)
}

func initDB(dbType, connStr, adminName, adminPassword string) error {
	if err := db.InitDBManager(dbType, connStr); err != nil {
		return errors.Wrap(err, "Failed to init database manager")
	}

	// Set Master Project if not exsits
	err := db.GetInst().ProjectAdd(&model.ProjectInfo{
		Name:      "master",
		CreatedAt: time.Now(),
		TokenConfig: &model.TokenConfig{
			AccessTokenLifeSpan:  model.DefaultAccessTokenExpiresTimeSec,
			RefreshTokenLifeSpan: model.DefaultRefreshTokenExpiresTimeSec,
			SigningAlgorithm:     "RS256",
		},
	})
	if err != nil {
		if errors.Cause(err) == model.ErrProjectAlreadyExists {
			logger.Info("Master Project is already exists.")
		} else {
			return errors.Wrap(err, "Failed to create master project")
		}
	}

	err = db.GetInst().UserAdd(&model.UserInfo{
		ID:           uuid.New().String(),
		ProjectName:  "master",
		Name:         adminName,
		CreatedAt:    time.Now(),
		PasswordHash: util.CreateHash(adminPassword),
		SystemRoles:  defaultrole.GetInst().GetList(), // set all roles
	})

	if err != nil {
		if errors.Cause(err) == model.ErrUserAlreadyExists {
			logger.Info("Admin user is already exists.")
		} else {
			return errors.Wrap(err, "Failed to create admin user")
		}
	}

	err = db.GetInst().ClientAdd(&model.ClientInfo{
		ID:          "admin-cli",
		ProjectName: "master",
		AccessType:  "public",
		CreatedAt:   time.Now(),
	})

	if err != nil {
		if errors.Cause(err) == model.ErrClientAlreadyExists {
			logger.Info("admin-cli client is already exists.")
		} else {
			return errors.Wrap(err, "Failed to create admin-cli client")
		}
	}

	return nil
}

func main() {
	// Read command line args
	const defaultConfigFilePath = "./config.yaml"
	configFilePath := flag.String("config", defaultConfigFilePath, "file name of config.yaml")
	flag.Parse()

	// Read configure
	cfg, err := config.InitConfig(*configFilePath)
	if err != nil {
		fmt.Printf("Failed to set config: %v", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.InitLogger(cfg.ModeDebug, cfg.LogFile)
	logger.Debug("Start with config: %v", *cfg)

	// Initialize Default Role Handler
	if err := defaultrole.InitHandler(); err != nil {
		logger.Error("Failed to initialize default role handler: %+v", err)
		os.Exit(1)
	}

	// Initialize Token Config
	// TODO(use https)
	token.InitConfig(false)

	// Initialize OIDC Config
	oidc.InitConfig(cfg.AuthCodeExpiresTime, cfg.AuthCodeUserLoginFile)

	// Initalize Database
	if err := initDB(cfg.DB.Type, cfg.DB.ConnectionString, cfg.AdminName, cfg.AdminPassword); err != nil {
		logger.Error("Failed to initialize database: %+v", err)
		os.Exit(1)
	}

	// Setup API
	r := mux.NewRouter()
	setAPI(r)

	// Run Server
	corsObj := handlers.AllowedOrigins([]string{"*"})
	addr := fmt.Sprintf("%s:%d", cfg.BindAddr, cfg.Port)
	logger.Info("start server with %s", addr)
	if err := http.ListenAndServe(addr, handlers.CORS(corsObj)(r)); err != nil {
		logger.Error("Failed to run server: %+v", err)
		os.Exit(1)
	}
}
