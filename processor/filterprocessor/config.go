// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package filterprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"

import (
	"errors"
	"fmt"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter/filterconfig"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter/filtermetric"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter/filterset"
	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter/filterset/regexp"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottldatapoint"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlmetric"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlspan"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottlspanevent"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor/internal/common"
)

// Config defines configuration for Resource processor.
type Config struct {
	config.ProcessorSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct

	Metrics MetricFilters `mapstructure:"metrics"`

	Logs LogFilters `mapstructure:"logs"`

	Spans SpanFilters `mapstructure:"spans"`

	Traces TraceFilters `mapstructure:"traces"`
}

// MetricFilters filters by Metric properties.
type MetricFilters struct {
	// Include match properties describe metrics that should be included in the Collector Service pipeline,
	// all other metrics should be dropped from further processing.
	// If both Include and Exclude are specified, Include filtering occurs first.
	Include *filtermetric.MatchProperties `mapstructure:"include"`

	// Exclude match properties describe metrics that should be excluded from the Collector Service pipeline,
	// all other metrics should be included.
	// If both Include and Exclude are specified, Include filtering occurs first.
	Exclude *filtermetric.MatchProperties `mapstructure:"exclude"`

	// RegexpConfig specifies options for the Regexp match type
	RegexpConfig *regexp.Config `mapstructure:"regexp"`

	// MetricConditions is a list of OTTL conditions for an ottlmetric context.
	// If any condition resolves to true, the metric will be dropped.
	// Supports `and`, `or`, and `()`
	MetricConditions []string `mapstructure:"metric"`

	// DataPointConditions is a list of OTTL conditions for an ottldatapoint context.
	// If any condition resolves to true, the datapoint will be dropped.
	// Supports `and`, `or`, and `()`
	DataPointConditions []string `mapstructure:"datapoint"`
}

// SpanFilters filters by Span attributes and various other fields, Regexp config is per matcher
type SpanFilters struct {
	// Include match properties describe spans that should be included in the Collector Service pipeline,
	// all other spans should be dropped from further processing.
	// If both Include and Exclude are specified, Include filtering occurs first.
	Include *filterconfig.MatchProperties `mapstructure:"include"`

	// Exclude match properties describe spans that should be excluded from the Collector Service pipeline,
	// all other spans should be included.
	// If both Include and Exclude are specified, Include filtering occurs first.
	Exclude *filterconfig.MatchProperties `mapstructure:"exclude"`
}

// TraceFilters filters by OTTL conditions
type TraceFilters struct {
	// SpanConditions is a list of OTTL conditions for an ottlspan context.
	// If any condition resolves to true, the span will be dropped.
	// Supports `and`, `or`, and `()`
	SpanConditions []string `mapstructure:"span"`

	// SpanEventConditions is a list of OTTL conditions for an ottlspanevent context.
	// If any condition resolves to true, the span event will be dropped.
	// Supports `and`, `or`, and `()`
	SpanEventConditions []string `mapstructure:"spanevent"`
}

// LogFilters filters by Log properties.
type LogFilters struct {
	// Include match properties describe logs that should be included in the Collector Service pipeline,
	// all other logs should be dropped from further processing.
	// If both Include and Exclude are specified, Include filtering occurs first.
	Include *LogMatchProperties `mapstructure:"include"`
	// Exclude match properties describe logs that should be excluded from the Collector Service pipeline,
	// all other logs should be included.
	// If both Include and Exclude are specified, Include filtering occurs first.
	Exclude *LogMatchProperties `mapstructure:"exclude"`

	// LogConditions is a list of OTTL conditions for an ottllog context.
	// If any condition resolves to true, the log event will be dropped.
	// Supports `and`, `or`, and `()`
	LogConditions []string `mapstructure:"log_record"`
}

// LogMatchType specifies the strategy for matching against `plog.Log`s.
type LogMatchType string

// These are the MatchTypes that users can specify for filtering
// `plog.Log`s.
const (
	Strict = LogMatchType(filterset.Strict)
	Regexp = LogMatchType(filterset.Regexp)
)

var severityToNumber = map[string]plog.SeverityNumber{
	"1":      plog.SeverityNumberTrace,
	"2":      plog.SeverityNumberTrace2,
	"3":      plog.SeverityNumberTrace3,
	"4":      plog.SeverityNumberTrace4,
	"5":      plog.SeverityNumberDebug,
	"6":      plog.SeverityNumberDebug2,
	"7":      plog.SeverityNumberDebug3,
	"8":      plog.SeverityNumberDebug4,
	"9":      plog.SeverityNumberInfo,
	"10":     plog.SeverityNumberInfo2,
	"11":     plog.SeverityNumberInfo3,
	"12":     plog.SeverityNumberInfo4,
	"13":     plog.SeverityNumberWarn,
	"14":     plog.SeverityNumberWarn2,
	"15":     plog.SeverityNumberWarn3,
	"16":     plog.SeverityNumberWarn4,
	"17":     plog.SeverityNumberError,
	"18":     plog.SeverityNumberError2,
	"19":     plog.SeverityNumberError3,
	"20":     plog.SeverityNumberError4,
	"21":     plog.SeverityNumberFatal,
	"22":     plog.SeverityNumberFatal2,
	"23":     plog.SeverityNumberFatal3,
	"24":     plog.SeverityNumberFatal4,
	"TRACE":  plog.SeverityNumberTrace,
	"TRACE2": plog.SeverityNumberTrace2,
	"TRACE3": plog.SeverityNumberTrace3,
	"TRACE4": plog.SeverityNumberTrace4,
	"DEBUG":  plog.SeverityNumberDebug,
	"DEBUG2": plog.SeverityNumberDebug2,
	"DEBUG3": plog.SeverityNumberDebug3,
	"DEBUG4": plog.SeverityNumberDebug4,
	"INFO":   plog.SeverityNumberInfo,
	"INFO2":  plog.SeverityNumberInfo2,
	"INFO3":  plog.SeverityNumberInfo3,
	"INFO4":  plog.SeverityNumberInfo4,
	"WARN":   plog.SeverityNumberWarn,
	"WARN2":  plog.SeverityNumberWarn2,
	"WARN3":  plog.SeverityNumberWarn3,
	"WARN4":  plog.SeverityNumberWarn4,
	"ERROR":  plog.SeverityNumberError,
	"ERROR2": plog.SeverityNumberError2,
	"ERROR3": plog.SeverityNumberError3,
	"ERROR4": plog.SeverityNumberError4,
	"FATAL":  plog.SeverityNumberFatal,
	"FATAL2": plog.SeverityNumberFatal2,
	"FATAL3": plog.SeverityNumberFatal3,
	"FATAL4": plog.SeverityNumberFatal4,
}

var errInvalidSeverity = errors.New("not a valid severity")

// logSeverity is a type that represents a SeverityNumber as a string
type logSeverity string

// validate checks that the logSeverity is valid
func (l logSeverity) validate() error {
	if l == "" {
		// No severity specified, which means to ignore this field.
		return nil
	}

	capsSeverity := strings.ToUpper(string(l))
	if _, ok := severityToNumber[capsSeverity]; !ok {
		return fmt.Errorf("'%s' is not a valid severity: %w", string(l), errInvalidSeverity)
	}
	return nil
}

// severityNumber returns the severity number that the logSeverity represents
func (l logSeverity) severityNumber() plog.SeverityNumber {
	capsSeverity := strings.ToUpper(string(l))
	return severityToNumber[capsSeverity]
}

// LogMatchProperties specifies the set of properties in a log to match against and the
// type of string pattern matching to use.
type LogMatchProperties struct {
	// LogMatchType specifies the type of matching desired
	LogMatchType LogMatchType `mapstructure:"match_type"`

	// ResourceAttributes defines a list of possible resource attributes to match logs against.
	// A match occurs if any resource attribute matches all expressions in this given list.
	ResourceAttributes []filterconfig.Attribute `mapstructure:"resource_attributes"`

	// RecordAttributes defines a list of possible record attributes to match logs against.
	// A match occurs if any record attribute matches at least one expression in this given list.
	RecordAttributes []filterconfig.Attribute `mapstructure:"record_attributes"`

	// SeverityTexts is a list of strings that the LogRecord's severity text field must match
	// against.
	SeverityTexts []string `mapstructure:"severity_texts"`

	// SeverityNumberProperties defines how to match against a log record's SeverityNumber, if defined.
	SeverityNumberProperties *LogSeverityNumberMatchProperties `mapstructure:"severity_number"`

	// LogBodies is a list of strings that the LogRecord's body field must match
	// against.
	LogBodies []string `mapstructure:"bodies"`
}

// validate checks that the LogMatchProperties is valid
func (lmp LogMatchProperties) validate() error {
	if lmp.SeverityNumberProperties != nil {
		return lmp.SeverityNumberProperties.validate()
	}
	return nil
}

// isEmpty returns true if the properties is "empty" (meaning, there are no filters specified)
// if this is the case, the filter should be ignored.
func (lmp LogMatchProperties) isEmpty() bool {
	return len(lmp.ResourceAttributes) == 0 && len(lmp.RecordAttributes) == 0 &&
		len(lmp.SeverityTexts) == 0 && len(lmp.LogBodies) == 0 &&
		lmp.SeverityNumberProperties == nil
}

// matchProperties converts the LogMatchProperties to a corresponding filterconfig.MatchProperties
func (lmp LogMatchProperties) matchProperties() *filterconfig.MatchProperties {
	mp := &filterconfig.MatchProperties{
		Config: filterset.Config{
			MatchType: filterset.MatchType(lmp.LogMatchType),
		},
		Resources:        lmp.ResourceAttributes,
		Attributes:       lmp.RecordAttributes,
		LogSeverityTexts: lmp.SeverityTexts,
		LogBodies:        lmp.LogBodies,
	}

	// Include SeverityNumberProperties if defined
	if lmp.SeverityNumberProperties != nil {
		mp.LogSeverityNumber = &filterconfig.LogSeverityNumberMatchProperties{
			Min:            lmp.SeverityNumberProperties.Min.severityNumber(),
			MatchUndefined: lmp.SeverityNumberProperties.MatchUndefined,
		}
	}

	return mp
}

type LogSeverityNumberMatchProperties struct {
	// Min is the minimum severity needed for the log record to match.
	// This corresponds to the short names specified here:
	// https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/logs/data-model.md#displaying-severity
	// this field is case-insensitive ("INFO" == "info")
	Min logSeverity `mapstructure:"min"`

	// MatchUndefined lets logs records with "unknown" severity match.
	// If MinSeverity is not set, this field is ignored, as fields are not matched based on severity.
	MatchUndefined bool `mapstructure:"match_undefined"`
}

// validate checks that the LogMatchProperties is valid
func (lmp LogSeverityNumberMatchProperties) validate() error {
	return lmp.Min.validate()
}

var _ component.ProcessorConfig = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	if (cfg.Traces.SpanConditions != nil || cfg.Traces.SpanEventConditions != nil) && (cfg.Spans.Include != nil || cfg.Spans.Exclude != nil) {
		return fmt.Errorf("cannot use ottl conditions and include/exclude for spans at the same time")
	}
	if (cfg.Metrics.MetricConditions != nil || cfg.Metrics.DataPointConditions != nil) && (cfg.Metrics.Include != nil || cfg.Metrics.Exclude != nil) {
		return fmt.Errorf("cannot use ottl conditions and include/exclude for metrics at the same time")
	}
	if cfg.Logs.LogConditions != nil && (cfg.Logs.Include != nil || cfg.Logs.Exclude != nil) {
		return fmt.Errorf("cannot use ottl conditions and include/exclude for logs at the same time")
	}

	var errors error

	if cfg.Traces.SpanConditions != nil {
		spanp := ottlspan.NewParser(common.Functions[ottlspan.TransformContext](), component.TelemetrySettings{Logger: zap.NewNop()})
		_, err := spanp.ParseStatements(common.PrepareConditionForParsing(cfg.Traces.SpanConditions))
		errors = multierr.Append(errors, err)
	}

	if cfg.Traces.SpanEventConditions != nil {
		spaneventp := ottlspanevent.NewParser(common.Functions[ottlspanevent.TransformContext](), component.TelemetrySettings{Logger: zap.NewNop()})
		_, err := spaneventp.ParseStatements(common.PrepareConditionForParsing(cfg.Traces.SpanEventConditions))
		errors = multierr.Append(errors, err)
	}

	if cfg.Metrics.MetricConditions != nil {
		metricp := ottlmetric.NewParser(common.Functions[ottlmetric.TransformContext](), component.TelemetrySettings{Logger: zap.NewNop()})
		_, err := metricp.ParseStatements(common.PrepareConditionForParsing(cfg.Metrics.MetricConditions))
		errors = multierr.Append(errors, err)
	}

	if cfg.Metrics.DataPointConditions != nil {
		datapointp := ottldatapoint.NewParser(common.Functions[ottldatapoint.TransformContext](), component.TelemetrySettings{Logger: zap.NewNop()})
		_, err := datapointp.ParseStatements(common.PrepareConditionForParsing(cfg.Metrics.DataPointConditions))
		errors = multierr.Append(errors, err)
	}

	if cfg.Logs.LogConditions != nil {
		logp := ottllog.NewParser(common.Functions[ottllog.TransformContext](), component.TelemetrySettings{Logger: zap.NewNop()})
		_, err := logp.ParseStatements(common.PrepareConditionForParsing(cfg.Logs.LogConditions))
		errors = multierr.Append(errors, err)
	}

	if cfg.Logs.LogConditions != nil && cfg.Logs.Include != nil {
		errors = multierr.Append(errors, cfg.Logs.Include.validate())
	}

	if cfg.Logs.LogConditions != nil && cfg.Logs.Exclude != nil {
		errors = multierr.Append(errors, cfg.Logs.Exclude.validate())
	}

	return errors
}
