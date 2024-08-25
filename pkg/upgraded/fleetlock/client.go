package fleetlock

import (
	"fmt"
	"net/http"

	"github.com/heathcliff26/fleetlock/pkg/server/client"
)

type FleetlockClient struct {
	url   string
	group string
	appID string
}

// Create a new client for fleetlock
func NewClient(url, group string) (*FleetlockClient, error) {
	if url == "" || group == "" {
		return nil, fmt.Errorf("at least one of the required parameters is empty")
	}

	appID, err := GetZincateAppID()
	if err != nil {
		return nil, fmt.Errorf("failed to create zincati app id: %v", err)
	}

	return &FleetlockClient{
		url:   TrimTrailingSlash(url),
		group: group,
		appID: appID,
	}, nil
}

// Aquire a lock for this machine
func (c *FleetlockClient) Lock() error {
	ok, res, err := c.doRequest("/v1/pre-reboot")
	if err != nil {
		return err
	} else if ok {
		return nil
	}
	return fmt.Errorf("failed to aquire lock kind=\"%s\" reason=\"%s\"", res.Kind, res.Value)
}

// Release the hold lock
func (c *FleetlockClient) Release() error {
	ok, res, err := c.doRequest("/v1/steady-state")
	if err != nil {
		return err
	} else if ok {
		return nil
	}
	return fmt.Errorf("failed to release lock kind=\"%s\" reason=\"%s\"", res.Kind, res.Value)
}

func (c *FleetlockClient) doRequest(path string) (bool, client.FleetLockResponse, error) {
	body, err := client.PrepareRequest(c.group, c.appID)
	if err != nil {
		return false, client.FleetLockResponse{}, fmt.Errorf("failed to prepare request body: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, c.url+path, body)
	if err != nil {
		return false, client.FleetLockResponse{}, fmt.Errorf("failed to create http post request: %v", err)
	}
	req.Header.Set("fleet-lock-protocol", "true")
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, client.FleetLockResponse{}, fmt.Errorf("failed to send request to server: %v", err)
	}

	resBody, err := client.ParseResponse(res.Body)
	if err != nil {
		return false, client.FleetLockResponse{}, fmt.Errorf("failed to prepare response body: %v", err)
	}

	return res.StatusCode == http.StatusOK, resBody, nil
}
