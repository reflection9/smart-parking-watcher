package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	observability "smart-parking-observability"
)

type ParkingSpotClient interface {
	GetByID(ctx context.Context, zoneID, spotID int64) (*ParkingSpotLookupResult, error)
	ReserveSpot(ctx context.Context, zoneID, spotID int64) error
	ReleaseSpot(ctx context.Context, zoneID, spotID int64) error
	OccupySpot(ctx context.Context, zoneID, spotID int64) error
}

type ParkingSpotLookupResult struct {
	ID        int64     `json:"id"`
	ZoneID    int64     `json:"zone_id"`
	Number    string    `json:"number"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type httpParkingSpotClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPParkingSpotClient(baseURL string) ParkingSpotClient {
	return &httpParkingSpotClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  observability.NewHTTPClient(3 * time.Second),
	}
}

func (c *httpParkingSpotClient) GetByID(ctx context.Context, zoneID, spotID int64) (*ParkingSpotLookupResult, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("%s/spots/%d/zones/%d", c.baseURL, spotID, zoneID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var spot ParkingSpotLookupResult
		if err := json.NewDecoder(resp.Body).Decode(&spot); err != nil {
			return nil, err
		}

		return &spot, nil
	case http.StatusNotFound:
		return nil, ErrSpotNotFound
	default:
		return nil, fmt.Errorf("%w: unexpected status code from parking-service: %d", ErrDependencyUnavailable, resp.StatusCode)
	}
}

func (c *httpParkingSpotClient) ReserveSpot(ctx context.Context, zoneID, spotID int64) error {
	return c.postTransition(ctx, zoneID, spotID, "reserve")
}

func (c *httpParkingSpotClient) ReleaseSpot(ctx context.Context, zoneID, spotID int64) error {
	return c.postTransition(ctx, zoneID, spotID, "release")
}

func (c *httpParkingSpotClient) OccupySpot(ctx context.Context, zoneID, spotID int64) error {
	return c.postTransition(ctx, zoneID, spotID, "occupy")
}

func (c *httpParkingSpotClient) postTransition(
	ctx context.Context,
	zoneID, spotID int64,
	action string,
) error {
	resp, err := c.doRequest(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/zones/%d/spots/%d/%s", c.baseURL, zoneID, spotID, action),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return ErrSpotNotFound
	case http.StatusConflict:
		return ErrSpotUnavailable
	default:
		return fmt.Errorf("%w: unexpected status code from parking-service: %d", ErrDependencyUnavailable, resp.StatusCode)
	}
}

func (c *httpParkingSpotClient) doRequest(ctx context.Context, method, url string) (*http.Response, error) {
	if c.baseURL == "" || strings.HasPrefix(c.baseURL, "/") {
		return nil, errors.New("invalid parking-service url")
	}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDependencyUnavailable, err)
	}

	return resp, nil
}
