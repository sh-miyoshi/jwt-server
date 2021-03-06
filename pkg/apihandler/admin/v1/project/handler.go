package projectapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sh-miyoshi/hekate/pkg/audit"
	"github.com/sh-miyoshi/hekate/pkg/db"
	"github.com/sh-miyoshi/hekate/pkg/db/model"
	"github.com/sh-miyoshi/hekate/pkg/errors"
	jwthttp "github.com/sh-miyoshi/hekate/pkg/http"
	"github.com/sh-miyoshi/hekate/pkg/logger"
	"github.com/sh-miyoshi/hekate/pkg/role"
)

// AllProjectGetHandler ...
//   require role: read-cluster
func AllProjectGetHandler(w http.ResponseWriter, r *http.Request) {
	// Authorize API Request
	if err := jwthttp.Authorize(r, "", role.ResCluster, role.TypeRead); err != nil {
		errors.PrintAsInfo(errors.Append(err, "Failed to authorize header"))
		errors.WriteToHTTP(w, errors.ErrUnpermitted, 0, "")
		return
	}

	projects, err := db.GetInst().ProjectGetList(nil)
	if err != nil {
		errors.Print(errors.Append(err, "Failed to get project list"))
		errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		return
	}

	res := []ProjectGetResponse{}
	for _, prj := range projects {
		grantTypes := []string{}
		for _, t := range prj.AllowGrantTypes {
			grantTypes = append(grantTypes, string(t))
		}

		res = append(res, ProjectGetResponse{
			Name:      prj.Name,
			CreatedAt: prj.CreatedAt.Format(time.RFC3339),
			TokenConfig: TokenConfig{
				AccessTokenLifeSpan:  prj.TokenConfig.AccessTokenLifeSpan,
				RefreshTokenLifeSpan: prj.TokenConfig.RefreshTokenLifeSpan,
				SigningAlgorithm:     prj.TokenConfig.SigningAlgorithm,
			},
			PasswordPolicy: PasswordPolicy{
				MinimumLength:       prj.PasswordPolicy.MinimumLength,
				NotUserName:         prj.PasswordPolicy.NotUserName,
				BlackList:           prj.PasswordPolicy.BlackList,
				UseCharacter:        string(prj.PasswordPolicy.UseCharacter),
				UseDigit:            prj.PasswordPolicy.UseDigit,
				UseSpecialCharacter: prj.PasswordPolicy.UseSpecialCharacter,
			},
			AllowGrantTypes: grantTypes,
			UserLock: UserLock{
				Enabled:          prj.UserLock.Enabled,
				MaxLoginFailure:  prj.UserLock.MaxLoginFailure,
				LockDuration:     prj.UserLock.LockDuration,
				FailureResetTime: prj.UserLock.FailureResetTime,
			},
		})
	}
	logger.Debug("Project List: %v", res)

	jwthttp.ResponseWrite(w, "AllProjectGetHandler", &res)
}

// ProjectCreateHandler ...
//   require role: write-cluster
func ProjectCreateHandler(w http.ResponseWriter, r *http.Request) {
	var err *errors.Error
	defer func() {
		// register project is access user project (maybe master)
		claims, e := jwthttp.ValidateAPIToken(r)
		if e != nil {
			err = e
		}

		msg := ""
		if err != nil {
			msg = err.Error()
		}
		if err = audit.GetInst().Save(claims.Project, time.Now(), "PROJECT", r.Method, r.URL.String(), msg); err != nil {
			errors.Print(errors.Append(err, "Failed to save audit event"))
		}
	}()

	// Authorize API Request
	if err = jwthttp.Authorize(r, "", role.ResCluster, role.TypeWrite); err != nil {
		errors.PrintAsInfo(errors.Append(err, "Failed to authorize header"))
		errors.WriteToHTTP(w, errors.ErrUnpermitted, 0, "")
		return
	}

	// Parse Request
	var request ProjectCreateRequest
	if e := json.NewDecoder(r.Body).Decode(&request); e != nil {
		err = errors.New("Invalid request", "Failed to decode project create request: %v", e)
		errors.PrintAsInfo(err)
		errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
		return
	}

	// Set Allow Grant Type List
	grantTypes := []model.GrantType{}
	for _, t := range request.AllowGrantTypes {
		v, err := model.GetGrantType(t)
		if err != nil {
			errors.PrintAsInfo(errors.Append(err, "Failed to get grant type %s", t))
			errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
			return
		}
		grantTypes = append(grantTypes, v)
	}

	// Create Project Entry
	project := model.ProjectInfo{
		Name:         request.Name,
		CreatedAt:    time.Now(),
		PermitDelete: true,
		TokenConfig: &model.TokenConfig{
			AccessTokenLifeSpan:  request.TokenConfig.AccessTokenLifeSpan,
			RefreshTokenLifeSpan: request.TokenConfig.RefreshTokenLifeSpan,
			SigningAlgorithm:     request.TokenConfig.SigningAlgorithm,
		},
		PasswordPolicy: model.PasswordPolicy{
			MinimumLength:       request.PasswordPolicy.MinimumLength,
			NotUserName:         request.PasswordPolicy.NotUserName,
			BlackList:           request.PasswordPolicy.BlackList,
			UseCharacter:        model.CharacterType(request.PasswordPolicy.UseCharacter),
			UseDigit:            request.PasswordPolicy.UseDigit,
			UseSpecialCharacter: request.PasswordPolicy.UseSpecialCharacter,
		},
		AllowGrantTypes: grantTypes,
		UserLock: model.UserLock{
			Enabled:          request.UserLock.Enabled,
			MaxLoginFailure:  request.UserLock.MaxLoginFailure,
			LockDuration:     request.UserLock.LockDuration,
			FailureResetTime: request.UserLock.FailureResetTime,
		},
	}

	// Create New Project
	if err = db.GetInst().ProjectAdd(&project); err != nil {
		if errors.Contains(err, model.ErrProjectAlreadyExists) {
			logger.Info("Project %s is already exists", request.Name)
			errors.WriteToHTTP(w, err, http.StatusConflict, "")
		} else if errors.Contains(err, model.ErrProjectValidateFailed) {
			errors.PrintAsInfo(errors.Append(err, "Invalid project entry is specified"))
			errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
		} else {
			errors.Print(errors.Append(err, "Failed to create project"))
			errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		}
		return
	}

	// Return Response
	res := ProjectGetResponse{
		Name:      project.Name,
		CreatedAt: project.CreatedAt.Format(time.RFC3339),
		TokenConfig: TokenConfig{
			AccessTokenLifeSpan:  project.TokenConfig.AccessTokenLifeSpan,
			RefreshTokenLifeSpan: project.TokenConfig.RefreshTokenLifeSpan,
			SigningAlgorithm:     project.TokenConfig.SigningAlgorithm,
		},
		PasswordPolicy: PasswordPolicy{
			MinimumLength:       project.PasswordPolicy.MinimumLength,
			NotUserName:         project.PasswordPolicy.NotUserName,
			BlackList:           project.PasswordPolicy.BlackList,
			UseCharacter:        string(project.PasswordPolicy.UseCharacter),
			UseDigit:            project.PasswordPolicy.UseDigit,
			UseSpecialCharacter: project.PasswordPolicy.UseSpecialCharacter,
		},
		AllowGrantTypes: request.AllowGrantTypes,
		UserLock: UserLock{
			Enabled:          project.UserLock.Enabled,
			MaxLoginFailure:  project.UserLock.MaxLoginFailure,
			LockDuration:     project.UserLock.LockDuration,
			FailureResetTime: project.UserLock.FailureResetTime,
		},
	}

	jwthttp.ResponseWrite(w, "ProjectCreateHandler", &res)
}

// ProjectDeleteHandler ...
//   require role: write-cluster
func ProjectDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]

	var err *errors.Error
	defer func() {
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		if err = audit.GetInst().Save(projectName, time.Now(), "PROJECT", r.Method, r.URL.String(), msg); err != nil {
			errors.Print(errors.Append(err, "Failed to save audit event"))
		}
	}()

	// Authorize API Request
	if err = jwthttp.Authorize(r, projectName, role.ResCluster, role.TypeWrite); err != nil {
		errors.PrintAsInfo(errors.Append(err, "Failed to authorize header"))
		errors.WriteToHTTP(w, errors.ErrUnpermitted, 0, "")
		return
	}

	if err = db.GetInst().ProjectDelete(projectName); err != nil {
		if errors.Contains(err, model.ErrDeleteBlockedProject) {
			errors.PrintAsInfo(errors.Append(err, "Failed to delete blocked project"))
			errors.WriteToHTTP(w, errors.ErrUnpermitted, 0, "")
		} else {
			errors.Print(errors.Append(err, "Failed to delete project"))
			errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		}
		return
	}

	// Return 204 (No content) for success
	w.WriteHeader(http.StatusNoContent)
	logger.Info("ProjectDeleteHandler method successfully finished")
}

// ProjectGetHandler ...
//   require role: read-project
func ProjectGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]

	// Authorize API Request
	if err := jwthttp.Authorize(r, projectName, role.ResProject, role.TypeRead); err != nil {
		errors.PrintAsInfo(errors.Append(err, "Failed to authorize header"))
		errors.WriteToHTTP(w, errors.ErrUnpermitted, 0, "")
		return
	}

	// Get Project
	project, err := db.GetInst().ProjectGet(projectName)
	if err != nil {
		errors.Print(errors.Append(err, "Failed to get project"))
		errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		return
	}

	grantTypes := []string{}
	for _, t := range project.AllowGrantTypes {
		grantTypes = append(grantTypes, string(t))
	}

	// Return Response
	res := ProjectGetResponse{
		Name:      project.Name,
		CreatedAt: project.CreatedAt.Format(time.RFC3339),
		TokenConfig: TokenConfig{
			AccessTokenLifeSpan:  project.TokenConfig.AccessTokenLifeSpan,
			RefreshTokenLifeSpan: project.TokenConfig.RefreshTokenLifeSpan,
			SigningAlgorithm:     project.TokenConfig.SigningAlgorithm,
		},
		PasswordPolicy: PasswordPolicy{
			MinimumLength:       project.PasswordPolicy.MinimumLength,
			NotUserName:         project.PasswordPolicy.NotUserName,
			BlackList:           project.PasswordPolicy.BlackList,
			UseCharacter:        string(project.PasswordPolicy.UseCharacter),
			UseDigit:            project.PasswordPolicy.UseDigit,
			UseSpecialCharacter: project.PasswordPolicy.UseSpecialCharacter,
		},
		AllowGrantTypes: grantTypes,
		UserLock: UserLock{
			Enabled:          project.UserLock.Enabled,
			MaxLoginFailure:  project.UserLock.MaxLoginFailure,
			LockDuration:     project.UserLock.LockDuration,
			FailureResetTime: project.UserLock.FailureResetTime,
		},
	}

	jwthttp.ResponseWrite(w, "ProjectGetHandler", &res)
}

// ProjectUpdateHandler ...
//   require role: write-project
func ProjectUpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]

	var err *errors.Error
	defer func() {
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		if err = audit.GetInst().Save(projectName, time.Now(), "PROJECT", r.Method, r.URL.String(), msg); err != nil {
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
	var request ProjectPutRequest
	if e := json.NewDecoder(r.Body).Decode(&request); e != nil {
		err = errors.New("Invalid request", "Failed to decode project update request: %v", e)
		errors.PrintAsInfo(err)
		errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
		return
	}

	// Get Previous Project Info
	project, err := db.GetInst().ProjectGet(projectName)
	if err != nil {
		errors.Print(errors.Append(err, "Failed to get project"))
		errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		return
	}

	// Update Parameters
	project.TokenConfig.AccessTokenLifeSpan = request.TokenConfig.AccessTokenLifeSpan
	project.TokenConfig.RefreshTokenLifeSpan = request.TokenConfig.RefreshTokenLifeSpan
	project.TokenConfig.SigningAlgorithm = request.TokenConfig.SigningAlgorithm
	project.PasswordPolicy.MinimumLength = request.PasswordPolicy.MinimumLength
	project.PasswordPolicy.NotUserName = request.PasswordPolicy.NotUserName
	project.PasswordPolicy.BlackList = request.PasswordPolicy.BlackList
	project.PasswordPolicy.UseCharacter = model.CharacterType(request.PasswordPolicy.UseCharacter)
	project.PasswordPolicy.UseDigit = request.PasswordPolicy.UseDigit
	project.PasswordPolicy.UseSpecialCharacter = request.PasswordPolicy.UseSpecialCharacter
	project.AllowGrantTypes = []model.GrantType{}
	for _, t := range request.AllowGrantTypes {
		v, err := model.GetGrantType(t)
		if err != nil {
			errors.PrintAsInfo(errors.Append(err, "Failed to get grant type %s", t))
			errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
			return
		}
		project.AllowGrantTypes = append(project.AllowGrantTypes, v)
	}
	project.UserLock = model.UserLock{
		Enabled:          request.UserLock.Enabled,
		MaxLoginFailure:  request.UserLock.MaxLoginFailure,
		LockDuration:     request.UserLock.LockDuration,
		FailureResetTime: request.UserLock.FailureResetTime,
	}

	// Update DB
	if err = db.GetInst().ProjectUpdate(project); err != nil {
		if errors.Contains(err, model.ErrProjectValidateFailed) {
			errors.PrintAsInfo(errors.Append(err, "Project info validation failed"))
			errors.WriteToHTTP(w, err, http.StatusBadRequest, "")
		} else {
			errors.Print(errors.Append(err, "Failed to update project"))
			errors.WriteToHTTP(w, err, http.StatusInternalServerError, "")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.Info("ProjectUpdateHandler method successfully finished")
}
