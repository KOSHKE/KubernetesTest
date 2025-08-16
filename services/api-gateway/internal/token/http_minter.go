package token

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type HTTPMinter struct {
	URL    string
	Secret string
	Client *http.Client
}

func (m *HTTPMinter) MintAccessToken(ctx context.Context, userID string) (string, error) {
	if m.Client == nil {
		m.Client = &http.Client{Timeout: 3 * time.Second}
	}
	b, _ := json.Marshal(map[string]string{"user_id": userID})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.URL, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if m.Secret != "" {
		req.Header.Set("X-Internal-Token", m.Secret)
	}
	resp, err := m.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", ErrMintFailed
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Token, nil
}

var ErrMintFailed = &mintErr{}

type mintErr struct{}

func (*mintErr) Error() string { return "mint access token failed" }
