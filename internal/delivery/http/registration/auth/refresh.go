package auth

import (
	"encoding/json"
	"net/http"

	"github.com/Alias1177/Auth/pkg/jwt"
)

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var ref jwt.RefreshTokenStruct
	if err := json.NewDecoder(r.Body).Decode(&ref); err != nil {
		http.Error(w, "Некорректный запрос", http.StatusBadRequest)
		return
	}
	newaccess, newrefresh, err := h.tokenManager.RefreshTokens(ref.Token)
	if err != nil {
		http.Error(w, "Неверный токен", http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  newaccess,
		"refresh_token": newrefresh,
	})
}
