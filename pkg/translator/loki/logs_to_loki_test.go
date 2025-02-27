// Copyright The OpenTelemetry Authors
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

package loki // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/loki"

import (
	"fmt"
	"testing"

	"github.com/grafana/loki/pkg/logproto"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
)

func TestLogsToLokiRequestWithGroupingByTenant(t *testing.T) {
	tests := []struct {
		name     string
		logs     plog.Logs
		expected map[string]PushRequest
	}{
		{
			name: "tenant from logs attributes",
			logs: func() plog.Logs {
				logs := plog.NewLogs()
				rl := logs.ResourceLogs().AppendEmpty()

				sl := rl.ScopeLogs().AppendEmpty()
				logRecord := sl.LogRecords().AppendEmpty()
				logRecord.Attributes().PutStr(hintTenant, "tenant.id")
				logRecord.Attributes().PutStr("tenant.id", "1")
				logRecord.Attributes().PutInt("http.status", 200)

				sl = rl.ScopeLogs().AppendEmpty()
				logRecord = sl.LogRecords().AppendEmpty()
				logRecord.Attributes().PutStr(hintTenant, "tenant.id")
				logRecord.Attributes().PutStr("tenant.id", "2")
				logRecord.Attributes().PutInt("http.status", 200)

				return logs
			}(),
			expected: map[string]PushRequest{
				"1": {
					PushRequest: &logproto.PushRequest{
						Streams: []logproto.Stream{
							{
								Labels: `{exporter="OTLP", tenant.id="1"}`,
								Entries: []logproto.Entry{
									{
										Line: `{"attributes":{"http.status":200}}`,
									},
								}},
						},
					},
				},
				"2": {
					PushRequest: &logproto.PushRequest{
						Streams: []logproto.Stream{
							{
								Labels: `{exporter="OTLP", tenant.id="2"}`,
								Entries: []logproto.Entry{
									{
										Line: `{"attributes":{"http.status":200}}`,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "tenant from resource attributes",
			logs: func() plog.Logs {
				logs := plog.NewLogs()
				rl := logs.ResourceLogs().AppendEmpty()
				rl.Resource().Attributes().PutStr(hintTenant, "tenant.id")
				rl.Resource().Attributes().PutStr("tenant.id", "11")

				sl := rl.ScopeLogs().AppendEmpty()
				logRecord := sl.LogRecords().AppendEmpty()
				logRecord.Attributes().PutInt("http.status", 200)

				rl = logs.ResourceLogs().AppendEmpty()
				rl.Resource().Attributes().PutStr(hintTenant, "tenant.id")
				rl.Resource().Attributes().PutStr("tenant.id", "12")

				sl = rl.ScopeLogs().AppendEmpty()
				logRecord = sl.LogRecords().AppendEmpty()
				logRecord.Attributes().PutInt("http.status", 200)

				return logs
			}(),
			expected: map[string]PushRequest{
				"11": {
					PushRequest: &logproto.PushRequest{
						Streams: []logproto.Stream{
							{
								Labels: `{exporter="OTLP", tenant.id="11"}`,
								Entries: []logproto.Entry{
									{
										Line: `{"attributes":{"http.status":200}}`,
									},
								}},
						},
					},
				},
				"12": {
					PushRequest: &logproto.PushRequest{
						Streams: []logproto.Stream{
							{
								Labels: `{exporter="OTLP", tenant.id="12"}`,
								Entries: []logproto.Entry{
									{
										Line: `{"attributes":{"http.status":200}}`,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "tenant hint attribute is not found in resource and logs attributes",
			logs: func() plog.Logs {
				logs := plog.NewLogs()
				rl := logs.ResourceLogs().AppendEmpty()

				sl := rl.ScopeLogs().AppendEmpty()
				logRecord := sl.LogRecords().AppendEmpty()
				logRecord.Attributes().PutStr(hintTenant, "tenant.id")
				logRecord.Attributes().PutInt("http.status", 200)

				return logs
			}(),
			expected: map[string]PushRequest{
				"": {
					PushRequest: &logproto.PushRequest{
						Streams: []logproto.Stream{
							{
								Labels: `{exporter="OTLP"}`,
								Entries: []logproto.Entry{
									{
										Line: `{"attributes":{"http.status":200}}`,
									},
								}},
						},
					},
				},
			},
		},
		{
			name: "use tenant resource attributes if both logs and resource attributes provided",
			logs: func() plog.Logs {
				logs := plog.NewLogs()

				rl := logs.ResourceLogs().AppendEmpty()
				rl.Resource().Attributes().PutStr(hintTenant, "tenant.id")
				rl.Resource().Attributes().PutStr("tenant.id", "21")

				sl := rl.ScopeLogs().AppendEmpty()
				logRecord := sl.LogRecords().AppendEmpty()
				logRecord.Attributes().PutStr(hintTenant, "tenant.id")
				logRecord.Attributes().PutStr("tenant.id", "31")
				logRecord.Attributes().PutInt("http.status", 200)

				rl = logs.ResourceLogs().AppendEmpty()
				rl.Resource().Attributes().PutStr(hintTenant, "tenant.id")
				rl.Resource().Attributes().PutStr("tenant.id", "22")

				sl = rl.ScopeLogs().AppendEmpty()
				logRecord = sl.LogRecords().AppendEmpty()
				logRecord.Attributes().PutStr(hintTenant, "tenant.id")
				logRecord.Attributes().PutStr("tenant.id", "32")
				logRecord.Attributes().PutInt("http.status", 200)

				return logs
			}(),
			expected: map[string]PushRequest{
				"21": {
					PushRequest: &logproto.PushRequest{
						Streams: []logproto.Stream{
							{
								Labels: `{exporter="OTLP", tenant.id="21"}`,
								Entries: []logproto.Entry{
									{
										Line: `{"attributes":{"http.status":200}}`,
									},
								}},
						},
					},
				},
				"22": {
					PushRequest: &logproto.PushRequest{
						Streams: []logproto.Stream{
							{
								Labels: `{exporter="OTLP", tenant.id="22"}`,
								Entries: []logproto.Entry{
									{
										Line: `{"attributes":{"http.status":200}}`,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := LogsToLokiRequests(tt.logs)

			for tenant, request := range requests {
				want, ok := tt.expected[tenant]
				assert.Equal(t, ok, true)

				streams := request.Streams
				for s := 0; s < len(streams); s++ {
					gotStream := request.Streams[s]
					wantStream := want.Streams[s]

					assert.Equal(t, wantStream.Labels, gotStream.Labels)
					for e := 0; e < len(gotStream.Entries); e++ {
						assert.Equal(t, wantStream.Entries[e].Line, gotStream.Entries[e].Line)
					}
				}
			}
		})
	}
}

func TestLogsToLokiRequestWithoutTenant(t *testing.T) {
	testCases := []struct {
		desc           string
		hints          map[string]interface{}
		attrs          map[string]interface{}
		res            map[string]interface{}
		severity       plog.SeverityNumber
		levelAttribute string
		expectedLabel  string
		expectedLines  []string
	}{
		{
			desc: "with attribute to label and regular attribute",
			attrs: map[string]interface{}{
				"host.name":   "guarana",
				"http.status": 200,
			},
			hints: map[string]interface{}{
				hintAttributes: "host.name",
			},
			expectedLabel: `{exporter="OTLP", host.name="guarana"}`,
			expectedLines: []string{
				`{"traceid":"01000000000000000000000000000000","attributes":{"http.status":200}}`,
				`{"traceid":"02000000000000000000000000000000","attributes":{"http.status":200}}`,
				`{"traceid":"03000000000000000000000000000000","attributes":{"http.status":200}}`,
			},
		},
		{
			desc: "with resource to label and regular resource",
			res: map[string]interface{}{
				"host.name": "guarana",
				"region.az": "eu-west-1a",
			},
			hints: map[string]interface{}{
				hintResources: "host.name",
			},
			expectedLabel: `{exporter="OTLP", host.name="guarana"}`,
			expectedLines: []string{
				`{"traceid":"01000000000000000000000000000000","resources":{"region.az":"eu-west-1a"}}`,
				`{"traceid":"02000000000000000000000000000000","resources":{"region.az":"eu-west-1a"}}`,
				`{"traceid":"03000000000000000000000000000000","resources":{"region.az":"eu-west-1a"}}`,
			},
		},
		{
			desc: "with logfmt format",
			attrs: map[string]interface{}{
				"host.name":   "guarana",
				"http.status": 200,
			},
			hints: map[string]interface{}{
				hintAttributes: "host.name",
				hintFormat:     formatLogfmt,
			},
			expectedLabel: `{exporter="OTLP", host.name="guarana"}`,
			expectedLines: []string{
				`traceID=01000000000000000000000000000000 attribute_http.status=200`,
				`traceID=02000000000000000000000000000000 attribute_http.status=200`,
				`traceID=03000000000000000000000000000000 attribute_http.status=200`,
			},
		},
		{
			desc:          "with severity to label",
			severity:      plog.SeverityNumberDebug4,
			expectedLabel: `{exporter="OTLP", level="DEBUG4"}`,
			expectedLines: []string{
				`{"traceid":"01000000000000000000000000000000"}`,
				`{"traceid":"02000000000000000000000000000000"}`,
				`{"traceid":"03000000000000000000000000000000"}`,
			},
		},
		{
			desc:           "with severity, already existing level",
			severity:       plog.SeverityNumberDebug4,
			levelAttribute: "dummy",
			expectedLabel:  `{exporter="OTLP", level="dummy"}`,
			expectedLines: []string{
				`{"traceid":"01000000000000000000000000000000"}`,
				`{"traceid":"02000000000000000000000000000000"}`,
				`{"traceid":"03000000000000000000000000000000"}`,
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			// prepare
			ld := plog.NewLogs()
			ld.ResourceLogs().AppendEmpty()
			for i := 0; i < 3; i++ {
				ld.ResourceLogs().At(0).ScopeLogs().AppendEmpty()
				ld.ResourceLogs().At(0).ScopeLogs().At(i).LogRecords().AppendEmpty()
				ld.ResourceLogs().At(0).ScopeLogs().At(i).LogRecords().At(0).SetTraceID([16]byte{byte(i + 1)})
				ld.ResourceLogs().At(0).ScopeLogs().At(i).LogRecords().At(0).SetSeverityNumber(tt.severity)
				if len(tt.levelAttribute) > 0 {
					ld.ResourceLogs().At(0).ScopeLogs().At(i).LogRecords().At(0).Attributes().PutStr(levelAttributeName, tt.levelAttribute)
				}
			}

			if len(tt.res) > 0 {
				assert.NoError(t, ld.ResourceLogs().At(0).Resource().Attributes().FromRaw(tt.res))
			}

			rlogs := ld.ResourceLogs()
			for i := 0; i < rlogs.Len(); i++ {
				slogs := rlogs.At(i).ScopeLogs()
				for j := 0; j < slogs.Len(); j++ {
					logs := slogs.At(j).LogRecords()
					for k := 0; k < logs.Len(); k++ {
						log := logs.At(k)
						if len(tt.attrs) > 0 {
							assert.NoError(t, log.Attributes().FromRaw(tt.attrs))
						}
						for k, v := range tt.hints {
							log.Attributes().PutStr(k, fmt.Sprintf("%v", v))
						}
					}
				}
			}

			// test
			requests := LogsToLokiRequests(ld)
			assert.Len(t, requests, 1)
			request := requests[""]

			// verify
			assert.Empty(t, request.Report.Errors)
			assert.Equal(t, 0, request.Report.NumDropped)
			assert.Equal(t, ld.LogRecordCount(), request.Report.NumSubmitted)
			assert.Len(t, request.Streams, 1)
			assert.Equal(t, tt.expectedLabel, request.Streams[0].Labels)

			entries := request.Streams[0].Entries
			for i := 0; i < len(entries); i++ {
				assert.Equal(t, tt.expectedLines[i], entries[i].Line)
			}
		})
	}
}

func TestLogsToLoki(t *testing.T) {
	testCases := []struct {
		desc           string
		hints          map[string]interface{}
		attrs          map[string]interface{}
		res            map[string]interface{}
		severity       plog.SeverityNumber
		levelAttribute string
		expectedLabel  string
		expectedLines  []string
	}{
		{
			desc: "with attribute to label and regular attribute",
			attrs: map[string]interface{}{
				"host.name":   "guarana",
				"http.status": 200,
			},
			hints: map[string]interface{}{
				hintAttributes: "host.name",
			},
			expectedLabel: `{exporter="OTLP", host.name="guarana"}`,
			expectedLines: []string{
				`{"traceid":"01020304000000000000000000000000","attributes":{"http.status":200}}`,
				`{"traceid":"01020304050000000000000000000000","attributes":{"http.status":200}}`,
				`{"traceid":"01020304050600000000000000000000","attributes":{"http.status":200}}`,
			},
		},
		{
			desc: "with resource to label and regular resource",
			res: map[string]interface{}{
				"host.name": "guarana",
				"region.az": "eu-west-1a",
			},
			hints: map[string]interface{}{
				hintResources: "host.name",
			},
			expectedLabel: `{exporter="OTLP", host.name="guarana"}`,
			expectedLines: []string{
				`{"traceid":"01020304000000000000000000000000","resources":{"region.az":"eu-west-1a"}}`,
				`{"traceid":"01020304050000000000000000000000","resources":{"region.az":"eu-west-1a"}}`,
				`{"traceid":"01020304050600000000000000000000","resources":{"region.az":"eu-west-1a"}}`,
			},
		},
		{
			desc: "with logfmt format",
			res: map[string]interface{}{
				"host.name": "guarana",
				"region.az": "eu-west-1a",
			},
			hints: map[string]interface{}{
				hintResources: "host.name",
				hintFormat:    formatLogfmt,
			},
			expectedLabel: `{exporter="OTLP", host.name="guarana"}`,
			expectedLines: []string{
				`traceID=01020304000000000000000000000000 resource_region.az=eu-west-1a`,
				`traceID=01020304050000000000000000000000 resource_region.az=eu-west-1a`,
				`traceID=01020304050600000000000000000000 resource_region.az=eu-west-1a`,
			},
		},
		{
			desc:          "with severity to label",
			severity:      plog.SeverityNumberDebug4,
			expectedLabel: `{exporter="OTLP", level="DEBUG4"}`,
			expectedLines: []string{
				`{"traceid":"01020304000000000000000000000000"}`,
				`{"traceid":"01020304050000000000000000000000"}`,
				`{"traceid":"01020304050600000000000000000000"}`,
			},
		},
		{
			desc:           "with severity, already existing level",
			severity:       plog.SeverityNumberDebug4,
			levelAttribute: "dummy",
			expectedLabel:  `{exporter="OTLP", level="dummy"}`,
			expectedLines: []string{
				`{"traceid":"01020304000000000000000000000000"}`,
				`{"traceid":"01020304050000000000000000000000"}`,
				`{"traceid":"01020304050600000000000000000000"}`,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// prepare
			ld := plog.NewLogs()
			ld.ResourceLogs().AppendEmpty()
			ld.ResourceLogs().At(0).ScopeLogs().AppendEmpty()
			ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().AppendEmpty()
			ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().AppendEmpty()
			ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().AppendEmpty()
			ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).SetSeverityNumber(tC.severity)
			ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).SetTraceID(pcommon.TraceID([16]byte{1, 2, 3, 4}))
			ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(1).SetSeverityNumber(tC.severity)
			ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(1).SetTraceID(pcommon.TraceID([16]byte{1, 2, 3, 4, 5}))
			ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(2).SetSeverityNumber(tC.severity)
			ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(2).SetTraceID(pcommon.TraceID([16]byte{1, 2, 3, 4, 5, 6}))

			// copy the attributes from the test case to the log entry
			if len(tC.attrs) > 0 {
				assert.NoError(t, ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Attributes().FromRaw(tC.attrs))
				assert.NoError(t, ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(1).Attributes().FromRaw(tC.attrs))
				assert.NoError(t, ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(2).Attributes().FromRaw(tC.attrs))
			}
			if len(tC.res) > 0 {
				assert.NoError(t, ld.ResourceLogs().At(0).Resource().Attributes().FromRaw(tC.res))
			}
			if len(tC.levelAttribute) > 0 {
				ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Attributes().PutStr(levelAttributeName, tC.levelAttribute)
				ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(1).Attributes().PutStr(levelAttributeName, tC.levelAttribute)
				ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(2).Attributes().PutStr(levelAttributeName, tC.levelAttribute)
			}

			// we can't use copy here, as the value (Value) will be used as string lookup later, so, we need to convert it to string now
			for k, v := range tC.hints {
				ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Attributes().PutStr(k, fmt.Sprintf("%v", v))
				ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(1).Attributes().PutStr(k, fmt.Sprintf("%v", v))
				ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(2).Attributes().PutStr(k, fmt.Sprintf("%v", v))
			}

			// test
			pushRequest, report := LogsToLoki(ld)
			entries := pushRequest.Streams[0].Entries

			var entriesLines []string
			for i := 0; i < len(entries); i++ {
				entriesLines = append(entriesLines, entries[i].Line)
			}

			// actualPushRequest is populated within the test http server, we check it here as assertions are better done at the
			// end of the test function
			assert.Empty(t, report.Errors)
			assert.Equal(t, 0, report.NumDropped)
			assert.Equal(t, ld.LogRecordCount(), report.NumSubmitted)
			assert.Len(t, pushRequest.Streams, 1)
			assert.Equal(t, tC.expectedLabel, pushRequest.Streams[0].Labels)
			assert.Len(t, entries, ld.LogRecordCount())
			assert.ElementsMatch(t, tC.expectedLines, entriesLines)
		})
	}
}
