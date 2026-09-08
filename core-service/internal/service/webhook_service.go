package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"core-service/internal/domain"
)

type webhookService struct {
	httpClient *http.Client
}

// NewWebhookService создает экземпляр WebhookNotifier.
func NewWebhookService(timeoutSec int) WebhookNotifier {
	timeout := time.Duration(timeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &webhookService{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (w *webhookService) NotifyNewOrder(ctx context.Context, webhookURL string, order *domain.Order) error {
	payload, err := json.Marshal(order)
	if err != nil {
		return err
	}

	targetURL := webhookURL
	if override := os.Getenv("WEBHOOK_HOST_OVERRIDE"); override != "" {
		targetURL = strings.Replace(targetURL, "localhost:8081", override, 1)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return domain.ErrWebhookFailed
	}

	return nil
}
