package hostmeta

import (
	"encoding/xml"
	"net/http"

	"github.com/pkg/errors"

	"fknsrs.biz/p/don/commonxml"
)

func Fetch(domain string) (*Response, error) {
	req, err := http.NewRequest("GET", "https://"+domain+"/.well-known/host-meta", nil)
	if err != nil {
		return nil, errors.Wrap(err, "hostmeta.Fetch")
	}
	req.Header.Set("accept", "application/xrd+xml")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "hostmeta.Fetch")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.Errorf("hostmeta.Fetch: invalid status code; expected 200 but got %d", res.StatusCode)
	}

	var v Response
	if err := xml.NewDecoder(res.Body).Decode(&v); err != nil {
		return nil, errors.Wrap(err, "hostmeta.Fetch")
	}

	return &v, nil
}

type Response struct {
	XMLName xml.Name `xml:"http://docs.oasis-open.org/ns/xri/xrd-1.0 XRD"`

	commonxml.HasLinks
	commonxml.HasProperties
}
