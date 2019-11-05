package gateway

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/alinz/baker"
	"github.com/alinz/baker/pkg/json"
	"github.com/alinz/baker/service"
)

// Services contains collection of same services
// it also implements basic round robin getter
type Services struct {
	mux     sync.RWMutex
	store   []*baker.Service
	current int
}

// Get implements basic round robin
func (s *Services) Get() *baker.Service {
	s.mux.RLock()
	defer s.mux.RUnlock()

	if s.current >= len(s.store) {
		s.current = s.current % len(s.store)
	}

	result := s.store[s.current]
	s.current++
	return result
}

// Add a service to pool
func (s *Services) Add(service *baker.Service) {
	s.mux.Lock()
	s.store = append(s.store, service)
	s.mux.Unlock()
}

// Remove a service from pool
func (s *Services) Remove(service *baker.Service) {
	s.mux.Lock()
	for i, serv := range s.store {
		if serv.Container.ID == service.Container.ID {
			// remove item from store using index
			s.store = append(s.store[:i], s.store[i:]...)
			break
		}
	}
	s.mux.Unlock()
}

// NewServices creates services object
func NewServices() *Services {
	return &Services{
		store: make([]*baker.Service, 0),
	}
}

// Paths contains collection of services belong to particuar path
type Paths struct {
	mux        sync.RWMutex
	store      map[string]*Services
	id2Service map[string]*baker.Service
}

// Services return services object associate with given path
func (p *Paths) Services(path string) *Services {
	p.mux.RLock()
	defer p.mux.RUnlock()

	services, ok := p.store[path]
	if !ok {
		return nil
	}

	return services
}

// Add a service to pool of same path
// NOTE: don't run Remove and Add in separate goroutine
func (p *Paths) Add(service *baker.Service) {
	p.mux.Lock()
	defer p.mux.Unlock()

	// ignore any services that don't have config or empty path
	if service.Config == nil || service.Config.Path == "" {
		return
	}

	services, ok := p.store[service.Config.Path]
	if !ok {
		services = NewServices()
		p.store[service.Config.Path] = services
	}

	p.id2Service[service.Container.ID] = service
	services.Add(service)
}

// Remove service from paths
// service might have not have path. in order to find the service
// Paths uses second id2Service to locate it and pass that
// NOTE: don't run Remove and Add in separate goroutine
func (p *Paths) Remove(service *baker.Service) {
	p.mux.Lock()
	defer p.mux.Unlock()

	// first uses id to find service
	cached, ok := p.id2Service[service.Container.ID]
	if !ok {
		return
	}

	services, ok := p.store[cached.Config.Path]
	if !ok {
		// this should not happen
		// if this happens, Add method must have some serious bug
		panic("service is not presented in under Paths structure")
	}

	delete(p.id2Service, service.Container.ID)
	services.Remove(cached)
}

// NewPaths create Paths object
func NewPaths() *Paths {
	return &Paths{
		store:      make(map[string]*Services),
		id2Service: make(map[string]*baker.Service),
	}
}

// Domains contains collection of paths belong to particular domain
type Domains struct {
	mux        sync.RWMutex
	store      map[string]*Paths
	id2Service map[string]*baker.Service
}

// Paths returns Paths object for given domain
func (d *Domains) Paths(domain string) *Paths {
	d.mux.RLock()
	defer d.mux.RUnlock()

	paths, ok := d.store[domain]
	if !ok {
		return nil
	}

	return paths
}

// Add a service to pool of same domain
func (d *Domains) Add(service *baker.Service) {
	d.mux.Lock()
	defer d.mux.Unlock()

	// ignore any services that don't have config or empty domain
	if service.Config == nil || service.Config.Domain == "" {
		return
	}

	paths, ok := d.store[service.Config.Domain]
	if !ok {
		paths = NewPaths()
		d.store[service.Config.Path] = paths
	}

	d.id2Service[service.Container.ID] = service
	paths.Add(service)
}

// Remove a service from pool of same domain
func (d *Domains) Remove(service *baker.Service) {
	d.mux.Lock()
	defer d.mux.Unlock()

	// first uses id to find service
	cached, ok := d.id2Service[service.Container.ID]
	if !ok {
		return
	}

	paths, ok := d.store[cached.Config.Domain]
	if !ok {
		// this should not happen
		// if this happens, Add method must have some serious bug
		panic("service is not presented in under Domains structure")
	}

	delete(d.id2Service, service.Container.ID)
	paths.Remove(cached)
}

// NewDomains creates a Domains object
func NewDomains() *Domains {
	return &Domains{
		store:      make(map[string]*Paths),
		id2Service: make(map[string]*baker.Service),
	}
}

type Handler struct {
	domains *Domains
}

var _ service.Consumer = (*Handler)(nil)
var _ http.Handler = (*Handler)(nil)

// Service will be called by service.Producer
// NOTE: do not call this directly
func (s *Handler) Service(service *baker.Service) error {
	if service.Config == nil || service.Config.Domain == "" {
		// service needs to be remove from list
		s.domains.Remove(service)
		return nil
	}

	s.domains.Add(service)

	return nil
}

// Close will be called by service.Producer
// NOTE: do not call this directly
func (s *Handler) Close(err error) {
	return
}

func (s *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// need to grab Host and Path from request
	host := r.Host
	path := r.URL.Path

	paths := s.domains.Paths(host)
	if paths == nil {
		json.ResponseAsError(w, http.StatusNotFound, fmt.Errorf("resource or service not found"))
		return
	}

	services := paths.Services(path)
	if services == nil {
		json.ResponseAsError(w, http.StatusNotFound, fmt.Errorf("resource or service not found"))
		return
	}

	service := services.Get()
	if service == nil {
		json.ResponseAsError(w, http.StatusNotFound, fmt.Errorf("resource or service not found"))
		return
	}

	if !service.Container.Active {
		json.ResponseAsError(w, http.StatusServiceUnavailable, fmt.Errorf("resource or service is unavailable"))
		return
	}

	r.Host = service.Container.Addr.Host()
	r.URL.Host = r.Host
	r.RequestURI = ""

	// need to have X-Forwarded-For header
	// RemoteAddr contains port as well, need to remove it
	remoteAddr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		json.ResponseAsError(w, http.StatusBadRequest, err)
		return
	}
	r.Header.Set("X-Forwarded-For", remoteAddr)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		json.ResponseAsError(w, http.StatusInternalServerError, err)
		return
	}

	// Copy the headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Set(key, value)
		}
	}

	// support stream
	doneStream := make(chan struct{})
	go func() {
		flusher, ok := w.(http.Flusher)
		if !ok {
			return
		}

		for {
			select {
			case <-time.Tick(10 * time.Millisecond):
				flusher.Flush()
			case <-doneStream:
				return
			}
		}
	}()

	// Trailer support
	trailers := bytes.Buffer{}
	firstTrailer := true
	trailersMap := map[string]string{}
	for key, values := range resp.Trailer {
		if !firstTrailer {
			trailers.WriteString(",")
		}
		trailers.WriteString(key)
		firstTrailer = false

		// saves the trailers key and value for final push
		for _, value := range values {
			trailersMap[key] = value
		}
	}
	if trailers.Len() > 0 {
		// announce trailer's keys
		w.Header().Set("Trailer", trailers.String())
	}

	// Copy the status code
	w.WriteHeader(resp.StatusCode)

	// Copy the body
	io.Copy(w, resp.Body)

	// Pass the trailers values after body
	for key, value := range trailersMap {
		w.Header().Set(key, value)
	}

	// signal the stream goroutine to stop flushing
	close(doneStream)
}

func NewHandler() *Handler {
	return &Handler{
		domains: NewDomains(),
	}
}
