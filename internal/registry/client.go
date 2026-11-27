package registry

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type Client struct {
	Endpoint string
}

func New(endpoint string) *Client { return &Client{Endpoint: endpoint} }

func (c *Client) Push(ctx context.Context, image, digest string) error {
	cmd := exec.CommandContext(ctx, "skopeo", "copy", "docker://"+image, "docker://"+c.Endpoint+"/"+image)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("skopeo push failed: %w: %s", err, stderr.String())
	}
	return nil
}

func (c *Client) Pull(ctx context.Context, image string) (string, error) {
	cmd := exec.CommandContext(ctx, "skopeo", "inspect", "docker://"+c.Endpoint+"/"+image)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("skopeo pull failed: %w", err)
	}
	return string(out), nil
}