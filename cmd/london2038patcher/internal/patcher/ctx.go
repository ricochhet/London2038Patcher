package patcher

import (
	"github.com/ricochhet/london2038patcher/pkg/contextutil"
)

type Context struct {
	*contextutil.Context[Patcher]
}

func NewContext() *Context {
	return &Context{
		Context: &contextutil.Context[Patcher]{},
	}
}
