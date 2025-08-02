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
			Timeout: 30 * time.Second, // Увеличиваем timeout для внешних подключений
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

	// Retry логика для внешних подключений
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := c.httpClient.Post(
			c.baseURL+"/api/validate",
			"application/json",
			bytes.NewBuffer(requestData),
		)

		if err != nil {
			if attempt == maxRetries {
				return false, fmt.Errorf("failed to send request after %d attempts: %w", maxRetries, err)
			}
			// Ждем перед повторной попыткой
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			if attempt == maxRetries {
				return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			}
			// Ждем перед повторной попыткой
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		var response ValidateCodeResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return false, fmt.Errorf("failed to decode response: %w", err)
		}

		return response.Valid, nil
	}

	return false, fmt.Errorf("failed to validate code after %d attempts", maxRetries)
}
