package config

// Root is config root
type Root struct {
	Logger  *Logger `toml:"logger"`
	Server  *Server
	Proxy   *Proxy
	API     *API
	Auth    *Auth
	MySQL   *MySQL `toml:"mysql" validate:"required"`
	Redis   *Redis `toml:"redis" validate:"required"`
	Develop *Develop
}

// Logger is zap logger property
type Logger struct {
	Service      string `toml:"service" validate:"required"`
	Env          string `toml:"env" validate:"oneof=dev prod custom"`
	Level        string `toml:"level" validate:"required"`
	IsStackTrace bool   `toml:"is_stacktrace"`
}

// Server is web server property
type Server struct {
	IsRelease bool      `toml:"is_release"`
	Scheme    string    `toml:"scheme" validate:"required"`
	Host      string    `toml:"host" validate:"required"`
	Port      int       `toml:"port" validate:"required"`
	Docs      Docs      `toml:"docs"`
	Session   Session   `toml:"session" validate:"required"`
	BasicAuth BasicAuth `toml:"basic_auth" validate:"required"`
}

// Docs is document root path of webserver
type Docs struct {
	Path string `toml:"path"`
}

// Session is for session property
type Session struct {
	Name     string `toml:"name"`
	Key      string `toml:"key"`
	MaxAge   int    `toml:"max_age"`
	Secure   bool   `toml:"secure"`
	HTTPOnly bool   `toml:"http_only"`
}

// BasicAuth is Basic Auth property
type BasicAuth struct {
	User string `toml:"user"`
	Pass string `toml:"pass"`
}

// Proxy is reverse proxy server property
type Proxy struct {
	Mode   uint8       `toml:"mode"` // 0:off, 1:go, 2,nginx
	Server ProxyServer `toml:"server"`
}

// ProxyServer is reverse proxy server property
type ProxyServer struct {
	Scheme  string `toml:"scheme"`
	Host    string `toml:"host"`
	Port    int    `toml:"port"`
	WebPort []int  `toml:"web_port"`
}

// API is Rest API property
type API struct {
	Ajax   bool    `toml:"only_ajax"`
	CORS   *CORS   `toml:"cors"`
	Header *Header `toml:"header"`
	JWT    *JWT    `toml:"jwt"`
}

// CORS is CORS property
type CORS struct {
	Enabled     bool     `toml:"enabled"`
	Origins     []string `toml:"origins"`
	Headers     []string `toml:"headers"`
	Methods     []string `toml:"methods"`
	Credentials bool     `toml:"credentials"`
}

// Header is original http header property for authentication
type Header struct {
	Enabled bool   `toml:"enabled"`
	Header  string `toml:"header"`
	Key     string `toml:"key"`
}

// JWT is JWT Auth property
type JWT struct {
	Mode       uint8  `toml:"mode"` // 0:off, 1:HMAC, 2:RSA
	Secret     string `toml:"secret_code"`
	PrivateKey string `toml:"private_key"`
	PublicKey  string `toml:"public_key"`
}

// Auth is authentication property for OAuth2
type Auth struct {
	Google   *Google   `toml:"google"`
	Facebook *Facebook `toml:"facebook"`
}

// Google is Google OAuth2 property
type Google struct {
	Encrypted    bool   `toml:"encrypted"`
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
	CallbackURL  string `toml:"call_back_url"`
}

// Facebook is Facebook OAuth2 property
type Facebook struct {
	Encrypted    bool   `toml:"encrypted"`
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
	CallbackURL  string `toml:"call_back_url"`
}

// MySQL is MySQL Server property
type MySQL struct {
	*MySQLContent
	Test *MySQLContent `toml:"test"`
}

// MySQLContent is MySQL Server property
type MySQLContent struct {
	Encrypted  bool   `toml:"encrypted"`
	Host       string `toml:"host"`
	Port       uint16 `toml:"port"`
	DBName     string `toml:"dbname"`
	User       string `toml:"user"`
	Pass       string `toml:"pass"`
	IsDebugLog bool   `toml:"is_debug_log"`
}

// Redis is Redis Server property
type Redis struct {
	Encrypted bool   `toml:"encrypted"`
	Host      string `toml:"host"`
	Port      uint16 `toml:"port"`
	Pass      string `toml:"pass"`
	IsSession bool   `toml:"is_session"`
	IsHeroku  bool   `toml:"is_heroku"`
}

// Develop is development use setting
type Develop struct {
	ProfileEnable bool `toml:"profile_enable"`
	RecoverEnable bool `toml:"recover_enable"`
}