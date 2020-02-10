package mongo

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sh-miyoshi/jwt-server/pkg/db/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// UserInfoHandler implement db.UserInfoHandler
type UserInfoHandler struct {
	session  mongo.Session
	dbClient *mongo.Client
}

// NewUserHandler ...
func NewUserHandler(dbClient *mongo.Client) (*UserInfoHandler, error) {
	res := &UserInfoHandler{
		dbClient: dbClient,
	}

	// Create Index to Project Name
	mod := mongo.IndexModel{
		Keys: bson.M{
			"id": 1, // index in ascending order
		},
		Options: options.Index().SetUnique(true),
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()

	col := res.dbClient.Database(databaseName).Collection(userCollectionName)
	_, err := col.Indexes().CreateOne(ctx, mod)

	return res, err
}

// Add ...
func (h *UserInfoHandler) Add(ent *model.UserInfo) error {
	loginSessions := []*loginSessionInfo{}
	for _, s := range ent.LoginSessions {
		loginSessions = append(loginSessions, &loginSessionInfo{
			VerifyCode:  s.VerifyCode,
			ExpiresIn:   s.ExpiresIn,
			ClientID:    s.ClientID,
			RedirectURI: s.RedirectURI,
		})
	}

	v := &userInfo{
		ID:            ent.ID,
		ProjectName:   ent.ProjectName,
		Name:          ent.Name,
		CreatedAt:     ent.CreatedAt,
		PasswordHash:  ent.PasswordHash,
		SystemRoles:   ent.SystemRoles,
		CustomRoles:   ent.CustomRoles,
		LoginSessions: loginSessions,
	}

	col := h.dbClient.Database(databaseName).Collection(userCollectionName)

	ctx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()

	_, err := col.InsertOne(ctx, v)
	if err != nil {
		return errors.Wrap(err, "Failed to insert user to mongodb")
	}

	return nil
}

// Delete ...
func (h *UserInfoHandler) Delete(userID string) error {
	col := h.dbClient.Database(databaseName).Collection(userCollectionName)
	filter := bson.D{
		{Key: "id", Value: userID},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()

	_, err := col.DeleteOne(ctx, filter)
	if err != nil {
		return errors.Wrap(err, "Failed to delete user from mongodb")
	}
	return nil
}

// GetList ...
func (h *UserInfoHandler) GetList(projectName string) ([]string, error) {
	col := h.dbClient.Database(databaseName).Collection(userCollectionName)

	filter := bson.D{
		{Key: "projectName", Value: projectName},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()

	cursor, err := col.Find(ctx, filter)
	if err != nil {
		return []string{}, errors.Wrap(err, "Failed to get user list from mongodb")
	}

	users := []userInfo{}
	if err := cursor.All(ctx, &users); err != nil {
		return []string{}, errors.Wrap(err, "Failed to get user list from mongodb")
	}

	res := []string{}
	for _, user := range users {
		res = append(res, user.ID)
	}

	return res, nil
}

// Get ...
func (h *UserInfoHandler) Get(userID string) (*model.UserInfo, error) {
	col := h.dbClient.Database(databaseName).Collection(userCollectionName)
	filter := bson.D{
		{Key: "id", Value: userID},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()

	res := &userInfo{}
	if err := col.FindOne(ctx, filter).Decode(res); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, model.ErrNoSuchUser
		}
		return nil, errors.Wrap(err, "Failed to get user from mongodb")
	}

	loginSessions := []*model.LoginSessionInfo{}
	for _, s := range res.LoginSessions {
		loginSessions = append(loginSessions, &model.LoginSessionInfo{
			VerifyCode:  s.VerifyCode,
			ExpiresIn:   s.ExpiresIn,
			ClientID:    s.ClientID,
			RedirectURI: s.RedirectURI,
		})
	}

	return &model.UserInfo{
		ID:            res.ID,
		ProjectName:   res.ProjectName,
		Name:          res.Name,
		CreatedAt:     res.CreatedAt,
		PasswordHash:  res.PasswordHash,
		SystemRoles:   res.SystemRoles,
		CustomRoles:   res.CustomRoles,
		LoginSessions: loginSessions,
	}, nil
}

// Update ...
func (h *UserInfoHandler) Update(ent *model.UserInfo) error {
	col := h.dbClient.Database(databaseName).Collection(projectCollectionName)
	filter := bson.D{
		{Key: "id", Value: ent.ID},
	}

	loginSessions := []*loginSessionInfo{}
	for _, s := range ent.LoginSessions {
		loginSessions = append(loginSessions, &loginSessionInfo{
			VerifyCode:  s.VerifyCode,
			ExpiresIn:   s.ExpiresIn,
			ClientID:    s.ClientID,
			RedirectURI: s.RedirectURI,
		})
	}

	v := &userInfo{
		ID:            ent.ID,
		ProjectName:   ent.ProjectName,
		Name:          ent.Name,
		CreatedAt:     ent.CreatedAt,
		PasswordHash:  ent.PasswordHash,
		SystemRoles:   ent.SystemRoles,
		CustomRoles:   ent.CustomRoles,
		LoginSessions: loginSessions,
	}

	updates := bson.D{
		{Key: "$set", Value: v},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()

	if _, err := col.UpdateOne(ctx, filter, updates); err != nil {
		return errors.Wrap(err, "Failed to update user in mongodb")
	}

	return nil
}

// GetByName ...
func (h *UserInfoHandler) GetByName(projectName string, userName string) (*model.UserInfo, error) {
	col := h.dbClient.Database(databaseName).Collection(userCollectionName)
	filter := bson.D{
		{Key: "projectName", Value: projectName},
		{Key: "name", Value: userName},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()

	res := &userInfo{}
	if err := col.FindOne(ctx, filter).Decode(res); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, model.ErrNoSuchUser
		}
		return nil, errors.Wrap(err, "Failed to get user by name from mongodb")
	}

	loginSessions := []*model.LoginSessionInfo{}
	for _, s := range res.LoginSessions {
		loginSessions = append(loginSessions, &model.LoginSessionInfo{
			VerifyCode:  s.VerifyCode,
			ExpiresIn:   s.ExpiresIn,
			ClientID:    s.ClientID,
			RedirectURI: s.RedirectURI,
		})
	}

	return &model.UserInfo{
		ID:            res.ID,
		ProjectName:   res.ProjectName,
		Name:          res.Name,
		CreatedAt:     res.CreatedAt,
		PasswordHash:  res.PasswordHash,
		SystemRoles:   res.SystemRoles,
		CustomRoles:   res.CustomRoles,
		LoginSessions: loginSessions,
	}, nil
}

// DeleteAll ...
func (h *UserInfoHandler) DeleteAll(projectName string) error {
	col := h.dbClient.Database(databaseName).Collection(userCollectionName)
	filter := bson.D{
		{Key: "projectName", Value: projectName},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()

	_, err := col.DeleteMany(ctx, filter)
	if err != nil {
		return errors.Wrap(err, "Failed to delete user from mongodb")
	}
	return nil
}

// AddRole ...
func (h *UserInfoHandler) AddRole(userID string, roleType model.RoleType, roleID string) error {
	col := h.dbClient.Database(databaseName).Collection(userCollectionName)
	filter := bson.D{
		{Key: "id", Value: userID},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()

	user := &userInfo{}
	if err := col.FindOne(ctx, filter).Decode(user); err != nil {
		if err == mongo.ErrNoDocuments {
			return model.ErrNoSuchUser
		}
		return errors.Wrap(err, "Failed to get user from mongodb")
	}

	if roleType == model.RoleSystem {
		for _, role := range user.SystemRoles {
			if role == roleID {
				return model.ErrRoleAlreadyAppended
			}
		}
		user.SystemRoles = append(user.SystemRoles, roleID)
	} else if roleType == model.RoleCustom {
		for _, role := range user.CustomRoles {
			if role == roleID {
				return model.ErrRoleAlreadyAppended
			}
		}
		user.CustomRoles = append(user.CustomRoles, roleID)
	}

	updates := bson.D{
		{Key: "$set", Value: user},
	}

	if _, err := col.UpdateOne(ctx, filter, updates); err != nil {
		return errors.Wrap(err, "Failed to add role to user in mongodb")
	}

	return nil
}

// DeleteRole ...
func (h *UserInfoHandler) DeleteRole(userID string, roleID string) error {
	col := h.dbClient.Database(databaseName).Collection(userCollectionName)
	filter := bson.D{
		{Key: "id", Value: userID},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()

	user := &userInfo{}
	if err := col.FindOne(ctx, filter).Decode(user); err != nil {
		if err == mongo.ErrNoDocuments {
			return model.ErrNoSuchUser
		}
		return errors.Wrap(err, "Failed to get user from mongodb")
	}

	deleted := false
	roles := []string{}
	for _, role := range user.SystemRoles {
		if role == roleID {
			deleted = true
		} else {
			roles = append(roles, role)
		}
	}

	if deleted {
		user.SystemRoles = roles
	} else {
		deleted = false
		roles = []string{}
		for _, role := range user.CustomRoles {
			if role == roleID {
				deleted = true
			} else {
				roles = append(roles, role)
			}
		}
		if !deleted {
			return model.ErrNoSuchRoleInUser
		}
		user.CustomRoles = roles
	}

	updates := bson.D{
		{Key: "$set", Value: user},
	}

	if _, err := col.UpdateOne(ctx, filter, updates); err != nil {
		return errors.Wrap(err, "Failed to add role to user in mongodb")
	}

	return nil
}

// AddLoginSession ...
func (h *UserInfoHandler) AddLoginSession(userID string, info *model.LoginSessionInfo) error {
	col := h.dbClient.Database(databaseName).Collection(userCollectionName)
	filter := bson.D{
		{Key: "id", Value: userID},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()

	user := &userInfo{}
	if err := col.FindOne(ctx, filter).Decode(user); err != nil {
		if err == mongo.ErrNoDocuments {
			return model.ErrNoSuchUser
		}
		return errors.Wrap(err, "Failed to get user from mongodb")
	}

	for _, s := range user.LoginSessions {
		if s.VerifyCode == info.VerifyCode {
			return model.ErrLoginSessionAlreadyExists
		}
	}

	user.LoginSessions = append(user.LoginSessions, &loginSessionInfo{
		VerifyCode:  info.VerifyCode,
		ExpiresIn:   info.ExpiresIn,
		ClientID:    info.ClientID,
		RedirectURI: info.RedirectURI,
	})

	updates := bson.D{
		{Key: "$set", Value: user},
	}

	if _, err := col.UpdateOne(ctx, filter, updates); err != nil {
		return errors.Wrap(err, "Failed to add login session to user in mongodb")
	}

	return nil
}

// DeleteLoginSession ...
func (h *UserInfoHandler) DeleteLoginSession(userID string, code string) error {
	col := h.dbClient.Database(databaseName).Collection(userCollectionName)
	filter := bson.D{
		{Key: "id", Value: userID},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()

	user := &userInfo{}
	if err := col.FindOne(ctx, filter).Decode(user); err != nil {
		if err == mongo.ErrNoDocuments {
			return model.ErrNoSuchUser
		}
		return errors.Wrap(err, "Failed to get user from mongodb")
	}

	found := false
	sessions := []*loginSessionInfo{}
	for _, s := range user.LoginSessions {
		if s.VerifyCode == code {
			found = true
		} else {
			sessions = append(sessions, s)
		}
	}

	if !found {
		return model.ErrNoSuchLoginSession
	}

	user.LoginSessions = sessions

	updates := bson.D{
		{Key: "$set", Value: user},
	}

	if _, err := col.UpdateOne(ctx, filter, updates); err != nil {
		return errors.Wrap(err, "Failed to delete session from user in mongodb")
	}
	return nil
}

// BeginTx ...
func (h *UserInfoHandler) BeginTx() error {
	var err error
	h.session, err = h.dbClient.StartSession()
	if err != nil {
		return err
	}
	err = h.session.StartTransaction()
	if err != nil {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
		defer cancel()
		h.session.EndSession(ctx)
		return err
	}
	return nil
}

// CommitTx ...
func (h *UserInfoHandler) CommitTx() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()

	err := h.session.CommitTransaction(ctx)
	h.session.EndSession(ctx)
	return err
}

// AbortTx ...
func (h *UserInfoHandler) AbortTx() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutSecond*time.Second)
	defer cancel()

	err := h.session.AbortTransaction(ctx)
	h.session.EndSession(ctx)
	return err
}
