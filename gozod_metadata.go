package gozod

import (
	"github.com/kaptinlin/gozod/core"
	"github.com/kaptinlin/gozod/internal/checks"
)

type Registry[M any] = core.Registry[M]
type GlobalMeta = core.GlobalMeta

func NewRegistry[M any]() *Registry[M] {
	return core.NewRegistry[M]()
}

var (
	GlobalRegistry = core.GlobalRegistry
	Describe       = checks.Describe
	Meta           = checks.Meta
)
