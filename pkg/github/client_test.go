package github

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_client_GetLatestReleaseTag(t *testing.T) {
	tests := []struct {
		name            string
		responseStatus  int
		responseContent string
		want            string
		wantErr         bool
	}{
		{"ok", 200, "latest.json", "v0.8.1", false},
		{"no tagname", 200, "latest_no_tagname.json", "", true},
		{"invalid json", 200, "invalid.json", "", true},
		{"404", 404, "latest_no_tagname.json", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(tt.responseStatus)
				content, e := ioutil.ReadFile("fixtures/" + tt.responseContent)
				assert.NoError(t, e)
				_, e = res.Write(content)
				assert.NoError(t, e)
			}))
			defer func() { testServer.Close() }()

			c := &client{
				apiHost: testServer.URL,
				client:  testServer.Client(),
			}

			got, err := c.GetLatestReleaseTag("dummy", "test")
			if (err != nil) != tt.wantErr {
				t.Errorf("client.GetLatestReleaseTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("client.GetLatestReleaseTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_client_DownloadReleaseAsset(t *testing.T) {
	tests := []struct {
		name            string
		assetName       string
		responseContent string
		wantErr         bool
	}{
		{"ok", "abc", "true", false},
		{"not found", "404", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				if strings.HasSuffix(req.RequestURI, "/dummy/test/releases/download/v1.0/abc") {
					res.WriteHeader(200)
				} else {
					res.WriteHeader(404)
				}
				_, e := res.Write([]byte(tt.responseContent))
				assert.NoError(t, e)
			}))
			defer func() { testServer.Close() }()

			c := &client{
				apiHost: testServer.URL,
				host:    testServer.URL,
				client:  testServer.Client(),
			}
			_, err := c.DownloadReleaseAsset("dummy", "test", "v1.0", tt.assetName)
			if (err != nil) != tt.wantErr {
				t.Errorf("DownloadReleaseAsset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_client_Authentication(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{"ok", "test", 204},
		{"no token", "", 404},
		{"wrong token", "invalid token", 401},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				authHeader := req.Header.Get("Authorization")
				if authHeader == "token test" {
					res.WriteHeader(204)
				} else if authHeader == "" {
					res.WriteHeader(404)
				} else {
					res.WriteHeader(401)
				}
			}))
			defer func() { testServer.Close() }()

			c := &client{
				apiHost: testServer.URL,
				client:  testServer.Client(),
			}
			c.SetApiToken(tt.token)

			resp, err := c.client.Get(testServer.URL)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}
