package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	observability "smart-parking-observability"
)

type UserLookupClient interface {
	Exists(ctx context.Context, userID int64) (bool, error)
}

type ZoneLookupClient interface {
	Exists(ctx context.Context, zoneID int64) (bool, error)
}

type httpUserLookupClient struct {
	baseURL string
	client  *http.Client
}

type httpZoneLookupClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPUserLookupClient(baseURL string) UserLookupClient {
	return &httpUserLookupClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  observability.NewHTTPClient(3 * time.Second),
	}
}

func NewHTTPZoneLookupClient(baseURL string) ZoneLookupClient {
	return &httpZoneLookupClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  observability.NewHTTPClient(3 * time.Second),
	}
}

func (c *httpUserLookupClient) Exists(ctx context.Context, userID int64) (bool, error) {
	return c.existsByURL(ctx, fmt.Sprintf("%s/users/%d", c.baseURL, userID))
}

func (c *httpZoneLookupClient) Exists(ctx context.Context, zoneID int64) (bool, error) {
	return c.existsByURL(ctx, fmt.Sprintf("%s/zones/%d", c.baseURL, zoneID))
}

func (c *httpUserLookupClient) existsByURL(ctx context.Context, url string) (bool, error) {
	return doExistsRequest(ctx, c.client, url)
}

func (c *httpZoneLookupClient) existsByURL(ctx context.Context, url string) (bool, error) {
	return doExistsRequest(ctx, c.client, url)
}

func doExistsRequest(ctx context.Context, client *http.Client, url string) (bool, error) {
	if url == "" || strings.HasPrefix(url, "/") {
		return false, errors.New("invalid dependency service url")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	resp, err := client.Do(req)
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
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
