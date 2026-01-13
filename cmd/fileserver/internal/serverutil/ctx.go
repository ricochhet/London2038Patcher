package serverutil

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ricochhet/london2038patcher/cmd/fileserver/internal/configutil"
	"github.com/ricochhet/london2038patcher/pkg/contextutil"
)

type HTTPServer struct {
	Router chi.Router

	TLS      *configutil.TLS
	Timeouts *configutil.Timeouts
}

type Context struct {
	*contextutil.Context[HTTPServer]
}

// NewContext creates an empty Context.
func NewContext() *Context {
	return &Context{}
}

func (h *Context) Handle(pattern string, handler http.Handler) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	h.Get().Router.Handle(pattern, handler)
}
