package service

import (
	"context"
	"sync"
	"time"

	"github.com/alinz/baker"
	"github.com/alinz/baker/container"
	"github.com/alinz/baker/pkg/interval"
)

// Consumer defines all required method which producer will Call
type Consumer interface {
	Service(service *baker.Service) error
	Close(err error)
}

// Producer has a simple api to grab the consumer and will call proper methods
type Producer interface {
	Start(consumer Consumer)
}

// ProduceService is an implementation for Producer, it implements container.Consumer
// to consume containers and tries to fetch config from each container to produce service object
type ProduceService struct {
	configLoader ConfigLoader
	pingInterval time.Duration
	mux          sync.RWMutex
	table        map[string]*baker.Container
	containers   chan *baker.Container
}

var _ Producer = (*ProduceService)(nil)
var _ container.Consumer = (*ProduceService)(nil)
var _ interval.Ticker = (*ProduceService)(nil)

// Start will be called by higher implementation which give us
// the actual consumer for passing services
// NOTE: this method is a blocking call
func (p *ProduceService) Start(consumer Consumer) {
	ctx, cancel := context.WithCancel(context.Background())

	// interval.Run is a blocking call and needs to be run
	// inside a goroutine. In order to cancel it, a cancelable context is
	// used and once this method is terminated, defer function will cancel
	// the interval
	go interval.Run(ctx, p, p.pingInterval)

	defer func() {
		cancel()
		consumer.Close(nil)
	}()

	for container := range p.containers {
		var config *baker.Config
		var err error

		err = container.Err

		if err == nil {
			config, err = p.configLoader.Config(container.PingAddr)
		}

		service := &baker.Service{
			Container: container,
			Config:    config,
			Err:       err,
		}

		err = consumer.Service(service)
		if err != nil {
			// TODO log err?
		}
	}
}

// Container calls by container.Producer when a new container is available.
// Container will delete record in table if container is not active
func (p *ProduceService) Container(container *baker.Container) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	if !container.Active {
		delete(p.table, container.ID)
		return nil
	}

	c, ok := p.table[container.ID]
	if !ok {
		p.table[container.ID] = container
		return nil
	}

	c.Err = container.Err
	return nil
}

// Close calls by container.Producer when an error happens
func (p *ProduceService) Close(err error) {
	return
}

// Tick reads the internal table and push all the items in table into a channel
// NOTE: do not call this method, this will be called by interval.Run package.
func (p *ProduceService) Tick(ctx context.Context) error {
	p.mux.RLock()
	defer p.mux.RUnlock()

	for _, container := range p.table {
		select {
		case p.containers <- container:
		case <-ctx.Done():
			close(p.containers)
			break
		}
	}

	return nil
}

// New initialize ServiceProdicer object
func New(configLoader ConfigLoader, pingInterval time.Duration) *ProduceService {
	return &ProduceService{
		configLoader: configLoader,
		pingInterval: pingInterval,
	}
}
