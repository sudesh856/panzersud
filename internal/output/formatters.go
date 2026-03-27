// Package output handles all --output flag formats for suddpanzer.
// Formats: text | json | csv | junit
// Drop-in addition — text and json behaviour is unchanged from Phase 1.
package output

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/sudesh856/suddpanzer/internal/report"
)

// WriteJSON writes a pretty-printed JSON summary.
// Same as your existing json.MarshalIndent block, but accepts any io.Writer.
func WriteJSON(w io.Writer, sum report.Summary) error {
	data, err := json.MarshalIndent(sum, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(data))
	return err
}

// WriteCSV writes a header line + one data line CSV.
// Append to a file across runs to build a trend log.
func WriteCSV(w io.Writer, sum report.Summary) error {
	header := "scenario_name,url,duration_secs,total_requests,avg_rps," +
		"p50_ms,p95_ms,p99_ms,error_count,error_rate_pct"

	row := fmt.Sprintf("%s,%s,%.2f,%d,%.2f,%d,%d,%d,%d,%.2f",
		sum.ScenarioName,
		sum.URL,
		sum.DurationSecs,
		sum.TotalRequests,
		sum.AvgRPS,
		sum.P50,
		sum.P95,
		sum.P99,
		sum.Errors,
		sum.ErrorRate,
	)

	_, err := fmt.Fprintf(w, "%s\n%s\n", header, row)
	return err
}

// ── JUnit XML ─────────────────────────────────────────────────────────────────

type junitTestSuites struct {
	XMLName    xml.Name         `xml:"testsuites"`
	Name       string           `xml:"name,attr"`
	Tests      int              `xml:"tests,attr"`
	Failures   int              `xml:"failures,attr"`
	Time       string           `xml:"time,attr"`
	TestSuites []junitTestSuite `xml:"testsuite"`
}

type junitTestSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Time      string          `xml:"time,attr"`
	TestCases []junitTestCase `xml:"testcase"`
}

type junitTestCase struct {
	XMLName   xml.Name      `xml:"testcase"`
	Name      string        `xml:"name,attr"`
	ClassName string        `xml:"classname,attr"`
	Time      string        `xml:"time,attr"`
	Failure   *junitFailure `xml:"failure,omitempty"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Text    string `xml:",chardata"`
}

// WriteJUnit writes a JUnit XML report.
// thresholdFailures is the slice of failure messages from scenario.Thresholds.Evaluate().
// Each failure becomes a failed test case visible in GitHub Actions' Tests tab.
// If thresholdFailures is empty, one synthetic "pass" test case is emitted.
func WriteJUnit(w io.Writer, sum report.Summary, thresholdFailures []string) error {
	var cases []junitTestCase

	if len(thresholdFailures) > 0 {
		for _, msg := range thresholdFailures {
			cases = append(cases, junitTestCase{
				Name:      msg,
				ClassName: "suddpanzer.threshold",
				Time:      "0.000",
				Failure: &junitFailure{
					Message: msg,
					Type:    "ThresholdFailure",
					Text:    msg,
				},
			})
		}
	} else {
		// No failures — emit one passing test case summarising the run.
		cases = append(cases, junitTestCase{
			Name: fmt.Sprintf("suddpanzer: %.1f rps, p99=%dms, errors=%.2f%%",
				sum.AvgRPS, sum.P99, sum.ErrorRate),
			ClassName: "suddpanzer.run",
			Time:      fmt.Sprintf("%.3f", sum.DurationSecs),
		})
	}

	failCount := 0
	for _, tc := range cases {
		if tc.Failure != nil {
			failCount++
		}
	}

	suites := junitTestSuites{
		Name:     "suddpanzer",
		Tests:    len(cases),
		Failures: failCount,
		Time:     fmt.Sprintf("%.3f", sum.DurationSecs),
		TestSuites: []junitTestSuite{{
			Name:      "suddpanzer",
			Tests:     len(cases),
			Failures:  failCount,
			Time:      fmt.Sprintf("%.3f", sum.DurationSecs),
			TestCases: cases,
		}},
	}

	out, err := xml.MarshalIndent(suites, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal junit xml: %w", err)
	}

	_, err = fmt.Fprintf(w, "%s\n%s\n", xml.Header, string(out))
	return err
}

// WriteSeparator writes the ─── divider line used in text summaries.
func WriteSeparator(w io.Writer) {
	fmt.Fprintln(w, strings.Repeat("-", 35))
}