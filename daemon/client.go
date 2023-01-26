package daemon

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/maxmcd/steady/daemon/api"
)

var (
	ErrApplicationNotFound = errors.New("not found")
)

func errUnexpectedResponse(body []byte, code int) error {
	return fmt.Errorf("Unexpected response from server %d: %s", code, string(body))
}

type Client struct {
	client *api.ClientWithResponses
}

func NewClient(server string, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).Dial,
			},
		}
	}
	client, err := api.NewClientWithResponses(server,
		api.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}
	return &Client{client: client}, nil
}

func (c *Client) CreateApplication(ctx context.Context, name, script string) (*api.Application, error) {
	resp, err := c.client.CreateApplicationWithResponse(ctx, name, api.CreateApplicationJSONRequestBody{
		Script: script,
	})
	if err != nil {
		return nil, err
	}
	if resp.JSONDefault != nil {
		return nil, errors.New(resp.JSONDefault.Msg)
	}
	if resp.JSON201 != nil {
		return resp.JSON201, nil
	}
	return nil, errUnexpectedResponse(resp.Body, resp.StatusCode())
}

func (c *Client) DeleteApplication(ctx context.Context, name string) (*api.Application, error) {
	resp, err := c.client.DeleteApplicationWithResponse(ctx, name)
	if err != nil {
		return nil, err
	}
	if resp.JSONDefault != nil {
		return nil, errors.New(resp.JSONDefault.Msg)
	}

	if resp.JSON200 != nil {
		return resp.JSON200, nil
	}
	return nil, errUnexpectedResponse(resp.Body, resp.StatusCode())
}

func (c *Client) GetApplication(ctx context.Context, name string) (*api.Application, error) {
	resp, err := c.client.GetApplicationWithResponse(ctx, name)
	if err != nil {
		return nil, err
	}

	if resp.JSONDefault != nil {
		return nil, errors.New(resp.JSONDefault.Msg)
	}
	if resp.JSON200 != nil {
		return resp.JSON200, nil
	}
	return nil, errUnexpectedResponse(resp.Body, resp.StatusCode())
}
