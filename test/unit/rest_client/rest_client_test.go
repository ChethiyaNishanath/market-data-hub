package restclient__test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	rest "github.com/ChethiyaNishanath/market-data-hub/internal/rest-client"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestDoRequest_JSONSuccess(t *testing.T) {
	type Resp struct {
		Message string `json:"message"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hello" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Resp{Message: "ok"})
	}))
	defer server.Close()

	client := rest.NewRestClient(server.URL, 5*time.Second)

	var out Resp
	err := client.Get(context.Background(), "/hello", rest.RequestOptions{}, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Message != "ok" {
		t.Errorf("expected ok, got %s", out.Message)
	}
}

func TestDoRequest_TextPlain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("hello world"))
	}))
	defer server.Close()

	client := rest.NewRestClient(server.URL, 5*time.Second)

	var out string
	err := client.Get(context.Background(), "/", rest.RequestOptions{}, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out != "hello world" {
		t.Errorf("expected 'hello world', got %s", out)
	}
}

func TestDoRequest_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := rest.NewRestClient(server.URL, 5*time.Second)

	var out string
	err := client.Get(context.Background(), "/", rest.RequestOptions{}, &out)
	if err == nil {
		t.Fatalf("expected error for 500 response")
	}
}

func TestDoRequest_UnknownContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/test")
		w.Write([]byte("binary-data"))
	}))
	defer server.Close()

	client := rest.NewRestClient(server.URL, 5*time.Second)

	var out string
	err := client.Get(context.Background(), "/", rest.RequestOptions{}, &out)
	if err == nil {
		t.Fatalf("expected error for unknown content type")
	}
}

func TestDoRequest_QueryHeadersBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Query().Get("q") != "binance" {
			t.Errorf("missing query param")
		}

		if r.Header.Get("X-Test") != "123" {
			t.Errorf("missing header")
		}

		body, _ := io.ReadAll(r.Body)
		if !bytes.Contains(body, []byte(`"value":"hello"`)) {
			t.Errorf("missing JSON body")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := rest.NewRestClient(server.URL, 5*time.Second)

	reqBody := map[string]any{"value": "hello"}
	opts := rest.RequestOptions{
		Headers: map[string]string{"X-Test": "123"},
		Query:   map[string]string{"q": "binance"},
		Body:    reqBody,
	}

	var out map[string]any
	err := client.Post(context.Background(), "/", opts, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out["ok"] != true {
		t.Errorf("expected ok=true")
	}
}

func TestDoWithRetry_SuccessFirstTry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := rest.DoWithRetry(server.Client(), req, 3, []int{500})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200")
	}
}

func TestDoWithRetry_RetryThenSuccess(t *testing.T) {
	attempt := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt < 3 {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := rest.DoWithRetry(server.Client(), req, 3, []int{500})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if attempt != 3 {
		t.Errorf("expected 3 attempts, got %d", attempt)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected 200")
	}
}

func TestDoWithRetry_NetworkError(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("network fail")
		}),
	}

	req, _ := http.NewRequest(http.MethodGet, "http://test", nil)
	_, err := rest.DoWithRetry(client, req, 2, []int{500})

	if err == nil {
		t.Fatalf("expected network error")
	}
}

func TestDoWithRetry_StopAtMaxRetry(t *testing.T) {
	attempt := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		w.WriteHeader(503)
	}))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)

	resp, err := rest.DoWithRetry(server.Client(), req, 2, []int{503})

	if err != nil {
		t.Fatalf("unexpected error")
	}

	if attempt != 3 {
		t.Errorf("expected 3 attempts, got %d", attempt)
	}

	if resp.StatusCode != 503 {
		t.Errorf("expected 503")
	}
}
