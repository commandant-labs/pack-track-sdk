package transport

import "context"

// Request represents a payload to be sent to the ingest service.
type Request struct {
	Payload     []byte
	ContentType string
	// Path is appended to the endpoint, e.g., "/v1/logs".
	Path string
}

// Response is the transport response.
type Response struct {
	Status int
	Body   []byte
}

// Transport abstracts network transport for the SDK.
type Transport interface {
	Send(ctx context.Context, req Request) (Response, error)
	Close(ctx context.Context) error
}
