package mongo

import (
	"context"
	"time"

	"github.com/sh-miyoshi/hekate/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	projectCollectionName         = "project"
	userCollectionName            = "user"
	clientCollectionName          = "client"
	sessionCollectionName         = "session"
	roleCollectionName            = "customrole"
	authcodeSessionCollectionName = "authcodesession"
	roleInUserCollectionName      = "customroleinuser"
	deviceCollectionName          = "device"

	timeoutSecond = 5
)

var (
	databaseName = "hekate"
)

// NewClient ...
func NewClient(connStr string) (*mongo.Client, *errors.Error) {
	cli, err := mongo.NewClient(options.Client().ApplyURI(connStr))
	if err != nil {
		return nil, errors.New("DB failed", "Failed to create mongo client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()

	if err := cli.Connect(ctx); err != nil {
		return nil, errors.New("DB failed", "Failed to connect to mongo: %v", err)
	}

	if err := cli.Ping(ctx, nil); err != nil {
		return nil, errors.New("DB failed", "Failed to ping to mongo: %v", err)
	}

	return cli, nil
}

// ChangeDatabase changes the database of store data
// this method should call at first
func ChangeDatabase(name string) {
	if len(name) > 0 {
		databaseName = name
	}
}
