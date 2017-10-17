package keyhttp

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	gokithttp "github.com/go-kit/kit/transport/http"
	"github.com/jtacoma/uritemplates"
)

const DefaultKIDParameterName = "kid"

// SimpleEncodeKeyRequest does nothing to the request except apply custom headers.  This
// encoder function is appropriate when the full URL of the key can be configured ahead of time.
func SimpleEncodeKeyRequest(custom http.Header) gokithttp.EncodeRequestFunc {
	return func(_ context.Context, request *http.Request, _ interface{}) error {
		for name, values := range custom {
			for _, value := range values {
				request.Header.Add(name, value)
			}
		}

		return nil
	}
}

// TemplateEncodeKeyRequest discards the URL of an HTTP request in favor of a URI template expansion using
// the kid (key identifier).
func TemplateEncodeKeyRequest(rawTemplate, parameter string, custom http.Header) gokithttp.EncodeRequestFunc {
	if len(rawTemplate) == 0 {
		panic(errors.New("missing raw URI template"))
	}

	template, err := uritemplates.Parse(rawTemplate)
	if err != nil {
		panic(err)
	}

	if len(parameter) == 0 {
		parameter = DefaultKIDParameterName
	}

	return func(_ context.Context, request *http.Request, kid interface{}) error {
		expanded, err := template.Expand(map[string]interface{}{parameter: kid})
		if err != nil {
			return err
		}

		request.URL, err = url.Parse(expanded)
		if err != nil {
			return nil
		}

		for name, values := range custom {
			for _, value := range values {
				request.Header.Add(name, value)
			}
		}

		return nil
	}
}

// DecodeKeyResponse is the go-kit transport/http.DecodeResponseFunc that turns a PEM-encoded response
// into the appropriate key.
func DecodeKeyResponse(_ context.Context, response *http.Response) (interface{}, error) {
}
