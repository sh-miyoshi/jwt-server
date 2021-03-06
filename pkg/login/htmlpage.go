package login

import (
	"html/template"
	"net/http"

	"github.com/sh-miyoshi/hekate/pkg/config"
	"github.com/sh-miyoshi/hekate/pkg/errors"
	"github.com/sh-miyoshi/hekate/pkg/logger"
)

// WriteUserLoginPage ...
func WriteUserLoginPage(projectName, sessionID, errMsg, state string, w http.ResponseWriter) {
	cfg := config.Get()

	tpl, err := template.ParseFiles(cfg.LoginResource.IndexPage)
	if err != nil {
		logger.Error("Failed to parse template: %v", err)
		e := errors.ErrServerError
		e.SetDescription("User Login Page maybe broken")
		errors.WriteToHTTP(w, e, 0, "")
		return
	}

	url := "/authapi/v1/project/" + projectName + "/authn/login?login_session_id=" + sessionID
	if state != "" {
		url += "&state=" + state
	}

	d := map[string]string{
		"URL":                url,
		"StaticResourcePath": cfg.LoginStaticResourceURL + "/static",
		"Error":              errMsg,
	}

	w.Header().Add("Content-Type", "text/html; charset=UTF-8")
	tpl.Execute(w, d)
}

// WriteOTPVerifyPage ...
func WriteOTPVerifyPage(projectName, sessionID, state string, w http.ResponseWriter) {
	cfg := config.Get()

	tpl, err := template.ParseFiles(cfg.LoginResource.OTPVerifyPage)
	if err != nil {
		logger.Error("Failed to parse template: %v", err)
		e := errors.ErrServerError
		e.SetDescription("User Login OTP Verify Page maybe broken")
		errors.WriteToHTTP(w, e, 0, "")
		return
	}

	url := "/authapi/v1/project/" + projectName + "/authn/otpverify?login_session_id=" + sessionID
	if state != "" {
		url += "&state=" + state
	}

	d := map[string]string{
		"StaticResourcePath": cfg.LoginStaticResourceURL + "/static",
		"URL":                url,
	}

	w.Header().Add("Content-Type", "text/html; charset=UTF-8")
	tpl.Execute(w, d)
}

// WriteConsentPage ...
func WriteConsentPage(projectName, sessionID, state string, w http.ResponseWriter) {
	cfg := config.Get()

	tpl, err := template.ParseFiles(cfg.LoginResource.ConsentPage)
	if err != nil {
		logger.Error("Failed to parse template: %v", err)
		e := errors.ErrServerError
		e.SetDescription("User Login Consent Page maybe broken")
		errors.WriteToHTTP(w, e, 0, "")
		return
	}

	url := "/authapi/v1/project/" + projectName + "/authn/consent?login_session_id=" + sessionID
	if state != "" {
		url += "&state=" + state
	}

	d := map[string]string{
		"StaticResourcePath": cfg.LoginStaticResourceURL + "/static",
		"URL":                url,
	}

	w.Header().Add("Content-Type", "text/html; charset=UTF-8")
	tpl.Execute(w, d)
}

// WriteDeviceLoginPage ...
func WriteDeviceLoginPage(projectName, errMsg string, w http.ResponseWriter) {
	cfg := config.Get()

	tpl, err := template.ParseFiles(cfg.LoginResource.DeviceLoginPage)
	if err != nil {
		logger.Error("Failed to parse template: %v", err)
		e := errors.ErrServerError
		e.SetDescription("User Device Login Page maybe broken")
		errors.WriteToHTTP(w, e, 0, "")
		return
	}

	url := "/resource/project/" + projectName + "/deviceverify"
	d := map[string]string{
		"StaticResourcePath": cfg.LoginStaticResourceURL + "/static",
		"Error":              errMsg,
		"URL":                url,
	}

	w.Header().Add("Content-Type", "text/html; charset=UTF-8")
	tpl.Execute(w, d)
}

// WriteDeviceLoginCompletePage ...
func WriteDeviceLoginCompletePage(w http.ResponseWriter) {
	cfg := config.Get()

	tpl, err := template.ParseFiles(cfg.LoginResource.DeviceLoginCompletePage)
	if err != nil {
		logger.Error("Failed to parse template: %v", err)
		e := errors.ErrServerError
		e.SetDescription("User Device Login Complete Page maybe broken")
		errors.WriteToHTTP(w, e, 0, "")
		return
	}

	d := map[string]string{
		"StaticResourcePath": cfg.LoginStaticResourceURL + "/static",
	}

	w.Header().Add("Content-Type", "text/html; charset=UTF-8")
	tpl.Execute(w, d)
}
