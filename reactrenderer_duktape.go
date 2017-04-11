package main

import (
	"fmt"

	"github.com/olebedev/go-duktape"
)

type ReactRendererDuktape struct {
	count int
	vms   chan *reactRendererDuktapeVM
}

func NewReactRendererDuktape(count int) *ReactRendererDuktape {
	vms := make(chan *reactRendererDuktapeVM, count)
	for i := 0; i < count; i++ {
		vms <- nil
	}

	return &ReactRendererDuktape{
		count: count,
		vms:   vms,
	}
}

func (r *ReactRendererDuktape) Render(code, inputURL, inputJSON string) (string, error) {
	var html string

	if err := r.withVM(code, func(vm *duktape.Context) error {
		if err := vm.PevalString("module.exports"); err != nil {
			return err
		}

		vm.PushString(inputURL)
		vm.PushString(inputJSON)
		if rc := vm.Pcall(2); rc != 0 {
			return fmt.Errorf("ReactRendererDuktape.Render: error rendering app: %s", vm.SafeToString(-1))
		}
		html = vm.SafeToString(-1)
		vm.Pop()

		return nil
	}); err != nil {
		return "", err
	}

	return html, nil
}

type reactRendererDuktapeVM struct {
	vm   *duktape.Context
	code string
}

func (r *ReactRendererDuktape) withVM(code string, fn func(vm *duktape.Context) error) error {
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
			return err
		}

		if err := c.PevalString(code); err != nil {
			return err
		}

		vm = &reactRendererDuktapeVM{code: code, vm: c}
	}

	if err := fn(vm.vm); err != nil {
		vm.vm.DestroyHeap()
		vm = nil
		return err
	}

	return nil
}
