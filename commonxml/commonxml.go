package commonxml

import (
	"encoding/xml"
	"net/http"

	"github.com/pkg/errors"
)

func Fetch(u string, v interface{}) error {
	res, err := http.Get(u)
	if err != nil {
		return errors.Wrap(err, "commonxml.Fetch: couldn't make request")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return errors.Errorf("commonxml.Fetch: invalid status code; expected 200 but got %d", res.StatusCode)
	}

	if err := xml.NewDecoder(res.Body).Decode(&v); err != nil {
		return errors.Wrap(err, "commonxml.Fetch: couldn't decode xml")
	}

	return nil
}

func Parse(d []byte, v interface{}) error {
	return errors.Wrap(xml.Unmarshal(d, &v), "commonxml.Parse: couldn't decode xml")
}
