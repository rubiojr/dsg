package datahub

import (
	"encoding/json"
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
func (c *Client) PostDatasets(payload string) (int, error) {
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
			err := c.postSingleDataset(string(dataset))
			if err != nil {
				return 0, fmt.Errorf("error posting dataset %d: %w", i+1, err)
			}
		}

		return count, nil
	}

	// If it's not an array, post as single dataset
	return 1, c.postSingleDataset(payload)
}

// postSingleDataset sends a single dataset to the DataHub API
func (c *Client) postSingleDataset(payload string) error {
	url := fmt.Sprintf("%s/openapi/v3/entity/dataset?async=false&systemMetadata=false", c.URL)
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
