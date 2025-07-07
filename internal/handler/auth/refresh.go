package auth

import (
	"encoding/json"
	"net/http"

	"github.com/Alias1177/Auth/internal/dto"
	"github.com/Alias1177/Auth/pkg/errors"
	"github.com/Alias1177/Auth/pkg/httputil"
	"github.com/Alias1177/Auth/pkg/jwt"
)

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var ref jwt.RefreshTokenStruct
	if err := json.NewDecoder(r.Body).Decode(&ref); err != nil {
		httputil.JSONErrorWithID(w, http.StatusBadRequest, dto.MsgInvalidRequest)
		return
	}

	newaccess, newrefresh, err := h.tokenManager.RefreshTokens(ref.Token)
	if err != nil {
		httputil.JSONErrorWithID(w, http.StatusBadRequest, dto.MsgTokenInvalid)
		return
	}

	response := map[string]string{
		"access_token":  newaccess,
		"refresh_token": newrefresh,
	}

	if err := httputil.JSONSuccessWithID(w, http.StatusOK, dto.MsgSuccessTokenRefresh, response); err != nil {
		errors.HandleInternalError(w, err, h.logger, "encode response")
	}
}
