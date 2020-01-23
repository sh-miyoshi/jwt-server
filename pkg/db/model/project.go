package model

import (
	"github.com/pkg/errors"
	"time"
)

// TokenConfig ...
type TokenConfig struct {
	AccessTokenLifeSpan  uint
	RefreshTokenLifeSpan uint
	SigningAlgorithm     string
	SignPublicKey        []byte
	SignSecretKey        []byte
}

// ProjectInfo ...
type ProjectInfo struct {
	Name        string
	CreatedAt   time.Time
	TokenConfig *TokenConfig
}

const (
	// DefaultAccessTokenExpiresTimeSec is default expires time for access token(5 minutes)
	DefaultAccessTokenExpiresTimeSec = 5 * 60

	// DefaultRefreshTokenExpiresTimeSec is default expires time for refresh token(14 days)
	DefaultRefreshTokenExpiresTimeSec = 14 * 24 * 60 * 60
)

var (
	// ErrProjectAlreadyExists ...
	ErrProjectAlreadyExists = errors.New("Project Already Exists")

	// ErrNoSuchProject ...
	ErrNoSuchProject = errors.New("No such project")

	// ErrDeleteBlockedProject ...
	ErrDeleteBlockedProject = errors.New("Projects cannot be deleted")

	// ErrProjectValidateFailed ...
	ErrProjectValidateFailed = errors.New("Project Validation Failed")
)

// ProjectInfoHandler ...
type ProjectInfoHandler interface {
	Add(ent *ProjectInfo) error
	Delete(name string) error
	GetList() ([]string, error)
	Get(name string) (*ProjectInfo, error)

	// Update method updates existing project
	// It must return error if project is not found
	Update(ent *ProjectInfo) error

	// BeginTx method starts a transaction
	BeginTx() error

	// CommitTx method commits the transaction
	CommitTx() error

	// AbortTx method abort and rollback the transaction
	AbortTx() error
}

// Validate ...
func (p *ProjectInfo) Validate() error {
	if !validateProjectName(p.Name) {
		return errors.Wrap(ErrProjectValidateFailed, "Invalid Project Name format")
	}

	if !validateTokenSigningAlgorithm(p.TokenConfig.SigningAlgorithm) {
		return errors.Wrap(ErrProjectValidateFailed, "Invalid Token Signing Algorithm")
	}

	if !validateLifeSpan(p.TokenConfig.AccessTokenLifeSpan) {
		return errors.Wrap(ErrProjectValidateFailed, "Access Token Life Span must >= 1")
	}

	if !validateLifeSpan(p.TokenConfig.RefreshTokenLifeSpan) {
		return errors.Wrap(ErrProjectValidateFailed, "Refresh Token Life Span must >= 1")
	}

	return nil
}
