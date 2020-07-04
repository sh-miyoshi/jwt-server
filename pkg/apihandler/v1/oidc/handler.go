package oidc

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sh-miyoshi/hekate/pkg/client"
	"github.com/sh-miyoshi/hekate/pkg/db"
	"github.com/sh-miyoshi/hekate/pkg/db/model"
	"github.com/sh-miyoshi/hekate/pkg/errors"
	jwthttp "github.com/sh-miyoshi/hekate/pkg/http"
	"github.com/sh-miyoshi/hekate/pkg/logger"
	"github.com/sh-miyoshi/hekate/pkg/oidc"
	"github.com/sh-miyoshi/hekate/pkg/oidc/token"
	"github.com/sh-miyoshi/hekate/pkg/user"
	"github.com/stretchr/stew/slice"
)

// ConfigGetHandler method return a configuration of OpenID Connect
func ConfigGetHandler(w http.ResponseWriter, r *http.Request) {
	issuer := token.GetFullIssuer(r)
	logger.Debug("Issuer: %s", issuer)

	res := Config{
		Issuer:                 issuer,
		AuthorizationEndpoint:  issuer + "/openid-connect/auth",
		TokenEndpoint:          issuer + "/openid-connect/token",
		UserinfoEndpoint:       issuer + "/openid-connect/userinfo",
		JwksURI:                issuer + "/openid-connect/certs",
		ScopesSupported:        []string{"openid"},
		ResponseTypesSupported: oidc.GetSupportedResponseType(),
		SubjectTypesSupported:  []string{"public"},
		IDTokenSigningAlgValuesSupported: []string{
			"RS256",
		},
		ClaimsSupported: []string{
			"iss",
			"aud",
			"sub",
			"exp",
			"jti",
			"iat",
			"nbf",
		},
		ResponseModesSupported: []string{
			"query",
			"fragment",
		},
	}

	jwthttp.ResponseWrite(w, "ConfigGetHandler", &res)
}

// TokenHandler ...
func TokenHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]

	if err := r.ParseForm(); err != nil {
		logger.Info("Failed to parse form: %v", err)
		errors.WriteOAuthError(w, errors.ErrInvalidRequestObject, "")
		return
	}

	logger.Debug("Form: %v", r.Form)
	state := r.Form.Get("state")

	// Get Project Info for Token Config
	project, err := db.GetInst().ProjectGet(projectName)
	if err != nil {
		if errors.Contains(err, model.ErrNoSuchProject) {
			http.Error(w, "Project Not Found", http.StatusNotFound)
		} else {
			logger.Error("Failed to get project info: %+v", err)
			errors.WriteOAuthError(w, errors.ErrServerError, state)
		}
		return
	}

	clientID := r.Form.Get("client_id")
	clientSecret := r.Form.Get("client_secret")

	if clientID == "" {
		// maybe basic authentication
		i, s, ok := r.BasicAuth()
		if !ok {
			logger.Info("Failed to get client ID from request, Request header: %v", r.Header)
			errors.WriteOAuthError(w, errors.ErrInvalidClient, state)
			return
		}
		clientID = i
		clientSecret = s
	}

	if err := oidc.ClientAuth(projectName, clientID, clientSecret); err != nil {
		if errors.Contains(err, errors.ErrInvalidClient) {
			logger.Info("Failed to authenticate client: %s", clientID)
			errors.WriteOAuthError(w, errors.ErrInvalidClient, state)
		} else {
			logger.Error("Failed to authenticate client: %+v", err)
			errors.WriteOAuthError(w, errors.ErrServerError, state)
		}
		return
	}

	var tkn *oidc.TokenResponse

	if r.Form.Get("redirect_uri") != "" {
		if err := client.CheckRedirectURL(projectName, clientID, r.Form.Get("redirect_uri")); err != nil {
			if errors.Contains(err, client.ErrNoRedirectURL) {
				logger.Info("Redirect URL %s is not in Allowed list", r.Form.Get("redirect_uri"))
				errors.WriteOAuthError(w, errors.ErrInvalidRequestURI, state)
			} else if errors.Contains(err, model.ErrNoSuchClient) {
				logger.Info("Failed to get allowed callback urls: No such client %s", clientID)
				errors.WriteOAuthError(w, errors.ErrInvalidClient, state)
			} else {
				logger.Error("Failed to get allowed callback urls in client: %+v", err)
				errors.WriteOAuthError(w, errors.ErrServerError, state)
			}
			return
		}
	}

	// Authetication
	gtStr := r.Form.Get("grant_type")
	gt, err := model.GetGrantType(gtStr)
	if err != nil {
		logger.Info("No such Grant Type: %s", gtStr)
		errors.WriteOAuthError(w, errors.ErrInvalidGrant, state)
		return
	}
	if ok := slice.Contains(project.AllowGrantTypes, gt); !ok {
		logger.Info("Grant Type %s is not in allowed list %v", gtStr, project.AllowGrantTypes)
		errors.WriteOAuthError(w, errors.ErrUnsupportedGrantType, state)
	}

	switch gt {
	case model.GrantTypeClientCredentials:
		tkn, err = oidc.ReqAuthByClientCredentials(project, clientID, r)
	case model.GrantTypePassword:
		uname := r.Form.Get("username")
		passwd := r.Form.Get("password")
		tkn, err = oidc.ReqAuthByPassword(project, uname, passwd, r)
	case model.GrantTypeRefreshToken:
		refreshToken := r.Form.Get("refresh_token")
		tkn, err = oidc.ReqAuthByRefreshToken(project, clientID, refreshToken, r)

		if err != nil && errors.Contains(err, model.ErrNoSuchSession) {
			logger.Info("Refresh token is already revoked")
			errors.WriteOAuthError(w, errors.ErrInvalidRequest, state)
			return
		}
	case model.GrantTypeAuthorizationCode:
		code := r.Form.Get("code")
		tkn, err = oidc.ReqAuthByCode(project, clientID, code, r)
	default:
		logger.Info("Unexpected grant type got: %s", gt.String())
		errors.WriteOAuthError(w, errors.ErrServerError, state)
		return
	}

	if err != nil {
		if err.GetHTTPStatusCode() != 0 {
			logger.Info("Failed to verify request: %v", err)
			errors.WriteOAuthError(w, err, state)
		} else {
			logger.Error("Failed to verify request: %v", err)
			errors.WriteOAuthError(w, errors.ErrServerError, state)
		}
		return
	}

	res := &TokenResponse{
		TokenType:        tkn.TokenType,
		AccessToken:      tkn.AccessToken,
		ExpiresIn:        tkn.ExpiresIn,
		RefreshToken:     tkn.RefreshToken,
		RefreshExpiresIn: tkn.RefreshExpiresIn,
		IDToken:          tkn.IDToken,
	}

	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Pragma", "no-cache")
	jwthttp.ResponseWrite(w, "TokenHandler", res)
}

// CertsHandler ...
func CertsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]

	project, err := db.GetInst().ProjectGet(projectName)
	if err != nil {
		if errors.Contains(err, model.ErrNoSuchProject) {
			logger.Info("No such project: %s", projectName)
			http.Error(w, "Project Not Found", http.StatusNotFound)
		} else {
			logger.Error("Failed to get project: %+v", err)
			errors.WriteOAuthError(w, errors.ErrServerError, "")
		}
		return
	}

	res, err := oidc.GenerateJWKSet(project.TokenConfig.SigningAlgorithm, project.TokenConfig.SignPublicKey)
	if err != nil {
		logger.Error("Failed to generate JWT set: %+v", err)
		errors.WriteOAuthError(w, errors.ErrServerError, "")
		return
	}

	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Pragma", "no-cache")
	jwthttp.ResponseWrite(w, "CertsHandler", res)
}

// AuthGETHandler ...
func AuthGETHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]

	// Get data form Query
	queries := r.URL.Query()
	logger.Debug("Query: %v", queries)

	authHandler(w, projectName, queries)
}

// AuthPOSTHandler ...
func AuthPOSTHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]

	// Get data form Form
	if err := r.ParseForm(); err != nil {
		logger.Info("Failed to parse form: %v", err)
		errMsg := "Request failed. invalid form value"
		oidc.WriteErrorPage(errMsg, w)
		return
	}

	logger.Debug("Form: %v", r.Form)
	authHandler(w, projectName, r.Form)
}

// UserLoginHandler ...
func UserLoginHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]

	var err *errors.Error

	// Get data form Form
	if err := r.ParseForm(); err != nil {
		logger.Info("Failed to parse form: %v", err)
		errMsg := "Request failed. invalid form value"
		oidc.WriteErrorPage(errMsg, w)
		return
	}

	logger.Debug("Form: %v", r.Form)
	state := r.Form.Get("state")

	sessionID := r.Form.Get("login_session_id")

	// delete session if login failed
	defer func() {
		if err != nil {
			db.GetInst().AuthCodeSessionDelete(projectName, sessionID)
		}
	}()

	// Verify user login session code
	s, err := oidc.VerifySession(projectName, sessionID)
	if err != nil {
		logger.Info("Failed to verify user login session: %v", err)
		errMsg := "Request failed. internal server error occured."
		if errors.Contains(err, errors.ErrSessionExpired) {
			errMsg = "Request failed. the session was already expired."
		}
		oidc.WriteErrorPage(errMsg, w)
		return
	}

	// Verify user
	uname := r.Form.Get("username")
	passwd := r.Form.Get("password")
	usr, err := user.Verify(projectName, uname, passwd)
	if err != nil {
		if errors.Contains(err, user.ErrAuthFailed) {
			logger.Info("Failed to authenticate user %s: %v", uname, err)

			// delete old session and create new code for relogin
			if err := db.GetInst().AuthCodeSessionDelete(projectName, sessionID); err != nil {
				logger.Error("Failed to delete previous login session %+v", err)
				oidc.WriteErrorPage("Request failed. internal server error occuerd", w)
				return
			}

			authReq := &oidc.AuthRequest{
				Scope:        s.Scope,
				ResponseType: s.ResponseType,
				ClientID:     s.ClientID,
				RedirectURI:  s.RedirectURI,
				State:        state,
				Nonce:        s.Nonce,
				MaxAge:       s.MaxAge,
				ResponseMode: s.ResponseMode,
				Prompt:       s.Prompt,
			}

			code, err := oidc.StartLoginSession(projectName, authReq)
			if err != nil {
				logger.Error("Failed to register login session %+v", err)
				oidc.WriteErrorPage("Request failed. internal server error occuerd", w)
				return
			}
			oidc.WriteUserLoginPage(projectName, code, "invalid user name or password", state, w)
			err = nil // do not delete session in defer function
		} else {
			logger.Error("Failed to verify user: %+v", err)
			errMsg := "Request failed. internal server error occuerd"
			oidc.WriteErrorPage(errMsg, w)
		}
		return
	}

	s.UserID = usr.ID
	s.LoginDate = time.Now()

	if ok := slice.Contains(s.Prompt, "consent"); ok {
		// save user id to session info
		if err = db.GetInst().AuthCodeSessionUpdate(projectName, s); err != nil {
			logger.Error("Failed to update auth code session: %+v", err)
			oidc.WriteErrorPage("Request failed. internal server error occuerd", w)
			return
		}

		// show consent page
		oidc.WriteConsentPage(projectName, sessionID, state, w)
		return
	}

	issuer := token.GetFullIssuer(r)
	req, errMsg := createLoginRedirectInfo(s, state, issuer)
	if errMsg != "" {
		err = errors.New(errMsg)
		logger.Error(errMsg)
		oidc.WriteErrorPage("Request failed. internal server error occuerd", w)
		return
	}

	if ok := slice.Contains(s.ResponseType, "code"); !ok {
		// delete session
		err = errors.New("session end")
	} else {
		// Update auth code session info
		if err = db.GetInst().AuthCodeSessionUpdate(projectName, s); err != nil {
			logger.Error("Failed to update auth code session: %+v", err)
			oidc.WriteErrorPage("Request failed. internal server error occuerd", w)
			return
		}
	}

	http.Redirect(w, req, req.URL.String(), http.StatusFound)
}

// ConsentHandler ...
func ConsentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]

	sel := r.FormValue("select")
	logger.Info("Consent select: %s", sel)

	// Get data form Form
	if err := r.ParseForm(); err != nil {
		logger.Info("Failed to parse form: %v", err)
		errMsg := "Request failed. invalid form value"
		oidc.WriteErrorPage(errMsg, w)
		return
	}

	logger.Debug("Form: %v", r.Form)
	state := r.Form.Get("state")

	switch sel {
	case "yes":
		sessionID := r.Form.Get("login_session_id")
		s, err := oidc.VerifySession(projectName, sessionID)
		if err != nil {
			logger.Info("Failed to verify user login session: %v", err)
			errMsg := "Request failed. internal server error occured."
			if errors.Contains(err, errors.ErrSessionExpired) {
				errMsg = "Request failed. the session was already expired."
			}
			oidc.WriteErrorPage(errMsg, w)
			return
		}

		issuer := token.GetFullIssuer(r)
		req, errMsg := createLoginRedirectInfo(s, state, issuer)
		if errMsg != "" {
			logger.Error(errMsg)
			oidc.WriteErrorPage("Request failed. internal server error occuerd", w)
			return
		}

		http.Redirect(w, req, req.URL.String(), http.StatusFound)
	case "no":
		errors.WriteOAuthError(w, errors.ErrConsentRequired, state)
	default:
		logger.Info("Invalid select type %s. consent page maybe broken.", sel)
		errors.WriteOAuthError(w, errors.ErrServerError, state)
	}
}

// UserInfoHandler ...
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]

	claims, err := jwthttp.ValidateAPIRequest(r)
	if err != nil {
		logger.Info("Failed to validate header: %v", err)
		errors.WriteOAuthError(w, errors.ErrRequestUnauthorized, "")
		return
	}

	user, err := db.GetInst().UserGet(projectName, claims.Subject)
	if err != nil {
		// If token validate accepted, user absolutely exists
		logger.Error("Failed to get user: %+v", err)
		errors.WriteOAuthError(w, errors.ErrServerError, "")
		return
	}

	res := &UserInfo{
		Subject:  claims.Subject,
		UserName: user.Name,
	}

	w.Header().Add("Cache-Control", "no-store")
	w.Header().Add("Pragma", "no-cache")
	jwthttp.ResponseWrite(w, "UserInfoHandler", res)
}

// RevokeHandler ...
func RevokeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["projectName"]

	// Get data form Form
	if err := r.ParseForm(); err != nil {
		logger.Info("Failed to parse form: %v", err)
		errors.WriteOAuthError(w, errors.ErrInvalidRequestObject, "")
		return
	}

	tokenType := r.Form.Get("token_type_hint")
	if tokenType == "" {
		tokenType = "refresh_token" // default is refresh token
	}

	switch tokenType {
	case "access_token":
		errors.WriteOAuthError(w, errors.ErrUnsupportedTokenType, r.Form.Get("state"))
	case "refresh_token":
		refreshToken := r.Form.Get("token")
		claims := &token.RefreshTokenClaims{}
		issuer := token.GetExpectIssuer(r)
		if err := token.ValidateRefreshToken(claims, refreshToken, issuer); err != nil {
			logger.Info("Failed to validate refresh token: %v", err)
			w.WriteHeader(http.StatusOK)
			return
		}

		if err := db.GetInst().SessionDelete(projectName, claims.SessionID); err != nil {
			if errors.Contains(err, model.ErrNoSuchProject) || errors.Contains(err, model.ErrNoSuchSession) || errors.Contains(err, model.ErrSessionValidateFailed) {
				logger.Info("Failed to revoke session: %v", err)
				w.WriteHeader(http.StatusOK)
			} else {
				logger.Error("Failed to revoke session: %+v", err)
				errors.WriteOAuthError(w, errors.ErrServerError, r.Form.Get("state"))
			}
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		errors.WriteOAuthError(w, errors.ErrUnsupportedTokenType, r.Form.Get("state"))
	}
}

func authHandler(w http.ResponseWriter, projectName string, req url.Values) {
	authReq := oidc.NewAuthRequest(req)
	logger.Debug("Auth Request: %v", authReq)

	if err := authReq.Validate(); err != nil {
		logger.Info("Failed to validate request: %v", err)
		errMsg := "Request failed. "
		if err.GetHTTPStatusCode() == 0 {
			errMsg += "internal server error occured."
		} else {
			errMsg += err.Error()
		}
		oidc.WriteErrorPage(errMsg, w)
		return
	}

	// Check Redirect URL
	if err := client.CheckRedirectURL(projectName, authReq.ClientID, authReq.RedirectURI); err != nil {
		errMsg := ""
		if errors.Contains(err, client.ErrNoRedirectURL) {
			logger.Info("Redirect URL %s is not in Allowed list", authReq.RedirectURI)
			errMsg = "Request failed. the redirect url is not allowed"
		} else if errors.Contains(err, model.ErrNoSuchClient) {
			logger.Info("Failed to get allowed callback urls: No such client %s", authReq.ClientID)
			errMsg = "Request faild. no such client"
		} else {
			logger.Error("Failed to get allowed callback urls in client: %+v", err)
			errMsg = "Request faild. internal server error occured"
		}
		oidc.WriteErrorPage(errMsg, w)
		return
	}

	// TODO(if already logined (check by login_hint, prompt), redirect to callback)

	// Start session for login flow
	sessionID, err := oidc.StartLoginSession(projectName, authReq)
	if err != nil {
		logger.Error("Failed to register auth code session %+v", err)
		oidc.WriteErrorPage("Request failed. internal server error occuerd", w)
		return
	}

	oidc.WriteUserLoginPage(projectName, sessionID, "", authReq.State, w)
}

func createLoginRedirectInfo(session *model.AuthCodeSession, state, tokenIssuer string) (*http.Request, string) {
	values := url.Values{}
	if state != "" {
		values.Set("state", state)
	}

	for _, typ := range session.ResponseType {
		switch typ {
		case "code":
			code := uuid.New().String()
			session.Code = code
			values.Set("code", code)
		case "id_token":
			prj, err := db.GetInst().ProjectGet(session.ProjectName)
			if err != nil {
				return nil, fmt.Sprintf("Failed to get token lifespan in project: %+v", err)
			}

			audiences := []string{session.UserID, session.ClientID}
			tokenReq := token.Request{
				Issuer:          tokenIssuer,
				ExpiredTime:     time.Second * time.Duration(prj.TokenConfig.AccessTokenLifeSpan),
				ProjectName:     session.ProjectName,
				UserID:          session.UserID,
				Nonce:           session.Nonce,
				EndUserAuthTime: session.LoginDate,
			}
			tkn, err := token.GenerateIDToken(audiences, tokenReq)
			if err != nil {
				return nil, fmt.Sprintf("Failed to generate id token: %+v", err)
			}
			values.Set("id_token", tkn)
		case "token":
			prj, err := db.GetInst().ProjectGet(session.ProjectName)
			if err != nil {
				return nil, fmt.Sprintf("Failed to get token lifespan in project: %+v", err)
			}

			audiences := []string{session.UserID, session.ClientID}
			tokenReq := token.Request{
				Issuer:      tokenIssuer,
				ExpiredTime: time.Second * time.Duration(prj.TokenConfig.AccessTokenLifeSpan),
				ProjectName: session.ProjectName,
				UserID:      session.UserID,
			}
			tkn, err := token.GenerateAccessToken(audiences, tokenReq)
			if err != nil {
				return nil, fmt.Sprintf("Failed to generate access token: %+v", err)
			}
			values.Set("access_token", tkn)
		default:
			return nil, fmt.Sprintf("Unknown response type: %s", typ)
		}
	}

	req, err := http.NewRequest("GET", session.RedirectURI, nil)
	if err != nil {
		return nil, fmt.Sprintf("Failed to create response: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if session.ResponseMode == "query" {
		req.URL.RawQuery = values.Encode()
	} else if session.ResponseMode == "fragment" {
		req.URL.Fragment = values.Encode()
	} else {
		return nil, fmt.Sprintf("Invalid response mode %s is specified.", session.ResponseMode)
	}

	return req, ""
}
