package daemon

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/maxmcd/steady/daemon/api"
)

var (
	ErrApplicationNotFound = errors.New("not found")
)

type Client struct {
	client *api.ClientWithResponses
}

func NewClient(server string) (*Client, error) {
	client, err := api.NewClientWithResponses(server, api.WithHTTPClient(
		&http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).Dial,
			},
		},
	))
	if err != nil {
		return nil, err
	}
	return &Client{client: client}, nil
}

func (c *Client) CreateApplication(name, script string) (*api.Application, error) {
	resp, err := c.client.CreateApplicationWithResponse(context.Background(), name, api.CreateApplicationJSONRequestBody{
		Script: script,
	})
	if err != nil {
		return nil, err
	}
	if resp.JSONDefault != nil {
		return nil, errors.New(resp.JSONDefault.Msg)
	}
	return resp.JSON201, nil
}

func (c *Client) DeleteApplication(name string) (*api.Application, error) {
	resp, err := c.client.DeleteApplicationWithResponse(context.Background(), name)
	if err != nil {
		return nil, err
	}
	if resp.JSONDefault != nil {
		return nil, errors.New(resp.JSONDefault.Msg)
	}

	return resp.JSON200, nil
}
