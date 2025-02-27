// Copyright 2020, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dockerstatsreceiver

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/receiver/scraperhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver/internal/metadata"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id       component.ID
		expected component.ReceiverConfig
	}{
		{
			id:       component.NewIDWithName(typeStr, ""),
			expected: createDefaultConfig(),
		},
		{
			id: component.NewIDWithName(typeStr, "allsettings"),
			expected: &Config{
				ScraperControllerSettings: scraperhelper.ScraperControllerSettings{
					ReceiverSettings:   config.NewReceiverSettings(component.NewID(typeStr)),
					CollectionInterval: 2 * time.Second,
				},

				Endpoint:         "http://example.com/",
				Timeout:          20 * time.Second,
				DockerAPIVersion: 1.24,

				ProvidePerCoreCPUMetrics: true,
				ExcludedImages: []string{
					"undesired-container",
					"another-*-container",
				},

				ContainerLabelsToMetricLabels: map[string]string{
					"my.container.label":       "my-metric-label",
					"my.other.container.label": "my-other-metric-label",
				},

				EnvVarsToMetricLabels: map[string]string{
					"MY_ENVIRONMENT_VARIABLE":       "my-metric-label",
					"MY_OTHER_ENVIRONMENT_VARIABLE": "my-other-metric-label",
				},
				MetricsConfig: func() metadata.MetricsSettings {
					m := metadata.DefaultMetricsSettings()
					m.ContainerCPUUsageSystem = metadata.MetricSettings{
						Enabled: false,
					}
					m.ContainerMemoryTotalRss = metadata.MetricSettings{
						Enabled: true,
					}
					return m
				}(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
			require.NoError(t, err)

			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, component.UnmarshalReceiverConfig(sub, cfg))

			assert.NoError(t, component.ValidateConfig(cfg))
			if diff := cmp.Diff(tt.expected, cfg, cmpopts.IgnoreUnexported(config.ReceiverSettings{}, metadata.MetricSettings{})); diff != "" {
				t.Errorf("Config mismatch (-expected +actual):\n%s", diff)
			}
		})
	}
}

func TestValidateErrors(t *testing.T) {
	cfg := &Config{}
	assert.Equal(t, "endpoint must be specified", component.ValidateConfig(cfg).Error())

	cfg = &Config{Endpoint: "someEndpoint"}
	assert.Equal(t, "collection_interval must be a positive duration", component.ValidateConfig(cfg).Error())

	cfg = &Config{ScraperControllerSettings: scraperhelper.ScraperControllerSettings{CollectionInterval: 1 * time.Second}, Endpoint: "someEndpoint", DockerAPIVersion: 1.21}
	assert.Equal(t, "api_version must be at least 1.22", component.ValidateConfig(cfg).Error())
}
