// Copyright  OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apachereceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachereceiver"

import (
	"context"
	"net/url"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver/scraperhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachereceiver/internal/metadata"
)

const (
	typeStr   = "apache"
	stability = component.StabilityLevelBeta

	httpDefaultPort  = "80"
	httpsDefaultPort = "443"
)

// NewFactory creates a factory for apache receiver.
func NewFactory() component.ReceiverFactory {
	return component.NewReceiverFactory(
		typeStr,
		createDefaultConfig,
		component.WithMetricsReceiver(createMetricsReceiver, stability))
}

func createDefaultConfig() component.ReceiverConfig {
	return &Config{
		ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
			ReceiverSettings:   config.NewReceiverSettings(component.NewID(typeStr)),
			CollectionInterval: 10 * time.Second,
		},
		HTTPClientSettings: confighttp.HTTPClientSettings{
			Endpoint: defaultEndpoint,
			Timeout:  10 * time.Second,
		},
		Metrics: metadata.DefaultMetricsSettings(),
	}
}

func parseResourseAttributes(endpoint string) (string, string, error) {
	u, err := url.Parse(endpoint)
	serverName := u.Hostname()
	port := u.Port()

	if port == "" {
		if u.Scheme == "https" {
			port = httpsDefaultPort
		} else if u.Scheme == "http" {
			port = httpDefaultPort
		}
		// else: unknown scheme, leave port as empty string
	}

	return serverName, port, err
}

func createMetricsReceiver(
	_ context.Context,
	params component.ReceiverCreateSettings,
	rConf component.ReceiverConfig,
	consumer consumer.Metrics,
) (component.MetricsReceiver, error) {
	cfg := rConf.(*Config)
	serverName, port, err := parseResourseAttributes(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	ns := newApacheScraper(params, cfg, serverName, port)
	scraper, err := scraperhelper.NewScraper(typeStr, ns.scrape, scraperhelper.WithStart(ns.start))
	if err != nil {
		return nil, err
	}

	return scraperhelper.NewScraperControllerReceiver(
		&cfg.ScraperControllerSettings, params, consumer,
		scraperhelper.AddScraper(scraper),
	)
}
