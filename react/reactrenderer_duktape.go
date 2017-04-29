package react

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/olebedev/go-duktape.v3"
)

type DuktapeRenderer struct {
	count int
	vms   chan *duktapeVM
}

func NewDuktapeRenderer(count int) *DuktapeRenderer {
	vms := make(chan *duktapeVM, count)
	for i := 0; i < count; i++ {
		vms <- nil
	}

	return &DuktapeRenderer{
		count: count,
		vms:   vms,
	}
}

func (r *DuktapeRenderer) Render(code, inputURL, inputJSON string) (string, error) {
	var html string

	if err := r.withVM(code, func(vm *duktape.Context) error {
		if err := vm.PevalString("module.exports"); err != nil {
			return errors.Wrap(err, "DuktapeRenderer.Render")
		}

		vm.PushString(inputURL)
		vm.PushString(inputJSON)
		if rc := vm.Pcall(2); rc != 0 {
			return fmt.Errorf("DuktapeRenderer.Render: error rendering app: %s", vm.SafeToString(-1))
		}
		html = vm.SafeToString(-1)
		vm.Pop()

		return nil
	}); err != nil {
		return "", errors.Wrap(err, "DuktapeRenderer.Render")
	}

	return html, nil
}

type duktapeVM struct {
	vm   *duktape.Context
	code string
}

func (r *DuktapeRenderer) withVM(code string, fn func(vm *duktape.Context) error) error {
	vm := <-r.vms
	defer func() {
		r.vms <- vm
	}()

	defer func() {
		if e := recover(); e != nil {
			vm = nil
		}
	}()

	if vm != nil && vm.code != code {
		vm.vm.DestroyHeap()
		vm = nil
	}

	if vm == nil {
		c := duktape.New()

		if err := c.PevalString(`
			console = {
				log: function() {},
				warn: function() {},
				error: function() {},
				debug: function() {},
			};

			module = { exports: null };
		`); err != nil {
			return errors.Wrap(err, "DuktapeRenderer.withVM")
		}

		if err := c.PevalString(code); err != nil {
			return errors.Wrap(err, "DuktapeRenderer.withVM")
		}

		vm = &duktapeVM{code: code, vm: c}
	}

	if err := fn(vm.vm); err != nil {
		vm.vm.DestroyHeap()
		vm = nil
		return errors.Wrap(err, "DuktapeRenderer.withVM")
	}

	return nil
}
