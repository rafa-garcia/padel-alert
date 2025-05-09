package playtomic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	BaseURL        = "https://api.playtomic.io/v1"
	defaultTimeout = 30 * time.Second
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		baseURL: BaseURL,
	}
}

func (c *Client) WithTimeout(timeout time.Duration) *Client {
	c.httpClient.Timeout = timeout
	return c
}

// sendRequest sends a request to the Playtomic API and decodes the response into the provided result pointer
func (c *Client) sendRequest(ctx context.Context, endpoint string, queryParams string, result interface{}) error {
	reqURL := fmt.Sprintf("%s%s?%s", c.baseURL, endpoint, queryParams)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "PadelAlert/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	return nil
}

func (c *Client) GetClasses(ctx context.Context, params *SearchClassesParams) ([]Class, error) {
	var classes []Class
	err := c.sendRequest(ctx, "/classes", params.ToURLValues().Encode(), &classes)
	if err != nil {
		return nil, err
	}
	return classes, nil
}

func (c *Client) GetMatches(ctx context.Context, params *SearchMatchesParams) ([]Match, error) {
	var matches []Match
	err := c.sendRequest(ctx, "/matches", params.ToURLValues().Encode(), &matches)
	if err != nil {
		return nil, err
	}
	return matches, nil
}
