package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type UserLookupClient interface {
	Exists(ctx context.Context, userID int64) (bool, error)
}

type httpUserLookupClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPUserLookupClient(baseURL string) UserLookupClient {
	return &httpUserLookupClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (c *httpUserLookupClient) Exists(ctx context.Context, userID int64) (bool, error) {
	if c.baseURL == "" || strings.HasPrefix(c.baseURL, "/") {
		return false, errors.New("invalid user-service url")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/users/%d", c.baseURL, userID),
		nil,
	)
	if err != nil {
		return false, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("unexpected status code from user-service: %d", resp.StatusCode)
	}
}
