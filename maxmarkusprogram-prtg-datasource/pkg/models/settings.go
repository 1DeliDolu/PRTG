package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type PluginSettings struct {
	Path      string                `json:"path"`
	CacheTime time.Duration         `json:"cacheTime"`
	Secrets   *SecretPluginSettings `json:"-"`
	Timezone  string                `json:"timeZone"`
}

type SecretPluginSettings struct {
	ApiKey string `json:"apiKey"`
}

func LoadPluginSettings(source backend.DataSourceInstanceSettings) (*PluginSettings, error) {
	settings := PluginSettings{}

	// Unmarshal settings from frontend
	err := json.Unmarshal(source.JSONData, &settings)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal PluginSettings json: %w", err)
	}

	// Only set default timezone if not provided by frontend
	if settings.Timezone == "" {
		settings.Timezone = "Europe/Berlin"
		backend.Logger.Debug("No timezone configured in settings, using default", "timezone", settings.Timezone)
	} else {
		backend.Logger.Debug("Using configured timezone from settings", "timezone", settings.Timezone)
	}

	// Validate timezone
	_, err = time.LoadLocation(settings.Timezone)
	if err != nil {
		backend.Logger.Warn("Invalid timezone in settings, using UTC",
			"configured_timezone", settings.Timezone,
			"error", err)
		settings.Timezone = "UTC"
	}

	settings.Secrets = loadSecretPluginSettings(source.DecryptedSecureJSONData)

	return &settings, nil
}

func loadSecretPluginSettings(source map[string]string) *SecretPluginSettings {
	return &SecretPluginSettings{
		ApiKey: source["apiKey"],
	}
}