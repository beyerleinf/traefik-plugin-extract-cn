package traefik_plugin_extract_cn_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	plugin "github.com/beyerleinf/traefik-plugin-extract-cn"
)

func TestCreateConfig(t *testing.T) {
	config := plugin.CreateConfig()
	if config == nil {
		t.Fatal("CreateConfig() returned nil")
	}

	if config.DestHeader != "" {
		t.Errorf("Expected empty DestHeader, got %q", config.DestHeader)
	}
}

func TestNewValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *plugin.Config
		expectError bool
	}{
		{
			name: "valid configuration",
			config: &plugin.Config{
				DestHeader: "X-Dest",
			},
			expectError: false,
		},
		{
			name:        "empty config",
			config:      &plugin.Config{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
			_, err := plugin.New(context.Background(), next, tt.config, "test")

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestPlugin(t *testing.T) {
	tests := []struct {
		name            string
		destHeader      string
		inputHeaders    map[string]string
		expectedHeaders map[string]string
	}{
		{
			name:       "basic extraction",
			destHeader: "X-Dest",
			inputHeaders: map[string]string{
				"X-Forwarded-Tls-Client-Cert-Info": "Subject%3D%22CN%3Dexample.com%22",
			},
			expectedHeaders: map[string]string{
				"X-Dest": "example.com",
			},
		},
		{
			name:       "extra cert info",
			destHeader: "X-Dest",
			inputHeaders: map[string]string{
				"X-Forwarded-Tls-Client-Cert-Info": "Subject%3D%22CN%3Dexample.com%2C%20OU%3DExample%20Org%22",
			},
			expectedHeaders: map[string]string{
				"X-Dest": "example.com",
			},
		},
		{
			name:       "missing common name",
			destHeader: "X-Dest",
			inputHeaders: map[string]string{
				"X-Forwarded-Tls-Client-Cert-Info": "Subject%3D%22OU%3DExample Org%22",
			},
			expectedHeaders: map[string]string{},
		},
		{
			name:       "missing cert info header",
			destHeader: "X-Dest",
			inputHeaders: map[string]string{
				"X-Some-Header": "Test",
			},
			expectedHeaders: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedHeaders http.Header
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				receivedHeaders = req.Header.Clone()
			})

			config := plugin.CreateConfig()
			config.DestHeader = tt.destHeader

			ctx := context.Background()

			handler, err := plugin.New(ctx, next, config, "plugin")
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
			if err != nil {
				t.Fatal(err)
			}

			for k, v := range tt.inputHeaders {
				req.Header.Set(k, v)
			}

			handler.ServeHTTP(recorder, req)

			for k, v := range tt.expectedHeaders {
				if got := receivedHeaders.Get(k); got != v {
					t.Errorf("Expected header %q to be %q, got %q", k, v, got)
				}
			}
		})
	}
}

func TestMiddlewareChain(t *testing.T) {
	handlerCalled := false
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		handlerCalled = true
		rw.WriteHeader(http.StatusOK)
	})

	config := plugin.CreateConfig()
	config.DestHeader = "X-Dest"

	handler, err := plugin.New(context.Background(), next, config, "test")
	if err != nil {
		t.Fatalf("Failed to create middleware: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	rw := httptest.NewRecorder()

	handler.ServeHTTP(rw, req)

	if !handlerCalled {
		t.Error("Next handler was not called")
	}

	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}
}
