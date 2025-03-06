package datahub

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client represents the DataHub API client
type Client struct {
	URL        string
	Token      string
	HttpClient *http.Client
}

// NewClient creates a new DataHub client
func NewClient(url string, token string) *Client {
	if url == "" {
		url = "http://localhost:8080"
	}

	return &Client{
		URL:        url,
		Token:      token,
		HttpClient: http.DefaultClient,
	}
}

// PostDataset sends one or more datasets to the DataHub API
func (c *Client) PostEntity(resource, payload string) (int, error) {
	// Check if the payload is an array of datasets
	trimmedPayload := strings.TrimSpace(payload)

	// Simple check to see if it starts with [ and ends with ]
	if strings.HasPrefix(trimmedPayload, "[") && strings.HasSuffix(trimmedPayload, "]") {
		// Parse the JSON array using the standard library
		var datasets []json.RawMessage
		if err := json.Unmarshal([]byte(trimmedPayload), &datasets); err != nil {
			return 0, fmt.Errorf("error parsing dataset array: %w", err)
		}

		// Post each dataset individually
		count := len(datasets)
		for i, dataset := range datasets {
			err := c.postSingleEntity(resource, string(dataset))
			if err != nil {
				return 0, fmt.Errorf("error posting dataset %d: %w", i+1, err)
			}
		}

		return count, nil
	}

	return 0, errors.New("error parsing dataset array")
	// If it's not an array, post as single dataset
	///return 1, c.postSingleEntity(resource, payload)
}

// postSingleDataset sends a single dataset to the DataHub API
func (c *Client) postSingleEntity(resource, payload string) error {
	url := fmt.Sprintf("%s/openapi/v3/entity/%s?async=false&systemMetadata=false", c.URL, resource)
	req, err := http.NewRequest("POST", url, strings.NewReader("["+payload+"]"))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	return nil
}
