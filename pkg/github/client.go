package github

//go:generate mockgen -destination=fake/mocks.go -package=fake -source=client.go

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client interface {
	GetLatestReleaseTag(org string, repository string) (string, error)
	DownloadReleaseAsset(org string, repository string, tag string, assetName string) (io.ReadCloser, error)
}

type client struct {
	apiHost string
	host    string
	client  *http.Client
}

func NewClient(apiToken string) Client {
	httpClient := &http.Client{Timeout: 10 * time.Second}
	if apiToken != "" {
		httpClient.Transport = newAuthTransport(apiToken)
	}
	return &client{
		apiHost: "https://api.github.com",
		host:    "https://github.com",
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

func (c *client) DownloadReleaseAsset(org string, repository string, tag string, assetName string) (io.ReadCloser, error) {
	resp, e := c.client.Get(fmt.Sprintf("%s/%s/%s/releases/download/%s/%s", c.host, org, repository, tag, assetName))
	if e != nil {
		return nil, fmt.Errorf("can not download asset %s for %s/%s of release %s: %s", assetName, org, repository, tag, e)
	}

	return resp.Body, nil
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
