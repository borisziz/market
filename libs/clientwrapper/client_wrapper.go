package clientwrapper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type Wrapper[Req any, Res any] struct {
	url      string
	request  Req
	response Res
}

func New[Req any, Res any](url string, request Req, response Res) *Wrapper[Req, Res] {
	return &Wrapper[Req, Res]{
		url:      url,
		request:  request,
		response: response,
	}
}

func (w *Wrapper[Req, Res]) ClientRequest(ctx context.Context) (Res, error) {
	rawJSON, err := json.Marshal(w.request)
	if err != nil {
		return w.response, errors.Wrap(err, "marshaling json")
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewBuffer(rawJSON))
	if err != nil {
		return w.response, errors.Wrap(err, "creating http request")
	}
	httpResponse, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return w.response, errors.Wrap(err, "calling http")
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return w.response, fmt.Errorf("wrong status code: %d", httpResponse.StatusCode)
	}
	err = json.NewDecoder(httpResponse.Body).Decode(&w.response)
	if err != nil {
		return w.response, errors.Wrap(err, "decoding json")
	}
	return w.response, nil
}
