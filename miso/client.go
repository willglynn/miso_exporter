package miso

import (
	"net/http"
)

type Client struct {
	client http.Client
}

func NewClient(transport http.RoundTripper) *Client {
	if transport == nil {
		transport = http.DefaultTransport
	}

	client := http.Client{
		Transport: transport,
	}

	return &Client{client: client}
}
