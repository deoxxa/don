package main

import (
	"sync"

	"fknsrs.biz/p/multierror"
	"github.com/Sirupsen/logrus"
)

type WorkerGroupFunc func() error

type WorkerGroup struct {
	m       sync.Mutex
	pending []WorkerGroupFunc
	errors  multierror.MultiError
}

func (w *WorkerGroup) Add(fn WorkerGroupFunc) {
	w.m.Lock()
	defer w.m.Unlock()

	w.pending = append(w.pending, fn)
}

func (w *WorkerGroup) Run(concurrency int) error {
	n := len(w.pending)

	logrus.WithFields(logrus.Fields{
		"tasks":       n,
		"concurrency": concurrency,
	}).Debug("worker group: starting")

	complete := make(chan int)

	for i := 0; i < concurrency; i++ {
		i := i
		go w.work(i, complete)
	}

	for i := 0; i < concurrency; i++ {
		logrus.WithField("worker_id", <-complete).Debug("worker reaped")
	}

	logrus.WithFields(logrus.Fields{
		"tasks":       n,
		"concurrency": concurrency,
		"errors":      w.errors.Len(),
	}).Debug("worker group: finished")

	if w.errors.Len() == 0 {
		return nil
	}

	for _, err := range w.errors {
		logrus.WithError(err).Debug("worker group: error")
	}

	return w.errors
}

func (w *WorkerGroup) pop() WorkerGroupFunc {
	w.m.Lock()
	defer w.m.Unlock()

	if len(w.pending) == 0 {
		return nil
	}

	fn := w.pending[0]

	w.pending = w.pending[1:]

	return fn
}

func (w *WorkerGroup) pushError(err error) {
	if err == nil {
		return
	}

	w.m.Lock()
	defer w.m.Unlock()

	logrus.WithError(err).Warn("worker group: received error")

	w.errors.Add(err)
}

func (w *WorkerGroup) work(id int, complete chan<- int) {
	defer func() { complete <- id }()

	logrus.WithField("worker_id", id).Debug("worker: starting")

	for {
		fn := w.pop()
		if fn == nil {
			break
		}

		logrus.WithField("worker_id", id).Debug("worker: starting task")
		w.pushError(fn())
		logrus.WithField("worker_id", id).Debug("worker: completed task")
	}

	logrus.WithField("worker_id", id).Debug("worker: completed")
}
