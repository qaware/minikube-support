package certmanager

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_certManager_getLatestVersion(t *testing.T) {
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
			m := &certManager{server: testServer.URL}

			got, err := m.getLatestVersion()
			if (err != nil) != tt.wantErr {
				t.Errorf("certManager.getLatestVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("certManager.getLatestVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
