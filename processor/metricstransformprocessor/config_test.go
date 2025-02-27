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

package metricstransformprocessor

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		configFile string
		id         component.ID
		expected   component.ProcessorConfig
	}{
		{
			configFile: "config_full.yaml",
			id:         component.NewID(typeStr),
			expected: &Config{
				ProcessorSettings: config.NewProcessorSettings(component.NewID(typeStr)),
				Transforms: []Transform{
					{
						MetricIncludeFilter: FilterConfig{
							Include:   "name",
							MatchType: "",
						},
						Action:  "update",
						NewName: "new_name",
					},
				},
			},
		},
		{
			configFile: "config_full.yaml",
			id:         component.NewIDWithName(typeStr, "multiple"),
			expected: &Config{
				ProcessorSettings: config.NewProcessorSettings(component.NewID(typeStr)),
				Transforms: []Transform{
					{
						MetricIncludeFilter: FilterConfig{
							Include:   "name1",
							MatchType: "strict",
						},
						Action:  "insert",
						NewName: "new_name",
						Operations: []Operation{
							{
								Action:   "add_label",
								NewLabel: "my_label",
								NewValue: "my_value",
							},
						},
					},
					{
						MetricIncludeFilter: FilterConfig{
							Include:   "new_name",
							MatchType: "strict",
							MatchLabels: map[string]string{
								"my_label": "my_value",
							},
						},
						Action:  "insert",
						NewName: "new_name_copy_1",
					},
					{
						MetricIncludeFilter: FilterConfig{
							Include:   "new_name",
							MatchType: "regexp",
							MatchLabels: map[string]string{
								"my_label": ".*label",
							},
						},
						Action:  "insert",
						NewName: "new_name_copy_2",
					},
					{
						MetricIncludeFilter: FilterConfig{
							Include:   "name2",
							MatchType: "",
						},
						Action: "update",
						Operations: []Operation{
							{
								Action:   "update_label",
								Label:    "label",
								NewLabel: "new_label_key",
								ValueActions: []ValueAction{
									{Value: "label1", NewValue: "new_label1"},
								},
							},
							{
								Action:          "aggregate_labels",
								LabelSet:        []string{"new_label1", "label2"},
								AggregationType: "sum",
							},
							{
								Action:           "aggregate_label_values",
								Label:            "new_label1",
								AggregationType:  "sum",
								AggregatedValues: []string{"value1", "value2"},
								NewValue:         "new_value",
							},
						},
					},
					{
						MetricIncludeFilter: FilterConfig{
							Include:   "name3",
							MatchType: "strict",
						},
						Action: "update",
						Operations: []Operation{
							{
								Action:     "delete_label_value",
								Label:      "my_label",
								LabelValue: "delete_me",
							},
						},
					},
					{
						MetricIncludeFilter: FilterConfig{
							Include:   "^regexp (?P<my_label>.*)$",
							MatchType: "regexp",
						},
						Action:       "combine",
						NewName:      "combined_metric_name",
						SubmatchCase: "lower",
					},
					{
						MetricIncludeFilter: FilterConfig{
							Include:   "name2",
							MatchType: "strict",
						},
						Action:              "group",
						GroupResourceLabels: map[string]string{"metric_group": "2"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			cm, err := confmaptest.LoadConf(filepath.Join("testdata", tt.configFile))
			require.NoError(t, err)

			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, component.UnmarshalProcessorConfig(sub, cfg))

			assert.NoError(t, component.ValidateConfig(cfg))
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
