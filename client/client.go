package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/aprice/freenote/users"
)

// Client interfaces with the REST API.
type Client struct {
	*http.Client

	Username string
	Password string
	Host     string

	User users.User
}

// New constructs a new Client.
func New(user, pass, host string) (*Client, error) {
	if user == "" || pass == "" || host == "" {
		return nil, errors.New("username, password, and host are required")
	}
	c := &Client{
		Client:   new(http.Client),
		Username: user,
		Password: pass,
		Host:     host,
	}
	err := c.Connect()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Connect to the server and establish a session.
func (c *Client) Connect() error {
	return c.Call("GET", fmt.Sprintf("/users/%s", c.Username), "", nil, &c.User)
}

// Get an arbitrary object or collection from a route and unmarshal it.
func (c *Client) Get(route string, result interface{}) error {
	return c.Call("GET", route, "", nil, result)
}

// Send an arbitrary payload to a route and unmarshal the response.
func (c *Client) Send(method, route string, payload, result interface{}) error {
	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}
	return c.Call(method, route, "application/json", body, result)
}

// Call a route with all parameters supplied by the caller. Basic HTTP executor.
func (c *Client) Call(method, route, ctype string, payload []byte, result interface{}) error {
	u, err := url.Parse(c.Host)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, route)
	req, err := http.NewRequest(method, u.String(), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", ctype)

	res, err := c.Do(req)
	defer CleanupResponse(res)
	if err != nil {
		return err
	}
	if res.StatusCode >= 300 {
		return fmt.Errorf("request returned status %d: %s", res.StatusCode, res.Status)
	}
	if result == nil {
		return nil
	}
	return json.NewDecoder(res.Body).Decode(result)
}
