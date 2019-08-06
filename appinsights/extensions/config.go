package extensions

import "github.com/microsoft/ApplicationInsights-Go/appinsights"

type InsightsConfig struct {
	Role                   string
	Version                string
	TelemetryConfiguration *appinsights.TelemetryConfiguration
}

func NewInsighsConfig(ikey, role, version string) *InsightsConfig {
	aiCfg := &InsightsConfig{}
	aiCfg.Role = role
	aiCfg.Version = version
	aiCfg.TelemetryConfiguration = appinsights.NewTelemetryConfiguration(ikey)
	return aiCfg
}

func createTelemetryClient(config *InsightsConfig) appinsights.TelemetryClient {
	client := appinsights.NewTelemetryClientFromConfig(config.TelemetryConfiguration)

	if len(config.Role) > 0 {
		client.Context().Tags.Cloud().SetRole(config.Role)
	}

	if len(config.Version) > 0 {
		client.Context().Tags.Application().SetVer(config.Version)
	}

	return client
}
