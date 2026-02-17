package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWechatClient_SetBaseURLAndGetAccessToken(t *testing.T) {
	client := NewWechatClient("app", "secret")

	var tokenCalls int32
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/token", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&tokenCalls, 1)
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "abc", "expires_in": 7200})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client.SetBaseURL(server.URL)
	tok1, err := client.GetAccessToken()
	require.NoError(t, err)
	assert.Equal(t, "abc", tok1)

	tok2, err := client.GetAccessToken()
	require.NoError(t, err)
	assert.Equal(t, "abc", tok2)
	assert.Equal(t, int32(1), atomic.LoadInt32(&tokenCalls))
}

func TestWechatClient_GetAccessTokenErrorBranch(t *testing.T) {
	client := NewWechatClient("app", "secret")
	mux := http.NewServeMux()
	mux.HandleFunc("/cgi-bin/token", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"errcode": 40013, "errmsg": "invalid appid"})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client.SetBaseURL(server.URL)
	_, err := client.GetAccessToken()
	assert.Error(t, err)
}
