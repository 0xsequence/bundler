// bundler v0.1.0 0a5ac184a766eb430b897e8c6fb2c2f64d4c0df0
// --
// Code generated by webrpc-gen@v0.14.0-dev with golang generator. DO NOT EDIT.
//
// webrpc-gen -schema=rpc.ridl -target=golang -pkg=bundler -client -out=./clients/proto.gen.go
package bundler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// WebRPC description and code-gen version
func WebRPCVersion() string {
	return "v1"
}

// Schema version of your RIDL schema
func WebRPCSchemaVersion() string {
	return "v0.1.0"
}

// Schema hash generated from your RIDL schema
func WebRPCSchemaHash() string {
	return "0a5ac184a766eb430b897e8c6fb2c2f64d4c0df0"
}

//
// Types
//

type Version struct {
	WebrpcVersion string `json:"webrpcVersion"`
	SchemaVersion string `json:"schemaVersion"`
	SchemaHash    string `json:"schemaHash"`
	NodeVersion   string `json:"nodeVersion"`
}

type Status struct {
	HealthOK     bool      `json:"healthOK"`
	StartTime    time.Time `json:"startTime"`
	Uptime       uint64    `json:"uptime"`
	Ver          string    `json:"ver"`
	Branch       string    `json:"branch"`
	CommitHash   string    `json:"commitHash"`
	Multiaddrs   []string  `json:"multiaddrs"`
	PeerID       string    `json:"peerId"`
	Peers        []string  `json:"peers"`
	JoinedTopics []string  `json:"joinedTopics"`
}

type Bundler interface {
	Ping(ctx context.Context) (bool, error)
	Status(ctx context.Context) (*Status, error)
	Peers(ctx context.Context) ([]string, error)
}

var WebRPCServices = map[string][]string{
	"Bundler": {
		"Ping",
		"Status",
		"Peers",
	},
}

//
// Client
//

const BundlerPathPrefix = "/rpc/Bundler/"

type bundlerClient struct {
	client HTTPClient
	urls   [3]string
}

func NewBundlerClient(addr string, client HTTPClient) Bundler {
	prefix := urlBase(addr) + BundlerPathPrefix
	urls := [3]string{
		prefix + "Ping",
		prefix + "Status",
		prefix + "Peers",
	}
	return &bundlerClient{
		client: client,
		urls:   urls,
	}
}

func (c *bundlerClient) Ping(ctx context.Context) (bool, error) {
	out := struct {
		Ret0 bool `json:"status"`
	}{}

	err := doJSONRequest(ctx, c.client, c.urls[0], nil, &out)
	return out.Ret0, err
}

func (c *bundlerClient) Status(ctx context.Context) (*Status, error) {
	out := struct {
		Ret0 *Status `json:"status"`
	}{}

	err := doJSONRequest(ctx, c.client, c.urls[1], nil, &out)
	return out.Ret0, err
}

func (c *bundlerClient) Peers(ctx context.Context) ([]string, error) {
	out := struct {
		Ret0 []string `json:"peers"`
	}{}

	err := doJSONRequest(ctx, c.client, c.urls[2], nil, &out)
	return out.Ret0, err
}

// HTTPClient is the interface used by generated clients to send HTTP requests.
// It is fulfilled by *(net/http).Client, which is sufficient for most users.
// Users can provide their own implementation for special retry policies.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// urlBase helps ensure that addr specifies a scheme. If it is unparsable
// as a URL, it returns addr unchanged.
func urlBase(addr string) string {
	// If the addr specifies a scheme, use it. If not, default to
	// http. If url.Parse fails on it, return it unchanged.
	url, err := url.Parse(addr)
	if err != nil {
		return addr
	}
	if url.Scheme == "" {
		url.Scheme = "http"
	}
	return url.String()
}

// newRequest makes an http.Request from a client, adding common headers.
func newRequest(ctx context.Context, url string, reqBody io.Reader, contentType string) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", contentType)
	req.Header.Set("Content-Type", contentType)
	if headers, ok := HTTPRequestHeaders(ctx); ok {
		for k := range headers {
			for _, v := range headers[k] {
				req.Header.Add(k, v)
			}
		}
	}
	return req, nil
}

// doJSONRequest is common code to make a request to the remote service.
func doJSONRequest(ctx context.Context, client HTTPClient, url string, in, out interface{}) error {
	reqBody, err := json.Marshal(in)
	if err != nil {
		return ErrWebrpcRequestFailed.WithCause(fmt.Errorf("failed to marshal JSON body: %w", err))
	}
	if err = ctx.Err(); err != nil {
		return ErrWebrpcRequestFailed.WithCause(fmt.Errorf("aborted because context was done: %w", err))
	}

	req, err := newRequest(ctx, url, bytes.NewBuffer(reqBody), "application/json")
	if err != nil {
		return ErrWebrpcRequestFailed.WithCause(fmt.Errorf("could not build request: %w", err))
	}
	resp, err := client.Do(req)
	if err != nil {
		return ErrWebrpcRequestFailed.WithCause(err)
	}

	defer func() {
		cerr := resp.Body.Close()
		if err == nil && cerr != nil {
			err = ErrWebrpcRequestFailed.WithCause(fmt.Errorf("failed to close response body: %w", cerr))
		}
	}()

	if err = ctx.Err(); err != nil {
		return ErrWebrpcRequestFailed.WithCause(fmt.Errorf("aborted because context was done: %w", err))
	}

	if resp.StatusCode != 200 {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return ErrWebrpcBadResponse.WithCause(fmt.Errorf("failed to read server error response body: %w", err))
		}

		var rpcErr WebRPCError
		if err := json.Unmarshal(respBody, &rpcErr); err != nil {
			return ErrWebrpcBadResponse.WithCause(fmt.Errorf("failed to unmarshal server error: %w", err))
		}
		if rpcErr.Cause != "" {
			rpcErr.cause = errors.New(rpcErr.Cause)
		}
		return rpcErr
	}

	if out != nil {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return ErrWebrpcBadResponse.WithCause(fmt.Errorf("failed to read response body: %w", err))
		}

		err = json.Unmarshal(respBody, &out)
		if err != nil {
			return ErrWebrpcBadResponse.WithCause(fmt.Errorf("failed to unmarshal JSON response body: %w", err))
		}
	}

	return nil
}

func WithHTTPRequestHeaders(ctx context.Context, h http.Header) (context.Context, error) {
	if _, ok := h["Accept"]; ok {
		return nil, errors.New("provided header cannot set Accept")
	}
	if _, ok := h["Content-Type"]; ok {
		return nil, errors.New("provided header cannot set Content-Type")
	}

	copied := make(http.Header, len(h))
	for k, vv := range h {
		if vv == nil {
			copied[k] = nil
			continue
		}
		copied[k] = make([]string, len(vv))
		copy(copied[k], vv)
	}

	return context.WithValue(ctx, HTTPClientRequestHeadersCtxKey, copied), nil
}

func HTTPRequestHeaders(ctx context.Context) (http.Header, bool) {
	h, ok := ctx.Value(HTTPClientRequestHeadersCtxKey).(http.Header)
	return h, ok
}

//
// Helpers
//

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "webrpc context value " + k.name
}

var (
	HTTPClientRequestHeadersCtxKey = &contextKey{"HTTPClientRequestHeaders"}
	HTTPRequestCtxKey              = &contextKey{"HTTPRequest"}

	ServiceNameCtxKey = &contextKey{"ServiceName"}

	MethodNameCtxKey = &contextKey{"MethodName"}
)

func ServiceNameFromContext(ctx context.Context) string {
	service, _ := ctx.Value(ServiceNameCtxKey).(string)
	return service
}

func MethodNameFromContext(ctx context.Context) string {
	method, _ := ctx.Value(MethodNameCtxKey).(string)
	return method
}

func RequestFromContext(ctx context.Context) *http.Request {
	r, _ := ctx.Value(HTTPRequestCtxKey).(*http.Request)
	return r
}

//
// Errors
//

type WebRPCError struct {
	Name       string `json:"error"`
	Code       int    `json:"code"`
	Message    string `json:"msg"`
	Cause      string `json:"cause,omitempty"`
	HTTPStatus int    `json:"status"`
	cause      error
}

var _ error = WebRPCError{}

func (e WebRPCError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s %d: %s: %v", e.Name, e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s %d: %s", e.Name, e.Code, e.Message)
}

func (e WebRPCError) Is(target error) bool {
	if rpcErr, ok := target.(WebRPCError); ok {
		return rpcErr.Code == e.Code
	}
	return errors.Is(e.cause, target)
}

func (e WebRPCError) Unwrap() error {
	return e.cause
}

func (e WebRPCError) WithCause(cause error) WebRPCError {
	err := e
	err.cause = cause
	err.Cause = cause.Error()
	return err
}

// Deprecated: Use .WithCause() method on WebRPCError.
func ErrorWithCause(rpcErr WebRPCError, cause error) WebRPCError {
	return rpcErr.WithCause(cause)
}

// Webrpc errors
var (
	ErrWebrpcEndpoint           = WebRPCError{Code: 0, Name: "WebrpcEndpoint", Message: "endpoint error", HTTPStatus: 400}
	ErrWebrpcRequestFailed      = WebRPCError{Code: -1, Name: "WebrpcRequestFailed", Message: "request failed", HTTPStatus: 400}
	ErrWebrpcBadRoute           = WebRPCError{Code: -2, Name: "WebrpcBadRoute", Message: "bad route", HTTPStatus: 404}
	ErrWebrpcBadMethod          = WebRPCError{Code: -3, Name: "WebrpcBadMethod", Message: "bad method", HTTPStatus: 405}
	ErrWebrpcBadRequest         = WebRPCError{Code: -4, Name: "WebrpcBadRequest", Message: "bad request", HTTPStatus: 400}
	ErrWebrpcBadResponse        = WebRPCError{Code: -5, Name: "WebrpcBadResponse", Message: "bad response", HTTPStatus: 500}
	ErrWebrpcServerPanic        = WebRPCError{Code: -6, Name: "WebrpcServerPanic", Message: "server panic", HTTPStatus: 500}
	ErrWebrpcInternalError      = WebRPCError{Code: -7, Name: "WebrpcInternalError", Message: "internal error", HTTPStatus: 500}
	ErrWebrpcClientDisconnected = WebRPCError{Code: -8, Name: "WebrpcClientDisconnected", Message: "client disconnected", HTTPStatus: 400}
	ErrWebrpcStreamLost         = WebRPCError{Code: -9, Name: "WebrpcStreamLost", Message: "stream lost", HTTPStatus: 400}
	ErrWebrpcStreamFinished     = WebRPCError{Code: -10, Name: "WebrpcStreamFinished", Message: "stream finished", HTTPStatus: 200}
)
