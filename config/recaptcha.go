package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"
)

type recaptchaResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
}

// VerifyRecaptcha verifies a reCAPTCHA token using Google's siteverify API.
func VerifyRecaptcha(token string) (bool, error) {
	secret := os.Getenv("RECAPTCHA_SECRET")
	if secret == "" {
		return false, errors.New("recaptcha secret not configured")
	}

	endpoint := "https://www.google.com/recaptcha/api/siteverify"

	data := url.Values{}
	data.Set("secret", secret)
	data.Set("response", token)

	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.PostForm(endpoint, data)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result recaptchaResponse

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return false, err
	}

	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		return false, err
	}

	if !result.Success {
		return false, errors.New("recaptcha verification failed")
	}

	return true, nil
}
