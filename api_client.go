package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// APIClient gère les appels à l'API Tracker avec retry
type APIClient struct {
	baseURL    string
	httpClient *http.Client
	maxRetries int
}

// NewAPIClient crée un nouveau client API
func NewAPIClient() *APIClient {
	return &APIClient{
		baseURL: os.Getenv("TRACKER_HOST"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries: 3,
	}
}

// APIResponse représente la réponse de l'API
type APIResponse struct {
	Success    bool
	StatusCode int
	Body       []byte
	Error      error
}

// PostEventWithRetry envoie un événement à l'API avec retry et exponential backoff
func (c *APIClient) PostEventWithRetry(payload Payload) (*APIResponse, error) {
	endpoint := c.baseURL + "/api/v1alpha1/event"

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return &APIResponse{
			Success: false,
			Error:   fmt.Errorf("failed to marshal payload: %w", err),
		}, err
	}

	return c.makeRequestWithRetry("POST", endpoint, payloadBytes)
}

// PutEventWithRetry met à jour un événement dans l'API avec retry
func (c *APIClient) PutEventWithRetry(payload Payload) (*APIResponse, error) {
	endpoint := c.baseURL + "/api/v1alpha1/event"

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return &APIResponse{
			Success: false,
			Error:   fmt.Errorf("failed to marshal payload: %w", err),
		}, err
	}

	return c.makeRequestWithRetry("PUT", endpoint, payloadBytes)
}

// UpdateSlackIdWithRetry met à jour le SlackId d'un événement avec retry
func (c *APIClient) UpdateSlackIdWithRetry(tempSlackId, realSlackId string) (*APIResponse, error) {
	endpoint := c.baseURL + "/api/v1alpha1/event/" + tempSlackId + "/slack-id"

	payload := map[string]string{
		"slack_id": realSlackId,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return &APIResponse{
			Success: false,
			Error:   fmt.Errorf("failed to marshal slack_id payload: %w", err),
		}, err
	}

	return c.makeRequestWithRetry("PATCH", endpoint, payloadBytes)
}

// makeRequestWithRetry effectue une requête HTTP avec retry et backoff personnalisé
func (c *APIClient) makeRequestWithRetry(method, url string, body []byte) (*APIResponse, error) {
	var lastErr error
	// Délais de backoff personnalisés : 3s, 5s, 10s
	backoffDelays := []time.Duration{3 * time.Second, 5 * time.Second, 10 * time.Second}

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			backoffDuration := backoffDelays[attempt-1]
			fmt.Printf("API call failed, retrying in %v (attempt %d/%d)\n", backoffDuration, attempt, c.maxRetries)
			time.Sleep(backoffDuration)
		}

		fmt.Printf("Making %s request to %s (attempt %d/%d)\n", method, url, attempt+1, c.maxRetries+1)

		req, err := http.NewRequest(method, url, bytes.NewReader(body))
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			fmt.Printf("Request error: %v\n", err)
			continue
		}

		defer resp.Body.Close()

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			fmt.Printf("Response read error: %v\n", err)
			continue
		}

		// Considérer les codes 2xx comme succès
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			fmt.Printf("API call successful (status: %d)\n", resp.StatusCode)
			return &APIResponse{
				Success:    true,
				StatusCode: resp.StatusCode,
				Body:       responseBody,
				Error:      nil,
			}, nil
		}

		// Pour les erreurs 4xx, ne pas retry (erreur client)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			lastErr = fmt.Errorf("client error (status: %d): %s", resp.StatusCode, string(responseBody))
			fmt.Printf("Client error, not retrying: %v\n", lastErr)
			return &APIResponse{
				Success:    false,
				StatusCode: resp.StatusCode,
				Body:       responseBody,
				Error:      lastErr,
			}, lastErr
		}

		// Pour les erreurs 5xx, continuer à retry
		lastErr = fmt.Errorf("server error (status: %d): %s", resp.StatusCode, string(responseBody))
		fmt.Printf("Server error, will retry: %v\n", lastErr)
	}

	// Tous les tentatives ont échoué
	finalErr := fmt.Errorf("all %d attempts failed, last error: %w", c.maxRetries+1, lastErr)
	return &APIResponse{
		Success: false,
		Error:   finalErr,
	}, finalErr
}

// IsAPIAvailable vérifie si l'API est disponible
func (c *APIClient) IsAPIAvailable() bool {
	if c.baseURL == "" {
		fmt.Println("TRACKER_HOST not configured")
		return false
	}

	req, err := http.NewRequest("GET", c.baseURL+"/health", nil)
	if err != nil {
		fmt.Printf("Failed to create health check request: %v\n", err)
		return false
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Health check failed: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	available := resp.StatusCode >= 200 && resp.StatusCode < 300
	fmt.Printf("API health check: %v (status: %d)\n", available, resp.StatusCode)
	return available
}

// Instance globale du client API
var apiClient = NewAPIClient()
