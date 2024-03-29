package gateway

import (
	"sync"

	"github.com/alinz/baker"
	"github.com/alinz/baker/pkg/trie"
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

	max := len(s.store)
	if max == 0 {
		return nil
	}

	if s.current >= max {
		s.current = 0
	}

	result := s.store[s.current]
	s.current++
	return result
}

// Add a service to pool
func (s *Services) Add(service *baker.Service) {
	s.mux.Lock()
	defer s.mux.Unlock()
	// need to make sure not adding multiple same id container
	for _, s := range s.store {
		if s.Container.ID == service.Container.ID {
			return
		}
	}
	s.store = append(s.store, service)
}

// Remove a service from pool
func (s *Services) Remove(service *baker.Service) {
	s.mux.Lock()
	for i, serv := range s.store {
		if serv.Container.ID == service.Container.ID {
			// remove item from store using index
			s.store = append(s.store[:i], s.store[i+1:]...)
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
	store      trie.Store
	id2Service map[string]*baker.Service
}

// Services return services object associate with given path
func (p *Paths) Services(path string) *Services {
	p.mux.RLock()
	defer p.mux.RUnlock()

	services, err := p.store.Search([]byte(path))
	if err == trie.ErrNotFound {
		return nil
	}

	return services.(*Services)
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

	key := []byte(service.Config.Path)

	services, err := p.store.Search(key)
	if err == trie.ErrNotFound {
		services = NewServices()
		p.store.Insert(key, services)
	}

	p.id2Service[service.Container.ID] = service
	services.(*Services).Add(service)
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

	delete(p.id2Service, service.Container.ID)

	key := []byte(cached.Config.Path)

	value, err := p.store.Search(key)
	if err != nil {
		panic(err)
	}

	services := value.(*Services)
	services.Remove(cached)

	if len(services.store) == 0 {
		p.store.Remove(key)
	}
}

// NewPaths create Paths object
func NewPaths() *Paths {
	return &Paths{
		store:      trie.New(),
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
		d.store[service.Config.Domain] = paths
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
