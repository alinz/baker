package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/alinz/baker"
	"github.com/alinz/baker/pkg/acme"
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
var _ acme.PolicyManager = (*Handler)(nil)

func normalizeHost(host string) (string, bool) {
	hasWWW := false
	if strings.HasPrefix(host, "www.") {
		hasWWW = true
		host = strings.Replace(host, "www.", "", 1)
	}
	return host, hasWWW
}

func (s *Handler) HostPolicy(ctx context.Context, host string) error {
	logger.Info("checking %s for certificate", host)

	paths := s.domains.Paths(host)
	if paths == nil {
		return fmt.Errorf("acme/autocert: only %s host is allowed", host)
	}

	return nil
}

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
	host, hasWWW := normalizeHost(r.Host)

	paths := s.domains.Paths(host)
	if paths == nil {
		json.ResponseAsError(w, http.StatusNotFound, fmt.Errorf("resource or service not found on %s", host))
		return
	}

	services := paths.Services(r.URL.Path)
	if services == nil {
		json.ResponseAsError(w, http.StatusNotFound, fmt.Errorf("resource or service not found on path %s", r.URL.Path))
		return
	}

	service := services.Get()
	if service == nil {
		json.ResponseAsError(w, http.StatusNotFound, errors.New("resource or service not found"))
		return
	}

	if !service.Config.IncludeWWW && hasWWW {
		logger.Debug("service '%s%s' not supported www subdomain", service.Config.Domain, service.Config.Path)
		json.ResponseAsError(w, http.StatusNotFound, errors.New("resource or service not found"))
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
