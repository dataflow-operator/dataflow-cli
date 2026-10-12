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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client calls the dataflow-web API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates an API client. baseURL should be the API base (e.g. "http://localhost:8080/api").
func New(baseURL string) *Client {
	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{},
	}
}

func (c *Client) url(path string, q url.Values) string {
	u := c.baseURL + path
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	return u
}

// GetNamespaces returns list of namespace names.
func (c *Client) GetNamespaces() ([]string, error) {
	var out []string
	return out, c.do("GET", c.url("/namespaces", nil), nil, &out)
}

// ListDataFlows returns list of DataFlow items in the given namespace.
func (c *Client) ListDataFlows(namespace string) ([]interface{}, error) {
	q := url.Values{}
	if namespace != "" {
		q.Set("namespace", namespace)
	}
	var out []interface{}
	return out, c.do("GET", c.url("/dataflows", q), nil, &out)
}

// GetDataFlow returns one DataFlow by name.
func (c *Client) GetDataFlow(namespace, name string) (map[string]interface{}, error) {
	q := url.Values{}
	if namespace != "" {
		q.Set("namespace", namespace)
	}
	var out map[string]interface{}
	return out, c.do("GET", c.url("/dataflows/"+url.PathEscape(name), q), nil, &out)
}

// CreateDataFlow creates a DataFlow (body is the full DataFlow JSON).
func (c *Client) CreateDataFlow(namespace string, body []byte) (map[string]interface{}, error) {
	q := url.Values{}
	if namespace != "" {
		q.Set("namespace", namespace)
	}
	var out map[string]interface{}
	return out, c.do("POST", c.url("/dataflows", q), body, &out)
}

// UpdateDataFlow updates a DataFlow by name.
func (c *Client) UpdateDataFlow(namespace, name string, body []byte) (map[string]interface{}, error) {
	q := url.Values{}
	if namespace != "" {
		q.Set("namespace", namespace)
	}
	var out map[string]interface{}
	return out, c.do("PUT", c.url("/dataflows/"+url.PathEscape(name), q), body, &out)
}

// DeleteDataFlow deletes a DataFlow by name.
func (c *Client) DeleteDataFlow(namespace, name string) error {
	q := url.Values{}
	if namespace != "" {
		q.Set("namespace", namespace)
	}
	return c.doNoContent("DELETE", c.url("/dataflows/"+url.PathEscape(name), q), nil)
}

// GetLogs returns logs (non-streaming). For streaming use OpenLogStream.
func (c *Client) GetLogs(namespace, name string, tailLines int) (string, error) {
	q := url.Values{}
	q.Set("namespace", namespace)
	q.Set("name", name)
	q.Set("follow", "false")
	if tailLines > 0 {
		q.Set("tailLines", fmt.Sprintf("%d", tailLines))
	}
	resp, err := c.httpClient.Get(c.url("/logs", q))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("%s: %s", resp.Status, string(b))
	}
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}

// OpenLogStream opens a GET request to /logs?follow=true and returns the response body for streaming.
func (c *Client) OpenLogStream(namespace, name string, tailLines int) (io.ReadCloser, error) {
	q := url.Values{}
	q.Set("namespace", namespace)
	q.Set("name", name)
	q.Set("follow", "true")
	if tailLines > 0 {
		q.Set("tailLines", fmt.Sprintf("%d", tailLines))
	}
	resp, err := c.httpClient.Get(c.url("/logs", q))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("logs stream: %s", resp.Status)
	}
	return resp.Body, nil
}

// GetStatus returns status object (phase, message, processedCount, errorCount, lastProcessedTime).
func (c *Client) GetStatus(namespace, name string) (map[string]interface{}, error) {
	q := url.Values{}
	q.Set("namespace", namespace)
	q.Set("name", name)
	var out map[string]interface{}
	return out, c.do("GET", c.url("/status", q), nil, &out)
}

// GetMetrics returns metrics object.
func (c *Client) GetMetrics(namespace, name string) (map[string]interface{}, error) {
	q := url.Values{}
	q.Set("namespace", namespace)
	q.Set("name", name)
	var out map[string]interface{}
	return out, c.do("GET", c.url("/metrics", q), nil, &out)
}

func (c *Client) do(method, rawURL string, body []byte, out interface{}) error {
	var reqBody io.Reader
	if len(body) > 0 {
		reqBody = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, rawURL, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s: %s", resp.Status, string(data))
	}
	if out != nil && len(data) > 0 {
		return json.Unmarshal(data, out)
	}
	return nil
}

func (c *Client) doNoContent(method, rawURL string, body []byte) error {
	var reqBody io.Reader
	if len(body) > 0 {
		reqBody = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, rawURL, reqBody)
	if err != nil {
		return err
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		return nil
	}
	data, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("%s: %s", resp.Status, string(data))
}
