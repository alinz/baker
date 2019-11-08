package gateway

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/alinz/baker"
	"github.com/alinz/baker/pkg/endpoint"
	"github.com/alinz/baker/pkg/json"
	"github.com/alinz/baker/pkg/logger"
	"github.com/alinz/baker/service"
)

type Handler struct {
	domains *Domains
}

var _ service.Consumer = (*Handler)(nil)
var _ http.Handler = (*Handler)(nil)

// Service will be called by service.Producer
// NOTE: do not call this directly
func (s *Handler) Service(service *baker.Service) error {
	if service.Config == nil || service.Config.Domain == "" {
		logger.Debug("service %s has been removed", service.Container.ID)

		// service needs to be remove from list
		s.domains.Remove(service)
		return nil
	}

	logger.Debug("service %s has been added to domain '%s' and path %s", service.Container.ID, service.Config.Domain, service.Config.Path)
	s.domains.Add(service)

	return nil
}

// Close will be called by service.Producer
// NOTE: do not call this directly
func (s *Handler) Close(err error) {
	return
}

func (s *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	paths := s.domains.Paths(r.Host)
	if paths == nil {
		json.ResponseAsError(w, http.StatusNotFound, fmt.Errorf("resource or service not found on %s", r.Host))
		return
	}

	services := paths.Services(r.URL.Path)
	if services == nil {
		json.ResponseAsError(w, http.StatusNotFound, fmt.Errorf("resource or service not found on path %s", r.URL.Path))
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

	if !service.Config.Ready {
		json.ResponseAsError(w, http.StatusTooEarly, fmt.Errorf("resource or service is not ready yet"))
		return
	}

	target, err := url.Parse(endpoint.NewHTTPAddr(service.Container.Addr, r.URL.Path).String())
	if err != nil {
		json.ResponseAsError(w, http.StatusInternalServerError, err)
		return
	}

	logger.Debug("proxied %s%s -> %s", r.Host, r.URL, target)

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Collect all directors as one wrapped one
	director := func(r *http.Request) {}
	for _, requestUpdater := range service.Config.Rules.RequestUpdaters {
		director = requestUpdater.Director(director)
	}

	originalDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		logger.Debug("Original Request URL: %s", r.URL)
		// Need to clear URL.Path to empty as target is already known
		// Also, NewSingleHostReverseProxy.Director's default
		// will try to merge target.Path and r.URL.Path
		r.URL.Path = ""
		// originalDirector needs to be called first before calling other directors
		originalDirector(r)
		logger.Debug("Request URL after applying default director: %s", r.URL)
		director(r)
		logger.Debug("Request URL after applying all directors: %s", r.URL)
	}

	if service.Container.Addr.Secure() {
		panic("not implemented yet")
		// proxy.Transport = nil
	}

	proxy.ServeHTTP(w, r)
}

func NewHandler() *Handler {
	return &Handler{
		domains: NewDomains(),
	}
}
