package sensonet

import (
	"fmt"
	"net/http"

	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/request"
	"golang.org/x/oauth2"
)

// Connection is the Sensonet connection
type Connection struct {
	client   *request.Helper
	systemId string
}

// NewConnection creates a new Sensonet device connection.
func NewConnection(log *util.Logger, ts oauth2.TokenSource) (*Connection, error) {
	client := request.NewHelper(log)
	client.Transport = &oauth2.Transport{
		Source: ts,
		Base:   client.Transport,
	}

	conn := &Connection{
		client: client,
	}

	var err error
	conn.systemId, err = conn.GetHomes()
	if err != nil {
		return conn, fmt.Errorf("could not get systemId. error: %w", err)
	}

	return conn, nil
}

// Returns the http header for http requests to sensonet
func (c *Connection) getSensonetHttpHeader() http.Header {
	return http.Header{
		"Accept-Language":           {"en-GB"},
		"Accept":                    {"application/json, text/plain, */*"},
		"x-app-identifier":          {"VAILLANT"},
		"x-client-locale":           {"en-GB"},
		"x-idm-identifier":          {"KEYCLOAK"},
		"ocp-apim-subscription-key": {"1e0a2f3511fb4c5bbb1c7f9fedd20b1c"},
	}
}

func (c *Connection) GetHomes() (string, error) {
	uri := API_URL_BASE + "/homes"
	req, _ := http.NewRequest("GET", uri, nil)
	req.Header = c.getSensonetHttpHeader()

	var res Homes
	if err := c.client.DoJSON(req, &res); err != nil {
		return "", fmt.Errorf("error getting homes: %w", err)
	}
	return res[0].SystemID, nil
}

func (c *Connection) GetSystem() (SystemStatus, error) {
	var res SystemStatus

	url := API_URL_BASE + fmt.Sprintf("/systems/%s/tli", c.systemId)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header = c.getSensonetHttpHeader()

	err := c.client.DoJSON(req, &res)

	return res, err
}
