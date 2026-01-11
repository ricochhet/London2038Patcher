package configutil

type Config struct {
	Servers []Server `json:"servers"`
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
