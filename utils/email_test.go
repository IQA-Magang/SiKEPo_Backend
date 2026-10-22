package utils

import "testing"

func TestManagerVerificationURLUsesHashRouteWithoutID(t *testing.T) {
	t.Setenv("MANAGER_VERIFICATION_PATH", "/#/verifikasi")

	got, err := managerVerificationURL("http://localhost:5174/", 42)
	if err != nil {
		t.Fatalf("managerVerificationURL() returned error: %v", err)
	}
	want := "http://localhost:5174/#/verifikasi"
	if got != want {
		t.Fatalf("managerVerificationURL() = %q, want %q", got, want)
	}
}

func TestManagerVerificationURLSupportsIDPlaceholder(t *testing.T) {
	t.Setenv("MANAGER_VERIFICATION_PATH", "/#/verifikasi/{id}")

	got, err := managerVerificationURL("http://localhost:5174/", 42)
	if err != nil {
		t.Fatalf("managerVerificationURL() returned error: %v", err)
	}
	want := "http://localhost:5174/#/verifikasi/42"
	if got != want {
		t.Fatalf("managerVerificationURL() = %q, want %q", got, want)
	}
}

func TestSMTPUsername(t *testing.T) {
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_EMAIL", "manager@example.com")
	if got := smtpUsername(); got != "manager@example.com" {
		t.Fatalf("smtpUsername() = %q, want SMTP_EMAIL fallback", got)
	}

	t.Setenv("SMTP_USERNAME", "smtp-user@example.com")
	if got := smtpUsername(); got != "smtp-user@example.com" {
		t.Fatalf("smtpUsername() = %q, want SMTP_USERNAME to take precedence", got)
	}
}
