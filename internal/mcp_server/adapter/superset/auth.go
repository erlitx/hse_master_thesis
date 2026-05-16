package superset

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Тело запроса входа в Superset
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Provider string `json:"provider"`
	Refresh  bool   `json:"refresh"`
}

// Ответ входа в Superset
type loginResponse struct {
	AccessToken string `json:"access_token"`
}

// Получает токен доступа Superset
func (a *BiTool) accessToken(ctx context.Context, providedToken string) (string, error) {
	token := strings.TrimSpace(providedToken)
	if token != "" {
		return token, nil
	}

	token = strings.TrimSpace(a.cfg.JWTToken)
	if token != "" {
		return token, nil
	}

	reqBody, err := json.Marshal(loginRequest{
		Username: a.cfg.User,
		Password: a.cfg.Password,
		Provider: "db",
		Refresh:  true,
	})
	if err != nil {
		return "", fmt.Errorf("json.Marshal login request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.cfg.URL+"/api/v1/security/login", bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("http.NewRequestWithContext login: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("superset login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("superset login failed with status %d", resp.StatusCode)
	}

	var parsed loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("decode login response: %w", err)
	}
	if strings.TrimSpace(parsed.AccessToken) == "" {
		return "", fmt.Errorf("superset login returned empty access token")
	}

	return parsed.AccessToken, nil
}
