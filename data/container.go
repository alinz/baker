package data

import (
	"github.com/alinz/bake/pkg/endpoint"
)

type Container struct {
	ID       string            `json:"id"`
	Active   bool              `json:"active"`
	Addr     endpoint.Addr     `json:"addr"`
	PingAddr endpoint.HTTPAddr `json:"ping_addr"`
	Err      error             `json:"err"`
}
