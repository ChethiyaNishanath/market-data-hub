package restclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type RestClient struct {
	BaseURL    string
	ApiKey     string
	HTTPClient *http.Client
}

func NewRestClient(baseUrl string, timeout time.Duration) *RestClient {

	return &RestClient{
		BaseURL: baseUrl,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func NewRestClientAuth(baseUrl string, apiKey string, timeout time.Duration) *RestClient {

	return &RestClient{
		BaseURL: baseUrl,
		ApiKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 1 * time.Minute,
		},
	}
}

type RequestOptions struct {
	Headers map[string]string
	Query   map[string]string
	Body    any
}

func (c *RestClient) doRequest(
	ctx context.Context,
	method, path string,
	opts RequestOptions,
	respBody any,
) error {

	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return err
	}

	q := u.Query()
	for k, v := range opts.Query {
		q.Add(k, v)
	}
	u.RawQuery = q.Encode()

	var body io.Reader
	if opts.Body != nil {
		jsonBytes, err := json.Marshal(opts.Body)
		if err != nil {
			return err
		}
		body = bytes.NewBuffer(jsonBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("non-200 response: %d -> %s", res.StatusCode, string(resBytes))
	}

	contentType := res.Header.Get("Content-Type")

	switch {
	case strings.Contains(contentType, "application/json"):
		return json.Unmarshal(resBytes, respBody)

	case strings.Contains(contentType, "text/plain"),
		strings.Contains(contentType, "text/html"),
		strings.Contains(contentType, "application/xml"),
		strings.Contains(contentType, "text/xml"):

		switch v := respBody.(type) {
		case *string:
			*v = string(resBytes)
		case *[]byte:
			*v = resBytes
		default:
			return fmt.Errorf("cannot decode non-JSON into %T", respBody)
		}

		return nil

	default:
		if b, ok := respBody.(*[]byte); ok {
			*b = resBytes
			return nil
		}
		return fmt.Errorf("unknown content-type '%s', cannot decode into %T", contentType, respBody)
	}
}

func DoWithRetry(
	client *http.Client,
	req *http.Request,
	retry int,
	retryStatusCodes []int,
) (*http.Response, error) {

	var resp *http.Response
	var err error

	retryMap := make(map[int]bool)
	for _, code := range retryStatusCodes {
		retryMap[code] = true
	}

	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed reading request body: %w", err)
		}
	}

	for i := 0; i <= retry; i++ {

		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		resp, err = client.Do(req)

		if err != nil {
			if i == retry {
				return nil, err
			}
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		if !retryMap[resp.StatusCode] {
			return resp, nil
		}

		if i == retry {
			return resp, nil
		}

		io.Copy(io.Discard, resp.Body)
		defer resp.Body.Close()

		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return resp, err
}

func (c *RestClient) Get(ctx context.Context, path string, opts RequestOptions, resp any) error {
	return c.doRequest(ctx, http.MethodGet, path, opts, resp)
}

func (c *RestClient) Post(ctx context.Context, path string, opts RequestOptions, resp any) error {
	return c.doRequest(ctx, http.MethodPost, path, opts, resp)
}

func (c *RestClient) Put(ctx context.Context, path string, opts RequestOptions, resp any) error {
	return c.doRequest(ctx, http.MethodPut, path, opts, resp)
}

func (c *RestClient) Patch(ctx context.Context, path string, opts RequestOptions, resp any) error {
	return c.doRequest(ctx, http.MethodPatch, path, opts, resp)
}

func (c *RestClient) Delete(ctx context.Context, path string, opts RequestOptions, resp any) error {
	return c.doRequest(ctx, http.MethodDelete, path, opts, resp)
}
