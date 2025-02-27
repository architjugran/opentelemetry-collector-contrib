// Copyright The OpenTelemetry Authors
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

package filtermetric // import "github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter/filtermetric"

import (
	"go.opentelemetry.io/collector/pdata/pmetric"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter/filterexpr"
)

type exprMatcher struct {
	matchers []*filterexpr.Matcher
}

func newExprMatcher(expressions []string) (*exprMatcher, error) {
	m := &exprMatcher{}
	for _, expression := range expressions {
		matcher, err := filterexpr.NewMatcher(expression)
		if err != nil {
			return nil, err
		}
		m.matchers = append(m.matchers, matcher)
	}
	return m, nil
}

func (m *exprMatcher) MatchMetric(metric pmetric.Metric) (bool, error) {
	for _, matcher := range m.matchers {
		matched, err := matcher.MatchMetric(metric)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}
