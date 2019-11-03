package data

import (
	"github.com/alinz/bake/pkg/endpoint"
)

type Service struct {
	ID       string            `json:"id"`
	Active   bool              `json:"active"`
	Addr     endpoint.Addr     `json:"addr"`
	PingAddr endpoint.HTTPAddr `json:"ping_addr"`
	Domain   string            `json:"domain"`
	Path     string            `json:"path"`
}
