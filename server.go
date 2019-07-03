package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sh-miyoshi/jwt-server/pkg/logger"
	"github.com/sh-miyoshi/jwt-server/pkg/userdb"
)

type globalConfig struct {
	Port          int
	BindAddr      string
	LogFile       string
	ModeDebug     bool
	AdminName     string
	AdminPassword string
}

var config globalConfig

func parseCmdlineArgs() {
	const DefaultPort = 8080
	const DefaultBindAddr = "0.0.0.0"

	flag.IntVar(&config.Port, "port", DefaultPort, "set port number for server")
	flag.StringVar(&config.BindAddr, "bind", DefaultBindAddr, "set bind address for server")
	flag.StringVar(&config.LogFile, "logfile", "", "write log to file, output os.Stdout when do not set this")
	flag.BoolVar(&config.ModeDebug, "debug", false, "if true, run server as debug mode")
	flag.StringVar(&config.AdminName, "adminuser", "admin", "user name of system admin")
	flag.StringVar(&config.AdminPassword, "adminpassword", "password", "password of system admin")
	flag.Parse()
}

func initDB(dbfile string) error {
	// TODO(use localdb)
	if err := userdb.InitUserHandler(userdb.DBLocal); err != nil {
		logger.Error("Failed to initialize DB: %v", err)
		return err
	}

	if err := userdb.GetInst().ConnectDB(dbfile); err != nil {
		logger.Error("Failed to connect DB: %v", err)
		return err
	}

	admin := userdb.UserRequest{
		Name:     config.AdminName,
		Password: config.AdminPassword,
	}
	if err := userdb.GetInst().CreateUser(admin); err != nil {
		logger.Error("Falied to create system admin: %v", err)
		return err
	}

	return nil
}

func setAPI(r *mux.Router) {
	const basePath = "/api/v1"

	// TODO Add API

	// Health Check
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}).Methods("GET")
}

func main() {
	parseCmdlineArgs()

	logger.InitLogger(config.ModeDebug, config.LogFile)

	initDB("userdb.csv")

	r := mux.NewRouter()
	setAPI(r)

	corsObj := handlers.AllowedOrigins([]string{"*"})

	addr := fmt.Sprintf("%s:%d", config.BindAddr, config.Port)
	logger.Info("start server with %s", addr)
	if err := http.ListenAndServe(addr, handlers.CORS(corsObj)(r)); err != nil {
		os.Exit(1)
	}
}
