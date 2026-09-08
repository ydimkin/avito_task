package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CoreClient клиент для взаимодействия эмулятора ресторана с Core Service.
type CoreClient interface {
	UpdateOrderStatus(ctx context.Context, orderID string, status string, comment *string) error
	SetItemAvailability(ctx context.Context, restaurantID, itemID string, isAvailable bool) error
}

type coreClient struct {
	baseURL      string
	restaurantID string
	httpClient   *http.Client
}

// NewCoreClient создает экземпляр CoreClient.
func NewCoreClient(baseURL, restaurantID string) CoreClient {
	return &coreClient{
		baseURL:      baseURL,
		restaurantID: restaurantID,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type updateStatusReq struct {
	Status  string  `json:"status"`
	Comment *string `json:"comment,omitempty"`
}

func (c *coreClient) UpdateOrderStatus(ctx context.Context, orderID string, status string, comment *string) error {
	url := fmt.Sprintf("%s/api/v1/partner/orders/%s/status", c.baseURL, orderID)
	body, err := json.Marshal(updateStatusReq{Status: status, Comment: comment})
	if err != nil {
		return fmt.Errorf("failed to marshal status request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create patch request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.restaurantID != "" {
		req.Header.Set("X-Partner-ID", c.restaurantID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute status update: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("core service status update returned status: %d", resp.StatusCode)
	}

	return nil
}

type updateAvailabilityReq struct {
	IsAvailable bool `json:"is_available"`
}

func (c *coreClient) SetItemAvailability(ctx context.Context, restaurantID, itemID string, isAvailable bool) error {
	url := fmt.Sprintf("%s/api/v1/partner/restaurants/%s/items/%s/availability", c.baseURL, restaurantID, itemID)
	body, err := json.Marshal(updateAvailabilityReq{IsAvailable: isAvailable})
	if err != nil {
		return fmt.Errorf("failed to marshal availability request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create patch request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Partner-ID", restaurantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute availability update: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("core service availability update returned status: %d", resp.StatusCode)
	}

	return nil
}
