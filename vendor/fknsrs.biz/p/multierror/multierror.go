package multierror

import (
	"strings"
)

type MultiError []error

func Wrap(err error) MultiError {
	return MultiError{err}
}

func (m *MultiError) Add(e error) {
	*m = append(*m, e)
}

func (m MultiError) Len() int {
	return len(m)
}

func (m MultiError) Error() string {
	s := make([]string, len(m))

	for i, v := range m {
		s[i] = v.Error()
	}

	return strings.Join(s, "\n")
}
