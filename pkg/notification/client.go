package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// NotificationClient клиент для работы с Notification Service
type NotificationClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewNotificationClient создает новый клиент
func NewNotificationClient(baseURL string) *NotificationClient {
	return &NotificationClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ValidateCodeRequest запрос на валидацию кода
type ValidateCodeRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

// ValidateCodeResponse ответ на валидацию кода
type ValidateCodeResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
}

// ValidatePasswordResetCode валидирует код восстановления пароля
func (c *NotificationClient) ValidatePasswordResetCode(email, code string) (bool, error) {
	request := ValidateCodeRequest{
		Email: email,
		Code:  code,
	}

	requestData, err := json.Marshal(request)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/validate",
		"application/json",
		bytes.NewBuffer(requestData),
	)

	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response ValidateCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Valid, nil
}
