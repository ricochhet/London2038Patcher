package configutil

type Config struct {
	TLS     TLS      `json:"tls"`
	Servers []Server `json:"servers"`
}

type TLS struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
}

type Timeouts struct {
	ReadHeader int `json:"readHeader"`
	Read       int `json:"read"`
	Write      int `json:"write"`
	Idle       int `json:"idle"`
}

type Server struct {
	Port             int      `json:"port"`
	AllowCredentials bool     `json:"allowCredentials"`
	MaxAge           int      `json:"maxAge"`
	Timeouts         Timeouts `json:"timeouts"`

	FileEntries    []FileEntry    `json:"fileEntries"`
	ContentEntries []ContentEntry `json:"contentEntries"`
}

type FileEntry struct {
	Route string `json:"route"`
	Path  string `json:"path"`

	Info Info `json:"info"`
}

type ContentEntry struct {
	Route  string `json:"route"`
	Name   string `json:"name"`
	Base64 string `json:"base64"` // Unmarshal handles []byte as base64, so just handle the key as a string.

	Info Info `json:"info"`
}

type Info struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
}
