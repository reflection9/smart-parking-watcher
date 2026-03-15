package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type SubscriptionLookupClient interface {
	ListUserIDsByZoneID(ctx context.Context, zoneID int64) ([]int64, error)
}

type subscriptionLookupResponse struct {
	UserID int64 `json:"user_id"`
}

type httpSubscriptionLookupClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPSubscriptionLookupClient(baseURL string) SubscriptionLookupClient {
	return &httpSubscriptionLookupClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (c *httpSubscriptionLookupClient) ListUserIDsByZoneID(ctx context.Context, zoneID int64) ([]int64, error) {
	if c.baseURL == "" || strings.HasPrefix(c.baseURL, "/") {
		return nil, fmt.Errorf("invalid subscription-service url")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/subscriptions/zones/%d", c.baseURL, zoneID),
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
		var subscriptions []subscriptionLookupResponse
		if err := json.NewDecoder(resp.Body).Decode(&subscriptions); err != nil {
			return nil, err
		}

		userIDs := make([]int64, 0, len(subscriptions))
		for _, subscription := range subscriptions {
			userIDs = append(userIDs, subscription.UserID)
		}

		return userIDs, nil
	case http.StatusNotFound:
		return []int64{}, nil
	default:
		return nil, fmt.Errorf("unexpected status code from subscription-service: %d", resp.StatusCode)
	}
}
