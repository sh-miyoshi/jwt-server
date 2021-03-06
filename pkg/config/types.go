package config

// DBInfo ...
type DBInfo struct {
	Type             string `yaml:"type"`
	ConnectionString string `yaml:"connection_string"`
}

// HTTPSConfig ...
type HTTPSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert-file"`
	KeyFile  string `yaml:"key-file"`
}

// LoginResource ...
type LoginResource struct {
	IndexPage               string
	OTPVerifyPage           string
	ConsentPage             string
	DeviceLoginPage         string
	DeviceLoginCompletePage string
}

// GlobalConfig ...
type GlobalConfig struct {
	AdminName             string      `yaml:"admin_name"`
	AdminPassword         string      `yaml:"admin_password"`
	Port                  int         `yaml:"server_port"`
	BindAddr              string      `yaml:"server_bind_address"`
	HTTPSConfig           HTTPSConfig `yaml:"https"`
	LogFile               string      `yaml:"logfile"`
	ModeDebug             bool        `yaml:"debug_mode"`
	DB                    DBInfo      `yaml:"db"`
	AuditDB               DBInfo      `yaml:"audit_db"`
	LoginSessionExpiresIn uint64      `yaml:"login_session_expires_in"`
	SSOExpiresIn          uint64      `yaml:"sso_expires_in"`
	UserLoginResourceDir  string      `yaml:"user_login_page_res"`
	DBGCInterval          uint64      `yaml:"dbgc_interval"`

	SupportedResponseType  []string
	SupportedScope         []string
	LoginResource          LoginResource
	LoginStaticResourceURL string
}
