package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type RecaptchaResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
}

func VerifyRecaptcha(token string) (bool, error) {

	token = strings.TrimSpace(token)

	if token == "" {
		return false, errors.New("recaptcha token kosong")
	}

	secret := strings.TrimSpace(os.Getenv("RECAPTCHA_SECRET"))

	if secret == "" {
		return false, errors.New("RECAPTCHA_SECRET tidak ditemukan di environment")
	}

	data := url.Values{}
	data.Set("secret", secret)
	data.Set("response", token)

	resp, err := http.PostForm(
		"https://www.google.com/recaptcha/api/siteverify",
		data,
	)

	if err != nil {
		return false, fmt.Errorf(
			"gagal menghubungi Google reCAPTCHA: %w",
			err,
		)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf(
			"Google reCAPTCHA mengembalikan HTTP status %d",
			resp.StatusCode,
		)
	}

	var result RecaptchaResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf(
			"gagal membaca response Google reCAPTCHA: %w",
			err,
		)
	}

	if !result.Success {

		if len(result.ErrorCodes) > 0 {
			return false, fmt.Errorf(
				"Google reCAPTCHA menolak token: %s",
				strings.Join(result.ErrorCodes, ", "),
			)
		}

		return false, errors.New(
			"Google reCAPTCHA menolak token",
		)
	}

	return true, nil
}