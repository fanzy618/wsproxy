package config

type HTTPServerConfig struct {
	WebSocketAddr string
}

type ClientConfig struct {
	LocalAddr       string
	RemoteAddr      string
	ServerAddr      string
	InteractiveMode bool
}

type SSLConfig struct {
}

type HTTP3Config struct {
	
}

type Config struct {
	Server HTTPServerConfig
	Client ClientConfig
}
