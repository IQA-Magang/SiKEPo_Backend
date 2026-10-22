package utils

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const smtpTimeout = 15 * time.Second

func SendVerificationSubmissionEmail(
	recipientEmail string,
	applicantName string,
	equipmentName string,
	assetNumber string,
	verificationID uint64,
) error {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	username := smtpUsername()
	password := os.Getenv("SMTP_PASSWORD")
	fromAddress := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	frontendURL := strings.TrimSpace(os.Getenv("FRONTEND_URL"))

	if host == "" || port == "" || frontendURL == "" {
		return fmt.Errorf("SMTP_HOST, SMTP_PORT, dan FRONTEND_URL wajib dikonfigurasi")
	}
	if (username == "") != (password == "") {
		return fmt.Errorf("SMTP_EMAIL atau SMTP_USERNAME harus diisi bersama SMTP_PASSWORD")
	}
	if fromAddress == "" {
		fromAddress = username
	}

	from, err := mail.ParseAddress(fromAddress)
	if err != nil {
		return fmt.Errorf("SMTP_FROM tidak valid: %w", err)
	}
	recipient, err := mail.ParseAddress(recipientEmail)
	if err != nil {
		return fmt.Errorf("alamat email Manager tidak valid: %w", err)
	}

	verificationURL, err := managerVerificationURL(frontendURL, verificationID)
	if err != nil {
		return err
	}

	message := strings.Join([]string{
		"From: " + from.String(),
		"To: " + recipient.String(),
		"Subject: Pengajuan verifikasi peralatan baru",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		"Pengelola Lab mengajukan verifikasi peralatan berikut:",
		"",
		"Nama peralatan: " + equipmentName,
		"Nomor aset: " + assetNumber,
		"Nama pengaju: " + applicantName,
		"",
		"Tinjau pengajuan melalui halaman verifikasi Manager:",
		verificationURL,
		"",
	}, "\r\n")

	address := net.JoinHostPort(host, port)
	var connection net.Conn
	if port == "465" {
		connection, err = tls.DialWithDialer(
			&net.Dialer{Timeout: smtpTimeout},
			"tcp",
			address,
			&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12},
		)
	} else {
		connection, err = net.DialTimeout("tcp", address, smtpTimeout)
	}
	if err != nil {
		return fmt.Errorf("gagal terhubung ke server SMTP: %w", err)
	}
	defer connection.Close()

	if err := connection.SetDeadline(time.Now().Add(smtpTimeout)); err != nil {
		return fmt.Errorf("gagal menetapkan batas waktu koneksi SMTP: %w", err)
	}

	client, err := smtp.NewClient(connection, host)
	if err != nil {
		return fmt.Errorf("gagal memulai sesi SMTP: %w", err)
	}
	defer client.Close()

	if port != "465" {
		supported, _ := client.Extension("STARTTLS")
		if !supported {
			return fmt.Errorf("server SMTP tidak mendukung STARTTLS")
		}
		if err := client.StartTLS(&tls.Config{
			ServerName: host,
			MinVersion: tls.VersionTLS12,
		}); err != nil {
			return fmt.Errorf("gagal mengaktifkan TLS SMTP: %w", err)
		}
	}

	if username != "" {
		auth := smtp.PlainAuth("", username, password, host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("autentikasi SMTP gagal: %w", err)
		}
	}
	if err := client.Mail(from.Address); err != nil {
		return fmt.Errorf("SMTP menolak alamat pengirim: %w", err)
	}
	if err := client.Rcpt(recipient.Address); err != nil {
		return fmt.Errorf("SMTP menolak alamat penerima: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("gagal memulai pengiriman email SMTP: %w", err)
	}
	if _, err := writer.Write([]byte(message)); err != nil {
		_ = writer.Close()
		return fmt.Errorf("gagal mengirim isi email: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("server SMTP gagal menerima email: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("gagal menutup sesi SMTP: %w", err)
	}

	return nil
}

func smtpUsername() string {
	username := strings.TrimSpace(os.Getenv("SMTP_USERNAME"))
	if username == "" {
		username = strings.TrimSpace(os.Getenv("SMTP_EMAIL"))
	}
	return username
}

func managerVerificationURL(frontendURL string, verificationID uint64) (string, error) {
	base, err := url.Parse(frontendURL)
	if err != nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" {
		return "", fmt.Errorf("FRONTEND_URL harus berupa URL HTTP atau HTTPS yang valid")
	}

	path := strings.TrimSpace(os.Getenv("MANAGER_VERIFICATION_PATH"))
	if path == "" {
		path = "/#/verifikasi"
	}

	if strings.Contains(path, "{id}") {
		path = strings.ReplaceAll(path, "{id}", strconv.FormatUint(verificationID, 10))
	}

	route, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("MANAGER_VERIFICATION_PATH tidak valid: %w", err)
	}

	if route.Fragment != "" {
		if route.Path != "" && route.Path != "/" {
			base.Path = strings.TrimRight(base.Path, "/") + route.Path
		}
		base.Fragment = route.Fragment
	} else {
		routePath := route.Path
		if !strings.HasPrefix(routePath, "/") {
			routePath = "/" + routePath
		}
		base.Path = strings.TrimRight(base.Path, "/") + routePath
	}

	return base.String(), nil
}
