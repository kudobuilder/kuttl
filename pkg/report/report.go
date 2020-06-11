package report

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"path/filepath"
	"time"
)

// The structs below define the report output useful in either json or xml format.  The xml format and structs
// are junit xml compliant.  A number of resources were used but https://www.ibm.com/support/knowledgecenter/SSQ2R2_9.1.1/com.ibm.rsar.analysis.codereview.cobol.doc/topics/cac_useresults_junit.html
// was very useful.  As well as:  https://www.onlinetool.io/xmltogo/

// KUTTL is different than junit testing in that the test steps could be useful to have a report on.  There could be value in
// having a TestCaseStep struct providing step details.  Sticking with the JUnit standard for now.

// Type defines the report.type of report to create
type Type string

const (
	// XML defines the xml Type
	XML Type = "xml"
	// JSON defines the json Type
	JSON Type = "json"
)

// Property are name/value pairs which can be provided in the report for things such as kuttl.version
type Property struct {
	Name  string `xml:"name,attr" json:"name"`
	Value string `xml:"value,attr" json:"value"`
}

// Properties defines the collection of properties
type Properties struct {
	Property []Property `xml:"property" json:"property,omitempty"`
}

// Failure defines a test failure
type Failure struct {
	// Text provides detailed information regarding failure.  It supports multi-line output.
	Text string `xml:",chardata" json:"text,omitempty"`
	// Message provides the summary of the failure
	Message string `xml:"message,attr" json:"message"`
	Type    string `xml:"type,attr" json:"type,omitempty"`
}

// Testcase is the finest grain level of reporting, it is the kuttl test (which contains steps)
type Testcase struct {
	// Classname is a junit thing, for kuttl it is the testsuite name
	Classname string `xml:"classname,attr" json:"classname"`
	// Name is the name of the test (folder of test if not redefined by the TestStep)
	Name string `xml:"name,attr" json:"name"`
	// Time is the elapsed time of the test (and all of it's steps)
	Time string `xml:"time,attr" json:"time"`
	// Assertions is the number of asserts and errors defined in the test
	Assertions int `xml:"assertions,attr" json:"assertions,omitempty"`
	// Failure defines a failure in this testcase
	Failure *Failure `xml:"failure" json:"failure,omitempty"`

	// start and end are not reported.  They are used to calc duration times for testcase and testsuite.
	start time.Time
	end   time.Time
}

// TestSuite is a collection of Testcase and is a summary of those details
type Testsuite struct {
	// Tests is the number of Testcases in the collection
	Tests int `xml:"tests,attr" json:"tests"`
	// Failures is the summary number of all failure in the collection testcases
	Failures int `xml:"failures,attr" json:"failures"`
	// Time is the duration of time for this Testsuite, this is tricky as tests run concurrently.
	// This is the elapse time between the start of the testsuite and the end of the latest testcase in the collection.
	Time string `xml:"time,attr" json:"time"`
	// Name is the kuttl test name
	Name string `xml:"name,attr" json:"name"`
	// Properties which are specific to this suite
	Properties *Properties `xml:"properties" json:"properties,omitempty"`
	// Testcase is a collection of test cases
	Testcase []*Testcase `xml:"testcase" json:"testcase,omitempty"`

	start time.Time
}

// Testsuites is a collection of Testsuite and defines the rollup summary of all stats.
type Testsuites struct {
	// XMLName is required to refine the name (or case of the name) in the root xml element.  Otherwise it adds no value and is ignored for json output.
	XMLName xml.Name `json:"-"`
	// Name is the name of the full set of tests which is possible to set in kuttl but is rarely used :)
	Name string `xml:"name,attr" json:"name"`
	// Tests is a summary value of the total number of tests for all testsuites
	Tests int `xml:"tests,attr" json:"tests"`
	// Failures is a summary value of the total number of failures for all testsuites
	Failures int `xml:"failures,attr" json:"failures"`
	// Time is the elapsed time of the entire suite of tests
	Time string `xml:"time,attr" json:"time"`
	// Properties which are for the entire set of tests
	Properties *Properties `xml:"properties" json:"properties,omitempty"`
	// Testsuite is a collection of test suites
	Testsuite []*Testsuite `xml:"testsuite" json:"testsuite,omitempty"`

	start time.Time
}

// NewSuiteCollection returns the address of a newly created TestSuites
func NewSuiteCollection(name string) *Testsuites {
	start := time.Now()
	return &Testsuites{XMLName: xml.Name{Local: "testsuites"}, Name: name, start: start}
}

// NewSuite returns the address of a newly created TestSuite
func NewSuite(name string) *Testsuite {
	start := time.Now()
	return &Testsuite{Name: name, start: start}
}

// NewCase returns the address of a newly create Testcase
func NewCase(name string) *Testcase {
	start := time.Now()
	return &Testcase{Name: name, start: start}
}

// NewFailure returns the address of a newly created Failure
func NewFailure(msg string, errs []error) *Failure {
	f := &Failure{Message: msg}

	// the mental debate... when there are more than 1 errors, the most common case is
	// an assert of yaml that is incorrect.  the first error has the diff and the second has the specific
	// error that is interesting.  The diff can be so long... and the second error added to a concat string gets buried
	// in the noise.  Seems better to just see the reason and have the user look at test stdout for the larger context if desired.
	if len(errs) > 0 {
		f.Text = errs[len(errs)-1].Error()
	}
	return f
}

// AddTestcase adds a testcase to a suite, providing stats and calculations to both
func (ts *Testsuite) AddTestcase(testcase *Testcase) {
	// this is needed to calc elapse time of testsuite in a async work
	testcase.end = time.Now()
	elapsed := time.Since(testcase.start)
	testcase.Time = elapsed.String()
	testcase.Classname = filepath.Base(ts.Name)

	ts.Testcase = append(ts.Testcase, testcase)
	ts.Tests++
	if testcase.Failure != nil {
		ts.Failures++
	}
}

// AddProperty adds a property to a testsuite
func (ts *Testsuite) AddProperty(property Property) {
	if ts.Properties == nil {
		ts.Properties = &Properties{Property: []Property{property}}
		return
	}
	if ts.Properties.Property == nil {
		ts.Properties.Property = []Property{property}
		return
	}
	ts.Properties.Property = append(ts.Properties.Property, property)
}

// AddTestSuite is a convenience method to add a testsuite to the collection in testsuites
func (ts *Testsuites) AddTestSuite(testsuite *Testsuite) {
	// testsuite is added prior to stat availability, stat management in the close of the testsuites
	ts.Testsuite = append(ts.Testsuite, testsuite)
}

// AddProperty adds a property to a testsuites
func (ts *Testsuites) AddProperty(property Property) {
	if ts.Properties == nil {
		ts.Properties = &Properties{Property: []Property{property}}
		return
	}
	if ts.Properties.Property == nil {
		ts.Properties.Property = []Property{property}
		return
	}
	ts.Properties.Property = append(ts.Properties.Property, property)
}

// Close closes the report and does all end stat calculations
func (ts *Testsuites) Close() {
	elapsed := time.Since(ts.start)
	ts.Time = elapsed.String()

	// async work makes this necessary (stats for each testsuite)
	for _, testsuite := range ts.Testsuite {
		elapsed = latestEnd(testsuite.start, testsuite.Testcase).Sub(testsuite.start)
		testsuite.Time = elapsed.String()

		ts.Tests += testsuite.Tests
		ts.Failures += testsuite.Failures
	}
}

// latestEnd provides the time of the latest end out of the collection of testcases
func latestEnd(start time.Time, testcases []*Testcase) time.Time {
	end := start
	for _, testcase := range testcases {
		if testcase.end.After(end) {
			end = testcase.end
		}
	}
	return end
}

// Report prints a report for TestSuites to the directory.  ftype == json | xml
func (ts *Testsuites) Report(dir string, ftype Type) error {
	ts.Close()
	// don't print if there is nothing
	if len(ts.Testsuite) == 0 {
		return nil
	}
	switch ftype {
	case XML:
		return writeXMLReport(dir, ts)
	case JSON:
		fallthrough
	default:
		return writeJSONReport(dir, ts)
	}
}

// NewSuite creates and assigns a TestSuite to the TestSuites (then returns the suite)
func (ts *Testsuites) NewSuite(name string) *Testsuite {
	suite := NewSuite(name)
	ts.AddTestSuite(suite)
	return suite
}

func writeXMLReport(dir string, ts *Testsuites) error {
	file := filepath.Join(dir, "kuttl-report.xml")
	xDoc, err := xml.MarshalIndent(ts, " ", "  ")
	if err != nil {
		return err
	}
	xmlStr := string(xDoc)
	return ioutil.WriteFile(file, []byte(xmlStr), 0644)
}

func writeJSONReport(dir string, ts *Testsuites) error {
	file := filepath.Join(dir, "kuttl-report.json")
	jDoc, err := json.MarshalIndent(ts, " ", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(file, jDoc, 0644)
}
