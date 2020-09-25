package userapi

// CustomRole ...
type CustomRole struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LockState ...
type LockState struct {
	Locked            bool     `json:"locked"`
	VerifyFailedTimes []string `json:"verifyFailedTimes"`
}

// UserCreateRequest ...
type UserCreateRequest struct {
	Name        string   `json:"name"`
	Password    string   `json:"password"`
	SystemRoles []string `json:"system_roles"`
	CustomRoles []string `json:"custom_roles"`
}

// UserGetResponse ...
type UserGetResponse struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	CreatedAt   string       `json:"createdAt"`
	SystemRoles []string     `json:"system_roles"`
	CustomRoles []CustomRole `json:"custom_roles"`
	Sessions    []string     `json:"sessions"` // Array of session IDs
	LockState   LockState    `json:"lockState"`
}

// UserPutRequest ...
type UserPutRequest struct {
	Name        string   `json:"name"`
	SystemRoles []string `json:"system_roles"`
	CustomRoles []string `json:"custom_roles"`
}

// UserChangePasswordRequest ...
type UserChangePasswordRequest struct {
	Password string `json:"password"`
}
