package bake

import (
	"github.com/alinz/baker/pkg/endpoint"
)

type Config struct {
	Domain string `json:"domain"`
	Path   string `json:"path"`
}

type Container struct {
	ID       string            `json:"id"`
	Active   bool              `json:"active"`
	Addr     endpoint.Addr     `json:"addr"`
	PingAddr endpoint.HTTPAddr `json:"ping_addr"`
	Err      error             `json:"error"`
}

type Service struct {
	Container *Container `json:"container"`
	Config    *Config    `json:"config"`
	Err       error      `json:"error"`
}
