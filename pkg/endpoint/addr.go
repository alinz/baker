package endpoint

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type Addr interface {
	Host() string
	Port() int
	Secure() bool
	String() string
}

type HTTPAddr interface {
	Addr
	Path() string
	WithPath(p string) HTTPAddr
}

type addr struct {
	host   string
	port   int
	secure bool
}

func (a *addr) Host() string {
	return a.host
}

func (a *addr) Port() int {
	return a.port
}

func (a *addr) Secure() bool {
	return a.secure
}

func (a *addr) String() string {
	return fmt.Sprintf("%s:%d", a.host, a.port)
}

func NewAddr(host string, port int, secure bool) Addr {
	return &addr{
		host:   host,
		port:   port,
		secure: secure,
	}
}

type httpAddr struct {
	Addr
	path string
}

func (h *httpAddr) Path() string {
	return h.path
}

func (h *httpAddr) WithPath(path string) HTTPAddr {
	return NewHTTPAddr(h.Addr, path)
}

func (h *httpAddr) String() string {
	var buffer bytes.Buffer

	buffer.WriteString("http")

	if h.Secure() {
		buffer.WriteString("s")
	}

	buffer.WriteString("://")
	buffer.WriteString(h.Addr.String())

	if !strings.HasPrefix(h.path, "/") {
		buffer.WriteString("/")
	}

	buffer.WriteString(h.path)

	return buffer.String()
}

func NewHTTPAddr(addr Addr, path string) HTTPAddr {
	return &httpAddr{
		Addr: addr,
		path: path,
	}
}

func ParseHTTPAddr(u string) HTTPAddr {
	url, err := url.Parse(u)
	if err != nil {
		return nil
	}

	port, _ := strconv.ParseInt(url.Port(), 10, 32)
	if port == 0 {
		port = 80
	}

	addr := &addr{
		host: url.Hostname(),
		port: int(port),
	}

	if url.Scheme == "https" {
		addr.secure = true
	}

	var path string
	if url.Path != "/" {
		path = url.Path
	}

	return NewHTTPAddr(addr, path)
}
