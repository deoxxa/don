package commonxml

import (
	"encoding/xml"
	"net/http"

	"github.com/pkg/errors"
)

func Fetch(u string, v interface{}) error {
	res, err := http.Get(u)
	if err != nil {
		return errors.Wrap(err, "FetchXML")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return errors.Errorf("FetchXML: invalid status code; expected 200 but got %d", res.StatusCode)
	}

	if err := xml.NewDecoder(res.Body).Decode(&v); err != nil {
		return errors.Wrap(err, "FetchXML")
	}

	return nil
}
