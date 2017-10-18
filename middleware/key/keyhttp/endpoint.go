package keyhttp

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/Comcast/webpa-common/middleware/key"
	gokithttp "github.com/go-kit/kit/transport/http"
	"github.com/jtacoma/uritemplates"
)

const KIDParameterName = "kid"

// EncodeTemplateKeyRequest discards the URL of an HTTP request in favor of a URI template expansion using
// the kid (key identifier).
//
// This encoder is only necessary if the URL can vary for each key.  For situations where the same URL
// can be used all the time, no encoder is necessary.
func EncodeTemplateKeyRequest(rawTemplate string) gokithttp.EncodeRequestFunc {
	if len(rawTemplate) == 0 {
		panic(errors.New("missing raw URI template"))
	}

	template, err := uritemplates.Parse(rawTemplate)
	if err != nil {
		panic(err)
	}

	return func(_ context.Context, request *http.Request, v interface{}) error {
		resolverRequest := v.(*key.ResolverRequest)
		expanded, err := template.Expand(map[string]interface{}{KIDParameterName: resolverRequest.KID})
		if err != nil {
			return err
		}

		request.URL, err = url.Parse(expanded)
		if err != nil {
			return nil
		}

		return nil
	}
}

// DecodeVerifyKeyResponse is the go-kit transport/http.DecodeResponseFunc that turns a PEM-encoded response
// into the appropriate key.
func DecodeVerifyKeyResponse(ctx context.Context, response *http.Response) (interface{}, error) {
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	rr := key.GetResolverRequest(ctx)
	if rr == nil {
		return nil, errors.New("No ResolverRequest in context")
	}

	return nil, nil
}
