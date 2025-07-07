package dto

// MessageID константы для всех сообщений в системе
const (
	// Успешные операции (1000-1999)
	MsgSuccessLogin                  = 1000
	MsgSuccessRegister               = 1001
	MsgSuccessLogout                 = 1002
	MsgSuccessPasswordReset          = 1003
	MsgSuccessUserUpdated            = 1004
	MsgSuccessTokenRefresh           = 1005
	MsgSuccessPasswordChanged        = 1006
	MsgSuccessPasswordResetRequested = 1007
	MsgSuccessPasswordResetConfirmed = 1008
	MsgSuccessUserInfoRetrieved      = 1009

	// Ошибки валидации (2000-2999)
	MsgInvalidEmail           = 2000
	MsgInvalidPassword        = 2001
	MsgInvalidToken           = 2002
	MsgInvalidRequest         = 2003
	MsgInvalidResetCode       = 2004
	MsgPasswordTooWeak        = 2005
	MsgEmailAlreadyExists     = 2006
	MsgInvalidUserID          = 2007
	MsgMissingEmailOrPassword = 2008
	MsgInvalidRequestData     = 2009

	// Ошибки аутентификации (3000-3999)
	MsgWrongPassword        = 3000
	MsgUserNotFound         = 3001
	MsgTokenExpired         = 3002
	MsgTokenInvalid         = 3003
	MsgUnauthorized         = 3004
	MsgResetCodeExpired     = 3005
	MsgResetCodeInvalid     = 3006
	MsgResourceNotFound     = 3007
	MsgTooManyResetAttempts = 3008

	// Ошибки сервера (4000-4999)
	MsgInternalError        = 4000
	MsgDatabaseError        = 4001
	MsgEmailSendError       = 4002
	MsgTokenGenerationError = 4003
	MsgHashPasswordError    = 4004
	MsgEncodeResponseError  = 4005
	MsgGetUserClaimsError   = 4006
)

// ErrorResponse DTO для ошибок с id_message
type ErrorResponse struct {
	Error     string `json:"error"`
	IdMessage int    `json:"id_message"`
	Code      int    `json:"code,omitempty"`
}

// SuccessResponse DTO для успешных ответов с id_message
type SuccessResponse struct {
	IdMessage int         `json:"id_message"`
	Data      interface{} `json:"data,omitempty"`
}

// PaginationRequest DTO для пагинации запросов
type PaginationRequest struct {
	Page  int `json:"page" validate:"min=1"`
	Limit int `json:"limit" validate:"min=1,max=100"`
}

// PaginationResponse DTO для пагинации ответов
type PaginationResponse struct {
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"total_pages"`
	Data       interface{} `json:"data"`
}

// HealthResponse DTO для health check
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version,omitempty"`
}
