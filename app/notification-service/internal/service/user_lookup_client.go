package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	observability "smart-parking-observability"
)

type UserLookupClient interface {
	GetByID(ctx context.Context, userID int64) (*UserLookupResult, error)
}

type UserLookupResult struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type httpUserLookupClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPUserLookupClient(baseURL string) UserLookupClient {
	return &httpUserLookupClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  observability.NewHTTPClient(3 * time.Second),
	}
}

func (c *httpUserLookupClient) GetByID(ctx context.Context, userID int64) (*UserLookupResult, error) {
	if c.baseURL == "" || strings.HasPrefix(c.baseURL, "/") {
		return nil, fmt.Errorf("invalid user-service url")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/users/%d", c.baseURL, userID),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var user UserLookupResult
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			return nil, err
		}

		return &user, nil
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, fmt.Errorf("unexpected status code from user-service: %d", resp.StatusCode)
	}
}
