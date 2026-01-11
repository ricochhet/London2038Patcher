package configutil

type Config struct {
	Servers []Server `json:"servers"`
}

type Server struct {
	Port    int       `json:"port"`
	Files   []File    `json:"files"`
	Content []Content `json:"content"`
}

type File struct {
	Route string `json:"route"`
	Path  string `json:"path"`
}

type Content struct {
	Route string `json:"route"`
	Name  string `json:"name"`
	Bytes []byte `json:"bytes"`
}
