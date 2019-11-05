package container

import (
	"github.com/alinz/baker"
)

// Consumer responsbible for consuimg from producer
type Consumer interface {
	Container(container *baker.Container) error
	Close(err error)
}

// Producer produces container and calls Consumer's Container
type Producer interface {
	Start(u Consumer)
}
