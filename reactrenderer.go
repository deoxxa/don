package main

type ReactRenderer interface {
	Render(code, inputURL, inputJSON string) (output string, err error)
}
