package closer

import (
	"context"
	"fileuploader/pkg/app"
	"fileuploader/pkg/logger"
	"sync"
)

var cl *closer

func init() {
	cl = &closer{
		resourceClosers: make([]func() error, 0, 10),
	}
}

type closer struct {
	rw              sync.RWMutex
	resourceClosers []func() error
}

func NewCloser() app.App {
	return cl
}

func (c *closer) close() {
	c.rw.RLock()
	defer c.rw.RUnlock()
	for _, closeResource := range c.resourceClosers {
		if err := closeResource(); err != nil {
			logger.Errorf("failed to close: %v", err)
		}
	}
}

func (c *closer) Run(ctx context.Context) error {
	<-ctx.Done()
	c.close()
	return nil
}

func Add(c func() error) {
	cl.rw.Lock()
	defer cl.rw.Unlock()

	cl.resourceClosers = append(cl.resourceClosers, c)
}
