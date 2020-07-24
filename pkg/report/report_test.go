package report

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var updateGolden = flag.Bool("update", false, "update .golden files")

func TestXML(t *testing.T) {
	goldenXML := "report.xml"
	goldenJSON := "report.json"

	tcase := &Testcase{
		Classname: "pkg1.test.test_things",
		Name:      "test_params_func:2",
		Time:      "",
		Failure: &Failure{
			Text: `Traceback (most recent call last):
  File "nose2/plugins/loader/parameters.py", line 162, in func
    return obj(*argSet)
  File "nose2/tests/functional/support/scenario/tests_in_package/pkg1/test/test_things.py", line 64, in test_params_func
    assert a == 1
AssertionError`,
			Message: "test failure",
		},
	}
	suite := &Testsuite{
		Tests:    9,
		Failures: 1,
		Time:     "",
		Name:     "github.com/kubebuilder/kuttl/pkg/version",
		Testcase: []*Testcase{
			tcase,
		},
	}

	suites := Testsuites{
		XMLName:  xml.Name{Local: "testsuites"},
		Name:     "",
		Tests:    9,
		Failures: 1,
		Time:     "",
		Properties: &Properties{
			Property: []Property{
				{Name: "go.version", Value: "1.14"},
			},
		},
		Testsuite: []*Testsuite{
			suite,
		},
	}

	x, _ := xml.MarshalIndent(suites, " ", "  ")
	xout := string(x)
	j, _ := json.MarshalIndent(suites, " ", "  ")
	jout := string(j)

	xmlFile := filepath.Join("testdata", goldenXML+".golden")
	jsonFile := filepath.Join("testdata", goldenJSON+".golden")

	if *updateGolden {
		t.Logf("updating golden files %s and %s", goldenXML, goldenJSON)
		if err := ioutil.WriteFile(xmlFile, []byte(xout), 0600); err != nil {
			t.Fatalf("failed to update golden file: %s", err)
		}
		if err := ioutil.WriteFile(jsonFile, []byte(jout), 0600); err != nil {
			t.Fatalf("failed to update golden file: %s", err)
		}
	}
	gxml, err := ioutil.ReadFile(xmlFile)
	if err != nil {
		t.Fatalf("failed reading .golden: %s", err)
	}
	assert.Equal(t, string(gxml), xout, "for golden file: %s", xmlFile)
	gjson, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		t.Fatalf("failed reading .golden: %s", err)
	}
	assert.Equal(t, string(gjson), jout, "for golden file: %s", jsonFile)
}
