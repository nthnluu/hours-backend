package config

import (
	"log"
	"time"
)

var Config *ServerConfig

// ServerConfig is a struct that contains configuration values for the server.
type ServerConfig struct {
	// AllowedOrigins is a list of URLs that the server will accept requests from.
	AllowedOrigins []string
	// AllowedEmailDomains is a list of email domains that the server will allow account registrations from. If empty,
	// all domains will be allowed.
	AllowedEmailDomains []string
	// SessionCookieName is the name to use for the session cookie.
	SessionCookieName string
	// SessionCookieExpiration is the amount of time a session cookie is valid. Max 5 days.
	SessionCookieExpiration time.Duration
	// Port is the port the server should run on.
	Port int
}

func DefaultConfig() *ServerConfig {
	return &ServerConfig{
		AllowedOrigins:          []string{"http://localhost:3000"},
		AllowedEmailDomains:     []string{"brown.edu"},
		SessionCookieName:       "signmeup-session",
		SessionCookieExpiration: time.Hour * 24 * 14,
		Port:                    8080,
	}
}

func init() {
	log.Println("🙂️ No configuration provided. Using the default configuration.")
	Config = DefaultConfig()
}
