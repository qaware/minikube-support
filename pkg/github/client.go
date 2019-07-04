package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client interface {
	GetLatestReleaseTag(org string, repository string) (string, error)
	DownloadReleaseAsset(org string, repository string, tag string, assetName string) ([]byte, error)
}

type client struct {
	apiHost string
	client  *http.Client
}

func NewClient(apiToken string) Client {
	httpClient := &http.Client{Timeout: 2 * time.Second}
	httpClient.Transport = newAuthTransport(apiToken)
	return &client{
		apiHost: "https://api.github.com",
		client:  httpClient,
	}
}

func (c *client) GetLatestReleaseTag(org string, repository string) (string, error) {
	resp, e := c.client.Get(fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.apiHost, org, repository))
	if e != nil {
		return "", fmt.Errorf("can not get latest release tag for %s/%s: %s", org, repository, e)
	}

	data := make(map[string]interface{})
	decoder := json.NewDecoder(resp.Body)

	err := decoder.Decode(&data)
	if err != nil {
		return "", err
	}

	version, ok := data["tag_name"]
	if !ok {
		return "", fmt.Errorf("tag field not found")
	}
	v, ok := version.(string)
	if !ok {
		return "", fmt.Errorf("tag is not a string")
	}
	return v, nil
}

func (c *client) DownloadReleaseAsset(org string, repository string, tag string, assetName string) ([]byte, error) {
	panic("implement me")
}

type transport struct {
	underlyingTransport http.RoundTripper
	apiToken            string
}

func newAuthTransport(apiToken string) *transport {
	return &transport{
		apiToken:            apiToken,
		underlyingTransport: http.DefaultTransport,
	}
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.apiToken != "" {
		req.Header.Add("Authorization", "Bearer "+t.apiToken)
	}
	return t.underlyingTransport.RoundTrip(req)
}
