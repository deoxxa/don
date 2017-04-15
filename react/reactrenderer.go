package react

type Renderer interface {
	Render(code, inputURL, inputJSON string) (output string, err error)
}
