/*
 * Copyright (c) 2013-2016 Dave Collins <dave@davec.name>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package spew_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/picmonkey/go-spew/spew"
)

// spewFunc is used to identify which public function of the spew package or
// ConfigState a test applies to.
type spewFunc int

const (
	fCSFdump spewFunc = iota
	fCSFprint
	fCSFprintf
	fCSFprintln
	fCSPrint
	fCSPrintln
	fCSSdump
	fCSSprint
	fCSSprintf
	fCSSprintln
	fCSErrorf
	fCSNewFormatter
	fErrorf
	fFprint
	fFprintln
	fPrint
	fPrintln
	fSdump
	fSprint
	fSprintf
	fSprintln
)

// Map of spewFunc values to names for pretty printing.
var spewFuncStrings = map[spewFunc]string{
	fCSFdump:        "ConfigState.Fdump",
	fCSFprint:       "ConfigState.Fprint",
	fCSFprintf:      "ConfigState.Fprintf",
	fCSFprintln:     "ConfigState.Fprintln",
	fCSSdump:        "ConfigState.Sdump",
	fCSPrint:        "ConfigState.Print",
	fCSPrintln:      "ConfigState.Println",
	fCSSprint:       "ConfigState.Sprint",
	fCSSprintf:      "ConfigState.Sprintf",
	fCSSprintln:     "ConfigState.Sprintln",
	fCSErrorf:       "ConfigState.Errorf",
	fCSNewFormatter: "ConfigState.NewFormatter",
	fErrorf:         "spew.Errorf",
	fFprint:         "spew.Fprint",
	fFprintln:       "spew.Fprintln",
	fPrint:          "spew.Print",
	fPrintln:        "spew.Println",
	fSdump:          "spew.Sdump",
	fSprint:         "spew.Sprint",
	fSprintf:        "spew.Sprintf",
	fSprintln:       "spew.Sprintln",
}

func (f spewFunc) String() string {
	if s, ok := spewFuncStrings[f]; ok {
		return s
	}
	return fmt.Sprintf("Unknown spewFunc (%d)", int(f))
}

// spewTest is used to describe a test to be performed against the public
// functions of the spew package or ConfigState.
type spewTest struct {
	cs     *spew.ConfigState
	f      spewFunc
	format string
	in     interface{}
	want   string
}

// spewTests houses the tests to be performed against the public functions of
// the spew package and ConfigState.
//
// These tests are only intended to ensure the public functions are exercised
// and are intentionally not exhaustive of types.  The exhaustive type
// tests are handled in the dump and format tests.
var spewTests []spewTest

// redirStdout is a helper function to return the standard output from f as a
// byte slice.
func redirStdout(f func()) ([]byte, error) {
	tempFile, err := ioutil.TempFile("", "ss-test")
	if err != nil {
		return nil, err
	}
	fileName := tempFile.Name()
	defer os.Remove(fileName) // Ignore error

	origStdout := os.Stdout
	os.Stdout = tempFile
	f()
	os.Stdout = origStdout
	tempFile.Close()

	return ioutil.ReadFile(fileName)
}

type hashTester struct {
	str         string
	nonString   int       `spew:"nonStringHash,hash"`
	string      string    `spew:"hash,hash"`
	ptr         *string   `spew:"hashPtr,hash"`
	slice       []string  `spew:"hashSlice,hash"`
	ptrSlice    []*string `spew:"hashPtrSlice,hash"`
	norm1       string    `spew:"normalize,normalize"`
	normandhash string    `spew:"normalizeHash,normalize,hash"`
	hashNorm    string    `spew:"hashNormalize,hash,normalize"`
}

var str = "test"
var str2 = "test2"

var hashVal = &hashTester{
	str:         "test",
	nonString:   2,
	string:      "test",
	ptr:         &str,
	slice:       []string{"test", "test2"},
	ptrSlice:    []*string{&str, &str2},
	norm1:       "TeSt",
	normandhash: "TeSt",
	hashNorm:    "TeSt",
}

var hashValCopy = *hashVal

var expectedHashedDump = `(*spew_test.hashTester)({
str: (string) (len=4) "test",
nonStringHash: (int) 2,
hash: (string) (len=4) "098f6bcd4621d373cade4e832627b4f6",
hashPtr: (*string)((len=4) "098f6bcd4621d373cade4e832627b4f6"),
hashSlice: ([]string) (len=2 cap=2) {
(string) (len=4) "098f6bcd4621d373cade4e832627b4f6",
(string) (len=5) "ad0234829205b9033196ba818f7a872b"
},
hashPtrSlice: ([]*string) (len=2 cap=2) {
(*string)(<already shown>),
(*string)((len=5) "ad0234829205b9033196ba818f7a872b")
},
normalize: (string) (len=4) "test",
normalizeHash: (string) (len=4) "098f6bcd4621d373cade4e832627b4f6",
hashNormalize: (string) (len=4) "cc2d246d9a763bac2a3f5ec5ecb68012"
})
`

func initSpewTests() {
	// Config states with various settings.
	scsDefault := spew.NewDefaultConfig()
	scsNoMethods := &spew.ConfigState{Indent: " ", DisableMethods: true}
	scsNoPmethods := &spew.ConfigState{Indent: " ", DisablePointerMethods: true}
	scsMaxDepth := &spew.ConfigState{Indent: " ", MaxDepth: 1}
	scsContinue := &spew.ConfigState{Indent: " ", ContinueOnMethod: true}
	scsNoPtrAddr := &spew.ConfigState{DisablePointerAddresses: true}
	scsNoCap := &spew.ConfigState{DisableCapacities: true}

	// Variables for tests on types which implement Stringer interface with and
	// without a pointer receiver.
	ts := stringer("test")
	tps := pstringer("test")

	type ptrTester struct {
		s *struct{}
	}
	tptr := &ptrTester{s: &struct{}{}}

	// depthTester is used to test max depth handling for structs, array, slices
	// and maps.
	type depthTester struct {
		ic    indirCir1
		arr   [1]string
		slice []string
		m     map[string]int
	}
	dt := depthTester{indirCir1{nil}, [1]string{"arr"}, []string{"slice"},
		map[string]int{"one": 1}}

	// Variable for tests on types which implement error interface.
	te := customError(10)

	type ignoreTester struct {
		visible bool
		ignore  bool `spew:"-"`
	}

	type ignoreTester2 struct {
		visible  bool
		ignore   bool `spew:"-"`
		visible2 bool
	}

	spewTests = []spewTest{
		{scsDefault, fCSFdump, "", int8(127), "(int8) 127\n"},
		{scsDefault, fCSFprint, "", int16(32767), "32767"},
		{scsDefault, fCSFprintf, "%v", int32(2147483647), "2147483647"},
		{scsDefault, fCSFprintln, "", int(2147483647), "2147483647\n"},
		{scsDefault, fCSPrint, "", int64(9223372036854775807), "9223372036854775807"},
		{scsDefault, fCSPrintln, "", uint8(255), "255\n"},
		{scsDefault, fCSSdump, "", uint8(64), "(uint8) 64\n"},
		{scsDefault, fCSSprint, "", complex(1, 2), "(1+2i)"},
		{scsDefault, fCSSprintf, "%v", complex(float32(3), 4), "(3+4i)"},
		{scsDefault, fCSSprintln, "", complex(float64(5), 6), "(5+6i)\n"},
		{scsDefault, fCSErrorf, "%#v", uint16(65535), "(uint16)65535"},
		{scsDefault, fCSNewFormatter, "%v", uint32(4294967295), "4294967295"},
		{scsDefault, fErrorf, "%v", uint64(18446744073709551615), "18446744073709551615"},
		{scsDefault, fFprint, "", float32(3.14), "3.14"},
		{scsDefault, fFprintln, "", float64(6.28), "6.28\n"},
		{scsDefault, fPrint, "", true, "true"},
		{scsDefault, fPrintln, "", false, "false\n"},
		{scsDefault, fSdump, "", complex(-10, -20), "(complex128) (-10-20i)\n"},
		{scsDefault, fSprint, "", complex(-1, -2), "(-1-2i)"},
		{scsDefault, fSprintf, "%v", complex(float32(-3), -4), "(-3-4i)"},
		{scsDefault, fSprintln, "", complex(float64(-5), -6), "(-5-6i)\n"},
		{scsDefault, fCSFdump, "", ignoreTester{
			visible: true,
			ignore:  true,
		}, "(spew_test.ignoreTester) {\n visible: (bool) true\n}\n"},
		{scsDefault, fCSFdump, "", ignoreTester2{
			visible:  true,
			ignore:   true,
			visible2: true,
		}, "(spew_test.ignoreTester2) {\n visible: (bool) true,\n visible2: (bool) true\n}\n"},
		{scsNoPtrAddr, fCSFdump, "", hashVal, expectedHashedDump},
		{scsNoMethods, fCSFprint, "", ts, "test"},
		{scsNoMethods, fCSFprint, "", &ts, "<*>test"},
		{scsNoMethods, fCSFprint, "", tps, "test"},
		{scsNoMethods, fCSFprint, "", &tps, "<*>test"},
		{scsNoPmethods, fCSFprint, "", ts, "stringer test"},
		{scsNoPmethods, fCSFprint, "", &ts, "<*>stringer test"},
		{scsNoPmethods, fCSFprint, "", tps, "test"},
		{scsNoPmethods, fCSFprint, "", &tps, "<*>stringer test"},
		{scsMaxDepth, fCSFprint, "", dt, "{{<max>} [<max>] [<max>] map[<max>]}"},
		{scsMaxDepth, fCSFdump, "", dt, "(spew_test.depthTester) {\n" +
			" ic: (spew_test.indirCir1) {\n  <max depth reached>\n },\n" +
			" arr: ([1]string) (len=1 cap=1) {\n  <max depth reached>\n },\n" +
			" slice: ([]string) (len=1 cap=1) {\n  <max depth reached>\n },\n" +
			" m: (map[string]int) (len=1) {\n  <max depth reached>\n }\n}\n"},
		{scsContinue, fCSFprint, "", ts, "(stringer test) test"},
		{scsContinue, fCSFdump, "", ts, "(spew_test.stringer) " +
			"(len=4) (stringer test) \"test\"\n"},
		{scsContinue, fCSFprint, "", te, "(error: 10) 10"},
		{scsContinue, fCSFdump, "", te, "(spew_test.customError) " +
			"(error: 10) 10\n"},
		{scsNoPtrAddr, fCSFprint, "", tptr, "<*>{<*>{}}"},
		{scsNoPtrAddr, fCSSdump, "", tptr, "(*spew_test.ptrTester)({\ns: (*struct {})({\n})\n})\n"},
		{scsNoCap, fCSSdump, "", make([]string, 0, 10), "([]string) {\n}\n"},
		{scsNoCap, fCSSdump, "", make([]string, 1, 10), "([]string) (len=1) {\n(string) \"\"\n}\n"},
	}
}

// TestSpew executes all of the tests described by spewTests.
func TestSpew(t *testing.T) {
	initSpewTests()

	t.Logf("Running %d tests", len(spewTests))
	for i, test := range spewTests {
		buf := new(bytes.Buffer)
		switch test.f {
		case fCSFdump:
			test.cs.Fdump(buf, test.in)

		case fCSFprint:
			test.cs.Fprint(buf, test.in)

		case fCSFprintf:
			test.cs.Fprintf(buf, test.format, test.in)

		case fCSFprintln:
			test.cs.Fprintln(buf, test.in)

		case fCSPrint:
			b, err := redirStdout(func() { test.cs.Print(test.in) })
			if err != nil {
				t.Errorf("%v #%d %v", test.f, i, err)
				continue
			}
			buf.Write(b)

		case fCSPrintln:
			b, err := redirStdout(func() { test.cs.Println(test.in) })
			if err != nil {
				t.Errorf("%v #%d %v", test.f, i, err)
				continue
			}
			buf.Write(b)

		case fCSSdump:
			str := test.cs.Sdump(test.in)
			buf.WriteString(str)

		case fCSSprint:
			str := test.cs.Sprint(test.in)
			buf.WriteString(str)

		case fCSSprintf:
			str := test.cs.Sprintf(test.format, test.in)
			buf.WriteString(str)

		case fCSSprintln:
			str := test.cs.Sprintln(test.in)
			buf.WriteString(str)

		case fCSErrorf:
			err := test.cs.Errorf(test.format, test.in)
			buf.WriteString(err.Error())

		case fCSNewFormatter:
			fmt.Fprintf(buf, test.format, test.cs.NewFormatter(test.in))

		case fErrorf:
			err := spew.Errorf(test.format, test.in)
			buf.WriteString(err.Error())

		case fFprint:
			spew.Fprint(buf, test.in)

		case fFprintln:
			spew.Fprintln(buf, test.in)

		case fPrint:
			b, err := redirStdout(func() { spew.Print(test.in) })
			if err != nil {
				t.Errorf("%v #%d %v", test.f, i, err)
				continue
			}
			buf.Write(b)

		case fPrintln:
			b, err := redirStdout(func() { spew.Println(test.in) })
			if err != nil {
				t.Errorf("%v #%d %v", test.f, i, err)
				continue
			}
			buf.Write(b)

		case fSdump:
			str := spew.Sdump(test.in)
			buf.WriteString(str)

		case fSprint:
			str := spew.Sprint(test.in)
			buf.WriteString(str)

		case fSprintf:
			str := spew.Sprintf(test.format, test.in)
			buf.WriteString(str)

		case fSprintln:
			str := spew.Sprintln(test.in)
			buf.WriteString(str)

		default:
			t.Errorf("%v #%d unrecognized function", test.f, i)
			continue
		}
		s := buf.String()
		if test.want != s {
			t.Errorf("ConfigState #%d\n got: %s want: %s", i, s, test.want)
			continue
		}
	}
	if !reflect.DeepEqual(*hashVal, hashValCopy) {
		t.Errorf("%v != %v", *hashVal, hashValCopy)
	}
}
