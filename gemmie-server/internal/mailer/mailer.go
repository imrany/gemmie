package mailer

import (
	"fmt"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Email    string
}

// EmailData represent an email's content
type EmailData struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

// ValidateConfig checks if the SMTP configurations are valid
func ValidateConfig(config SMTPConfig) error {
	if config.Host == "" {
		return fmt.Errorf("SMTP Host required")
	}
	if config.Port == 0 {
		return fmt.Errorf("SMTP Port required")
	}
	if config.Email == "" {
		return fmt.Errorf("SMTP Email require")
	}
	if config.Password == "" {
		return fmt.Errorf("SMTP Password required")
	}
	if config.Username == "" {
		return fmt.Errorf("SMTP Username required")
	}
	return nil
}

