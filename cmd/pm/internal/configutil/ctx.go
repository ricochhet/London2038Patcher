package configutil

import (
	"github.com/ricochhet/london2038patcher/pkg/contextutil"
)

type Context struct {
	*contextutil.Context[Config]
}

func NewContext() *Context {
	return &Context{
		Context: &contextutil.Context[Config]{},
	}
}
