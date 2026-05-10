package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	observability "smart-parking-observability"
	"strings"
	"time"
)

type NotificationDispatchClient interface {
	NotifyCurrentAvailability(ctx context.Context, userID int64, zone *ZoneDetails) error
}

type httpNotificationDispatchClient struct {
	baseURL string
	client  *http.Client
}

type noopNotificationDispatchClient struct{}

type dispatchNotificationRequest struct {
	EventID   string  `json:"event_id"`
	EventType string  `json:"event_type"`
	ZoneID    int64   `json:"zone_id"`
	SpotID    *int64  `json:"spot_id,omitempty"`
	UserIDs   []int64 `json:"user_ids,omitempty"`
}

func NewHTTPNotificationDispatchClient(baseURL string) NotificationDispatchClient {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		return noopNotificationDispatchClient{}
	}

	return &httpNotificationDispatchClient{
		baseURL: baseURL,
		client:  observability.NewHTTPClient(3 * time.Second),
	}
}

func (noopNotificationDispatchClient) NotifyCurrentAvailability(context.Context, int64, *ZoneDetails) error {
	return nil
}

func (c *httpNotificationDispatchClient) NotifyCurrentAvailability(
	ctx context.Context,
	userID int64,
	zone *ZoneDetails,
) error {
	if zone == nil || zone.AvailableSpots <= 0 {
		return nil
	}

	if c.baseURL == "" || strings.HasPrefix(c.baseURL, "/") {
		return errors.New("invalid notification-service url")
	}

	reqBody := dispatchNotificationRequest{
		EventID:   generateEventID(),
		EventType: "spot_available",
		ZoneID:    zone.ID,
		SpotID:    firstFreeSpotID(zone.Spots),
		UserIDs:   []int64{userID},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/notifications/spot-freed", c.baseURL),
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code from notification-service: %d", resp.StatusCode)
	}

	return nil
}

func firstFreeSpotID(spots []ZoneSpot) *int64 {
	for _, spot := range spots {
		if strings.EqualFold(spot.Status, "FREE") {
			spotID := spot.ID
			return &spotID
		}
	}

	return nil
}

func generateEventID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("evt-%d", time.Now().UnixNano())
	}

	return "evt-" + hex.EncodeToString(buffer)
}
