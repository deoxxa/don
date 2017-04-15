package workergroup

import (
	"sync"

	"fknsrs.biz/p/multierror"
)

type WorkerFunc func() error

type Group struct {
	m       sync.Mutex
	pending []WorkerFunc
	errors  multierror.MultiError
}

func (g *Group) Add(fn WorkerFunc) {
	g.m.Lock()
	defer g.m.Unlock()

	g.pending = append(g.pending, fn)
}

func (g *Group) Run(concurrency int) error {
	complete := make(chan int, len(g.pending))

	for i := 0; i < concurrency; i++ {
		i := i
		go g.work(i, complete)
	}

	for i := 0; i < concurrency; i++ {
		<-complete
	}

	if g.errors.Len() == 0 {
		return nil
	}

	return g.errors
}

func (g *Group) pop() WorkerFunc {
	g.m.Lock()
	defer g.m.Unlock()

	if len(g.pending) == 0 {
		return nil
	}

	fn := g.pending[0]

	g.pending = g.pending[1:]

	return fn
}

func (g *Group) pushError(err error) {
	if err == nil {
		return
	}

	g.m.Lock()
	defer g.m.Unlock()

	g.errors.Add(err)
}

func (g *Group) work(id int, complete chan<- int) {
	defer func() { complete <- id }()

	for {
		fn := g.pop()
		if fn == nil {
			break
		}

		g.pushError(fn())
	}
}
