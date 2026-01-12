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

type Server struct {
	Port           int            `json:"port"`
	FileEntries    []FileEntry    `json:"fileEntries"`
	ContentEntries []ContentEntry `json:"contentEntries"`
}

type FileEntry struct {
	Route string `json:"route"`
	Path  string `json:"path"`

	Info Info `json:"info"`
}

type ContentEntry struct {
	Route string `json:"route"`
	Name  string `json:"name"`
	Bytes []byte `json:"bytes"`

	Info Info `json:"info"`
}

type Info struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
}
