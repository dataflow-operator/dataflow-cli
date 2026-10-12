/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package apiclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetNamespaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/namespaces", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]string{"default", "other"})
	}))
	defer server.Close()

	client := New(server.URL)
	list, err := client.GetNamespaces()
	require.NoError(t, err)
	assert.Equal(t, []string{"default", "other"}, list)
}

func TestClient_GetStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/status", r.URL.Path)
		assert.Equal(t, "default", r.URL.Query().Get("namespace"))
		assert.Equal(t, "my-df", r.URL.Query().Get("name"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"phase":          "Running",
			"processedCount": float64(42),
		})
	}))
	defer server.Close()

	client := New(server.URL)
	out, err := client.GetStatus("default", "my-df")
	require.NoError(t, err)
	assert.Equal(t, "Running", out["phase"])
	assert.Equal(t, float64(42), out["processedCount"])
}
