package domain

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

const (
	StatusAvailable   = "available"
	StatusUnavailable = "unavailable"

	MaxURLsPerBatch = 100
	MaxURLLength    = 2048
)

var ErrInvalidURL = errors.New("некорректный URL")

func ValidateURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, ErrInvalidURL
	}
	if len(raw) > MaxURLLength {
		return nil, ErrInvalidURL
	}

	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, ErrInvalidURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, ErrInvalidURL
	}
	if parsed.Host == "" {
		return nil, ErrInvalidURL
	}
	host := parsed.Hostname()
	if host == "" {
		return nil, ErrInvalidURL
	}
	if strings.EqualFold(host, "localhost") {
		return nil, ErrInvalidURL
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() {
			return nil, ErrInvalidURL
		}
	}
	return parsed, nil
}
