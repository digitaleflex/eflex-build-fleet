package registry

import "context"

type Client struct {
	Endpoint string
}

func New(endpoint string) *Client { return &Client{Endpoint: endpoint} }

func (c *Client) Push(ctx context.Context, image, digest string) error {
	return ErrNotImplemented
}

func (c *Client) Pull(ctx context.Context, image string) (string, error) {
	return "", ErrNotImplemented
}

type ErrNotImplemented struct{}

func (ErrNotImplemented) Error() string { return "registry client not implemented" }