package tars

import "github.com/TarsCloud/TarsGo/tars/registry"

type Option func(o *options)

type options struct {
	registrar registry.Registrar
}

// Registrar returns an Option to use the Registrar
func Registrar(r registry.Registrar) Option {
	return func(o *options) {
		o.registrar = r
	}
}
