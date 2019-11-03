package driver

import "github.com/alinz/bake/data"

type Updater interface {
	Update(container *data.Container) error
	Close(err error)
}

type Watcher interface {
	Start(u Updater)
}
