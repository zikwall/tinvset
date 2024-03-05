package invest

import (
	"github.com/tinkoff/invest-api-go-sdk/investgo"
)

type Client struct {
	client *investgo.Client
}

func (c *Client) Tinkoff() *investgo.Client {
	return c.client
}

func (c *Client) Drop() error {
	return c.client.Stop()
}

func (c *Client) DropMsg() string {
	return "close invest client"
}

func New(client *investgo.Client) *Client {
	return &Client{client: client}
}
