package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware(t *testing.T) {
	tests := []struct {
		name               string
		acceptEncoding     string
		contentType        string
		contentEncoding    string
		body               string
		expectedGzip       bool
		expectedStatusCode int
	}{
		{
			name:               "no gzip request body and client not accept response body gzip encoding",
			contentType:        "application/json",
			contentEncoding:    "",
			body:               `{"answer": 42}`,
			acceptEncoding:     "",
			expectedGzip:       false,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "gzip request body and client not accept response body gzip encoding",
			contentType:        "application/json",
			contentEncoding:    "gzip",
			body:               `{"answer": 42}`,
			acceptEncoding:     "",
			expectedGzip:       false,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "no gzip request body and client accept response body gzip encoding",
			contentType:        "application/json",
			contentEncoding:    "",
			body:               `{"answer": 42}`,
			acceptEncoding:     "gzip",
			expectedGzip:       true,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "gzip request body and client accept response body gzip encoding",
			contentType:        "application/json",
			contentEncoding:    "gzip",
			body:               `{"answer": 42}`,
			acceptEncoding:     "gzip",
			expectedGzip:       true,
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err, "Error reading request body")
				if tt.expectedGzip {
					w.Header().Set("Content-Type", tt.contentType)
					_, err = w.Write(body)
					require.NoError(t, err, "Error writing response body")
				} else {
					_, err = w.Write(body)
					require.NoError(t, err, "Error writing response body")
				}
			})

			var requestBody io.Reader
			if tt.contentEncoding == "gzip" {
				var b bytes.Buffer
				gz := gzip.NewWriter(&b)
				_, err := gz.Write([]byte(tt.body))
				require.NoError(t, err, "Error zipping request body")
				err = gz.Close()
				require.NoError(t, err, "Error closing gzip writer")
				requestBody = &b
			} else {
				requestBody = strings.NewReader(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/", requestBody)
			req.Header.Set("Content-Type", tt.contentType)
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}
			req.Header.Set("Accept-Encoding", tt.acceptEncoding)

			w := httptest.NewRecorder()
			GzipMiddleware(handler).ServeHTTP(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode, "Response status code does not match expected status")

			if tt.expectedGzip {
				assert.Contains(t, w.Header().Get("Content-Encoding"), "gzip", "Expected gzipped response body")

				gz, err := gzip.NewReader(w.Body)
				require.NoError(t, err, "Error reading gzipped response body")
				responseBody, err := io.ReadAll(gz)
				require.NoError(t, err, "Error reading gzipped response body")
				err = gz.Close()
				require.NoError(t, err, "Error closing gzip writer")

				assert.Equal(t, tt.body, string(responseBody), "Expected response body %s, got %s", tt.body, responseBody)
			} else {
				assert.NotContains(t, w.Header().Get("Content-Encoding"), "gzip", "Unexpected gzipped response body")
				assert.Equal(t, tt.body, w.Body.String(), "Expected response body %s, got %s", tt.body, w.Body.String())
			}
		})
	}
}
