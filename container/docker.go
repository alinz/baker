package container

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/alinz/baker"
	"github.com/alinz/baker/pkg/endpoint"
	"github.com/alinz/baker/pkg/logger"
)

type event struct {
	id     string
	active bool
}

// Docker is an implementation of Docker's container producer
type Docker struct {
	client *http.Client
	addr   endpoint.HTTPAddr
	err    error
}

var _ Producer = (*Docker)(nil)

// Start starts the process of consuming Docker events and produces container
// object.
// NOTE: this method is blocking call
func (d *Docker) Start(consumer Consumer) {
	events := make(chan *event, 10)
	containers := make(chan *baker.Container, 10)

	// Run this in background, this make sures
	// that events being pushed, can be transformed and
	// pushed into containers
	go d.eventsToContainers(events, containers)

	// Need to process already running containers
	// need to wait until all already running containers
	// being transformed into events
	d.processRunningContainers(events)
	if d.err != nil {
		consumer.Close(d.err)
		return
	}

	// we are now ready to listen to process
	// live events
	go d.processLiveEvents(events)

	// simply go through each container object and call update
	for container := range containers {
		consumer.Container(container)
	}

	consumer.Close(nil)
}

// processRunningContainers will be called at first to make sure running containers
// also get updated. This makes sure the already running containers get registered
func (d *Docker) processRunningContainers(events chan<- *event) {
	addr := d.addr.WithPath("/containers/json")
	resp, err := d.client.Get(addr.String())
	if err != nil {
		d.err = err
		close(events)
		return
	}

	payload := []struct {
		ID    string `json:"Id"`
		State string `json:"State"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		d.err = err
		close(events)
		return
	}

	for _, data := range payload {
		events <- &event{
			id:     data.ID,
			active: data.State == "running",
		}
	}
}

// processLiveEvents process all incoming events, in case of any error,
// error value will be set to internal err property and events' channel will
// be closed.
// NOTE: This method is blocking and will terminate if an error occurs
func (d *Docker) processLiveEvents(events chan<- *event) {
	addr := d.addr.WithPath("/events")
	resp, err := d.client.Get(addr.String())
	if err != nil {
		d.err = err
		close(events)
		return
	}

	eventsDecoder := json.NewDecoder(resp.Body)

	for {
		payload := struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}{}

		err = eventsDecoder.Decode(&payload)
		if err != nil {
			d.err = err
			close(events)
			return
		}

		if payload.Status != "die" && payload.Status != "start" {
			continue
		}

		events <- &event{
			id:     payload.ID,
			active: payload.Status == "start",
		}
	}
}

// eventsToContainers processes all the events and tries to push containers objects to containers channel.
// in case of any errors, an err value will be set on container object and will be pushed to containers' channel
// TODO: this method is blocking
func (d *Docker) eventsToContainers(events <-chan *event, containers chan<- *baker.Container) {
	defer close(containers)

	for event := range events {
		// if event is not active, it means container has no longer
		// exists and there is no need to fetch its data
		if !event.active {
			logger.Debug("container %s is not active", event.id)
			containers <- &baker.Container{
				ID: event.id,
			}
			continue
		}

		logger.Debug("container %s is active", event.id)

		addr := d.addr.WithPath("/containers/" + event.id + "/json")
		resp, err := d.client.Get(addr.String())
		if err != nil {
			logger.Error("fetch container '%s' info because %s", event.id, err)

			containers <- &baker.Container{
				ID:  event.id,
				Err: err,
			}
			continue
		}

		payload := &struct {
			ID string `json:"Id"`

			Config *struct {
				Labels *struct {
					Network     string `json:"baker.network"`
					ServicePort string `json:"baker.service.port"`
					ServicePing string `json:"baker.service.ping"`
					ServiceSSL  string `json:"baker.service.ssl"`
				} `json:"Labels"`
			} `json:"Config"`

			NetworkSettings struct {
				Networks map[string]struct {
					IPAddress string `json:"IPAddress"`
				} `json:"Networks"`
			} `json:"NetworkSettings"`
		}{}

		err = json.NewDecoder(resp.Body).Decode(payload)
		if err != nil {
			logger.Error("failed to parse container %s", event.id)

			containers <- &baker.Container{
				ID:  event.id,
				Err: err,
			}
			continue
		}

		network, ok := payload.NetworkSettings.Networks[payload.Config.Labels.Network]
		if !ok {
			logger.Debug("network %s not exisits in label for container %s", payload.Config.Labels.Network, event.id)

			containers <- &baker.Container{
				ID:  event.id,
				Err: fmt.Errorf("network '%s' not exists in labels", payload.Config.Labels.Network),
			}
			continue
		}

		port, err := strconv.ParseInt(payload.Config.Labels.ServicePort, 10, 32)
		if err != nil {
			logger.Debug("failed to parse port for container '%s' because %s", event.id, err)

			containers <- &baker.Container{
				ID:  event.id,
				Err: fmt.Errorf("failed to parse port for container '%s' because %s", event.id, err),
			}
			continue
		}

		serviceAddr := endpoint.NewAddr(network.IPAddress, int(port), payload.Config.Labels.ServiceSSL == "true")

		containers <- &baker.Container{
			ID:       event.id,
			Active:   true,
			Addr:     serviceAddr,
			PingAddr: endpoint.NewHTTPAddr(serviceAddr, payload.Config.Labels.ServicePing),
		}
	}
}

// DefaultClient is a default client which uses unix protocol
var DefaultClient = &http.Client{
	Transport: &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", "/var/run/docker.sock")
		},
	},
}

// DefaultAddr is a default docker host and port which communicating with Docker deamon
const DefaultAddr = "http://localhost"

// NewDocker creates a new docker watcher
func NewDocker(client *http.Client, addr string) *Docker {
	return &Docker{
		client: client,
		addr:   endpoint.ParseHTTPAddr(addr),
	}
}
