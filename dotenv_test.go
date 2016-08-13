package dotenv

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

const SampleSource = `
GOGOLFING_DOTENV_A=A #comment
GOGOLFING_DOTENV_B="B"

#GOGOLFING_DOTENV_C=C
`

func TestErrSourcing_Error(t *testing.T) {
	err := &ErrSourcing{
		Line:      100,
		LineError: fmt.Errorf("line error"),
	}
	if err.Error() != "dotenv: line 100 line error" {
		t.Fail()
	}
}

func TestErrInvalidWhitespaceVariablePrefix_Error(t *testing.T) {
	err := ErrInvalidWhitespaceValuePrefix(" value")
	if err.Error() != `invalid whitespace at beginning of value " value"` {
		t.Fail()
	}
}

func TestErrVariableUnclosedQuote_Error(t *testing.T) {
	err := &ErrValueUnclosedQuote{
		Variable: `"value`,
		Quote:    `"`,
	}
	if err.Error() != `value "\"value" cannot start with unclosed quote "\""` {
		t.Fail()
	}
}

func TestErrNonVariableLine_Error(t *testing.T) {
	err := ErrNonVariableLine("line")
	if err.Error() != `line does not contain a variable definition "line"` {
		t.Fail()
	}
}

func TestErrInvalidName_Error(t *testing.T) {
	err := ErrInvalidName("invalid name")
	if err.Error() != `name "invalid name" is invalid` {
		t.Fail()
	}
}

func TestNewSourcer(t *testing.T) {
	s := NewSourcer()
	if s == nil {
		t.Fail()
	}
	if s.Comment != DefaultComment || s.Export != DefaultExport || s.Quote != DefaultQuote {
		t.Fail()
	}
	if s.Unquote == nil {
		t.Fail()
	}
}

func TestSourcer_SourceFile(t *testing.T) {
	file, err := ioutil.TempFile("", "gogolfing.dotenv")
	if err != nil {
		t.Error(err)
	}

	defer os.Remove(file.Name())

	os.Setenv("GOGOLFING_DOTENV_A", "")
	os.Setenv("GOGOLFING_DOTENV_B", "")
	os.Setenv("GOGOLFING_DOTENV_C", "")

	if _, err := fmt.Fprint(file, SampleSource); err != nil {
		t.Error(err)
	}

	if err := file.Close(); err != nil {
		t.Error(err)
	}

	sourcer := NewSourcer()

	if err := sourcer.SourceFile(file.Name()); err != nil {
		t.Error(err)
	}

	if os.Getenv("GOGOLFING_DOTENV_A") != "A" {
		t.Fail()
	}
	if os.Getenv("GOGOLFING_DOTENV_B") != "B" {
		t.Fail()
	}
	if os.Getenv("GOGOLFING_DOTENV_C") != "" {
		t.Fail()
	}
}

func TestSourcer_Source_success(t *testing.T) {
	sourcer := NewSourcer()

	if err := sourcer.Source(strings.NewReader(SampleSource)); err != nil {
		t.Error(err)
	}
}

func TestSourcer_Source_error(t *testing.T) {
	sourcer := NewSourcer()

	line := "export"

	_, _, lineError := sourcer.NameVar(line)

	err := sourcer.Source(strings.NewReader(line))

	sourceError := err.(*ErrSourcing)

	if sourceError.Line != 1 {
		t.Fail()
	}
	if sourceError.LineError != lineError {
		t.Fail()
	}
}

func TestSourcer_NameVars_success(t *testing.T) {
	sourcer := NewSourcer()
	nameVars, err := sourcer.NameVars(strings.NewReader("name=value"))
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(nameVars, [][2]string{{"name", "value"}}) {
		t.Fail()
	}
}

func TestSourcer_NameVars_error(t *testing.T) {
	sourcer := NewSourcer()
	nameVars, err := sourcer.NameVars(strings.NewReader("name"))
	if nameVars != nil || err == nil {
		t.Fail()
	}
}

func TestSourcer_sourceVisitor(t *testing.T) {
	visitor := func(name, v string) error {
		return errors.New("visitor error")
	}
	sourcer := NewSourcer()
	err := sourcer.sourceVisitor(strings.NewReader("name=value"), visitor)
	if !reflect.DeepEqual(err, &ErrSourcing{1, errors.New("visitor error")}) {
		t.Fail()
	}
}

func TestSourcer_NameVar_default(t *testing.T) {
	testSourcerNameVarCases(
		t,
		NewSourcer(),
		[]*nameVarCase{
			{"", "", "", ErrEmptyLine},
			{SpaceTab, "", "", ErrEmptyLine},
			{"#comment", "", "", ErrEmptyLine},
			{SpaceTab + "#comment", "", "", ErrEmptyLine},
			{"a", "", "", ErrNonVariableLine("a")},

			{"export", "", "", ErrNonVariableLine("export")},
			{"export" + SpaceTab, "", "", ErrNonVariableLine("export" + SpaceTab)},
			{"export#comment", "", "", ErrNonVariableLine("export#comment")},
			{"export \t#comment", "", "", ErrNonVariableLine("export \t#comment")},
			{"export a", "", "", ErrNonVariableLine("export a")},

			{"#name=value", "", "", ErrEmptyLine},
			{" #name=value", "", "", ErrEmptyLine},
			{"# name=value", "", "", ErrEmptyLine},
			{" # name=value", "", "", ErrEmptyLine},
			{"export #name=value", "", "", ErrNonVariableLine("export #name=value")},
			{"export # name=value", "", "", ErrNonVariableLine("export # name=value")},

			{"=", "", "", ErrInvalidName("")},
			{" = ", "", "", ErrInvalidName("")},
			{"=a", "", "", ErrInvalidName("")},
			{"a= b", "a", "", ErrInvalidWhitespaceValuePrefix(" b")},
			{`a= "b`, "a", "", ErrInvalidWhitespaceValuePrefix(` "b`)},
			{`a="`, "a", "", &ErrValueUnclosedQuote{`"`, `"`}},
			{`a="  b`, "a", "", &ErrValueUnclosedQuote{`"  b`, `"`}},
			{"a#b=value", "", "", ErrInvalidName("a#b")},
			{"a b=value", "", "", ErrInvalidName("a b")},

			{"export =", "", "", ErrInvalidName("")},
			{"export  = ", "", "", ErrInvalidName("")},
			{"export =a", "", "", ErrInvalidName("")},
			{"export a= b", "a", "", ErrInvalidWhitespaceValuePrefix(" b")},
			{`export a="`, "a", "", &ErrValueUnclosedQuote{`"`, `"`}},
			{`export a="  b`, "a", "", &ErrValueUnclosedQuote{`"  b`, `"`}},

			{"a=", "a", "", nil},
			{"a= ", "a", "", nil},
			{"a=#", "a", "", nil},
			{"a= #", "a", "", nil},
			{"a=b", "a", "b", nil},
			{"a=b ", "a", "b", nil},
			{"a=b  c", "a", "b  c", nil},
			//have a looksee
			{`a=b"c`, "a", `b"c`, nil},
			{`abcd="foobar"`, "abcd", "foobar", nil},
			{`ab"cd=foobar`, `ab"cd`, "foobar", nil},
			{`A_B_C_D="foo\nbar"`, "A_B_C_D", "foo\nbar", nil},

			{"export a=", "a", "", nil},
			{"export  a= ", "a", "", nil},
			{"export \t\ta=#", "a", "", nil},
			{"export a= #", "a", "", nil},
			{"export a=b", "a", "b", nil},
			{"export a=b ", "a", "b", nil},
			{"export a=b  c", "a", "b  c", nil},
			{`export abcd="foobar"`, "abcd", "foobar", nil},
			{`export A_B_C_D="foo\nbar"`, "A_B_C_D", "foo\nbar", nil},

			{" export a=", "a", "", nil},
			{"  export  a= ", "a", "", nil},
			{" \t\texport \t\ta=#", "a", "", nil},
			{" export a= #", "a", "", nil},
			{" export a=b", "a", "b", nil},
			{" export a=b ", "a", "b", nil},
			{" export a=b  c", "a", "b  c", nil},
			{` export abcd="foobar"`, "abcd", "foobar", nil},
			{` export A_B_C_D="foo\nbar"`, "A_B_C_D", "foo\nbar", nil},
		},
	)
}

func TestSourcer_NameVar_emptyExport(t *testing.T) {
	s := NewSourcer()
	s.Export = ""
	testSourcerNameVarCases(
		t,
		s,
		[]*nameVarCase{
			{"", "", "", ErrEmptyLine},
			{SpaceTab, "", "", ErrEmptyLine},
			{"#comment", "", "", ErrEmptyLine},
			{SpaceTab + "#comment", "", "", ErrEmptyLine},
			{"a", "", "", ErrNonVariableLine("a")},

			{"export", "", "", ErrNonVariableLine("export")},
			{"export" + SpaceTab, "", "", ErrNonVariableLine("export" + SpaceTab)},
			{"export#comment", "", "", ErrNonVariableLine("export#comment")},
			{"export \t#comment", "", "", ErrNonVariableLine("export \t#comment")},
			{"export a", "", "", ErrNonVariableLine("export a")},

			{"#name=value", "", "", ErrEmptyLine},
			{" #name=value", "", "", ErrEmptyLine},
			{"# name=value", "", "", ErrEmptyLine},
			{" # name=value", "", "", ErrEmptyLine},
			{"export #name=value", "", "", ErrInvalidName("export #name")},
			{"export # name=value", "", "", ErrInvalidName("export # name")},

			{"=", "", "", ErrInvalidName("")},
			{" = ", "", "", ErrInvalidName("")},
			{"=a", "", "", ErrInvalidName("")},
			{"a= b", "a", "", ErrInvalidWhitespaceValuePrefix(" b")},
			{`a="`, "a", "", &ErrValueUnclosedQuote{`"`, `"`}},
			{`a="  b`, "a", "", &ErrValueUnclosedQuote{`"  b`, `"`}},
			{"a#b=value", "", "", ErrInvalidName("a#b")},

			{"export =", "", "", ErrInvalidName("export ")},
			{"export  = ", "", "", ErrInvalidName("export  ")},
			{"export =a", "", "", ErrInvalidName("export ")},
			{"export a= b", "", "", ErrInvalidName("export a")},
			{`export a="`, "", "", ErrInvalidName("export a")},
			{`export a="  b`, "", "", ErrInvalidName("export a")},
			{"export a=b", "", "", ErrInvalidName("export a")},

			{"a=", "a", "", nil},
			{"a= ", "a", "", nil},
			{"a=#", "a", "", nil},
			{"a= #", "a", "", nil},
			{"a=b", "a", "b", nil},
			{"a=b ", "a", "b", nil},
			{"a=b  c", "a", "b  c", nil},
			{`abcd="foobar"`, "abcd", "foobar", nil},
			{`A_B_C_D="foo\nbar"`, "A_B_C_D", "foo\nbar", nil},

			{" a=", "a", "", nil},
			{"  a= ", "a", "", nil},
			{" \t\ta=#", "a", "", nil},
			{" a= #", "a", "", nil},
			{" a=b", "a", "b", nil},
			{" a=b ", "a", "b", nil},
			{" a=b  c", "a", "b  c", nil},
			{` abcd="foobar"`, "abcd", "foobar", nil},
			{` A_B_C_D="foo\nbar"`, "A_B_C_D", "foo\nbar", nil},
		},
	)
}

func TestSourcer_NameVar_emptyComment(t *testing.T) {
	s := NewSourcer()
	s.Comment = ""
	testSourcerNameVarCases(
		t,
		s,
		[]*nameVarCase{
			{"", "", "", ErrEmptyLine},
			{SpaceTab, "", "", ErrEmptyLine},
			{"#comment", "", "", ErrNonVariableLine("#comment")},
			{SpaceTab + "#comment", "", "", ErrNonVariableLine(SpaceTab + "#comment")},
			{"a", "", "", ErrNonVariableLine("a")},

			{"#name=value", "#name", "value", nil},
			{" #name=value", "#name", "value", nil},
			{"# name=value", "", "", ErrInvalidName("# name")},
			{" # name=value", "", "", ErrInvalidName("# name")},
			{"export #name=value", "", "", ErrNonVariableLine("export #name=value")},
			{"export # name=value", "", "", ErrNonVariableLine("export # name=value")},

			{"a#b=something", "a#b", "something", nil},
			{"ab=some#thing", "ab", "some#thing", nil},
			{"ab=something    #", "ab", "something    #", nil},
			{"ab=some#thing    ", "ab", "some#thing", nil},
			{"ab=something    ", "ab", "something", nil},
			{"a//b=some//thing    ", "a//b", "some//thing", nil},
		},
	)
}

func TestSourcer_NameVar_emptyQuote(t *testing.T) {
	s := NewSourcer()
	s.Quote = ""
	testSourcerNameVarCases(
		t,
		s,
		[]*nameVarCase{
			{`a="hello"`, "a", `"hello"`, nil},
			{`a="hello`, "a", `"hello`, nil},
			{`a="hel\tlo"`, "a", `"hel\tlo"`, nil},
		},
	)
}

func TestSourcer_NameVar_emptyCommentAndQuote(t *testing.T) {
	s := NewSourcer()
	s.Quote = ""
	s.Comment = ""
	testSourcerNameVarCases(
		t,
		s,
		[]*nameVarCase{
			{"", "", "", ErrEmptyLine},
			{SpaceTab, "", "", ErrEmptyLine},
			{"#comment", "", "", ErrNonVariableLine("#comment")},
			{SpaceTab + "#comment", "", "", ErrNonVariableLine(SpaceTab + "#comment")},
			{"a", "", "", ErrNonVariableLine("a")},

			{"a#b=something", "a#b", "something", nil},
			{"ab=some#thing", "ab", "some#thing", nil},
			{"ab=something    #", "ab", "something    #", nil},
			{"ab=some#thing    ", "ab", "some#thing", nil},
			{"ab=something    ", "ab", "something", nil},
			{"a//b=some//thing    ", "a//b", "some//thing", nil},

			{`a="hello"`, "a", `"hello"`, nil},
			{`a="hello`, "a", `"hello`, nil},
			{`a="hel\tlo"`, "a", `"hel\tlo"`, nil},
		},
	)
}

func testSourcerNameVarCases(t *testing.T, s *Sourcer, cases []*nameVarCase) {
	for caseIndex, nvc := range cases {
		name, v, err := s.NameVar(nvc.line)
		if name != nvc.name || v != nvc.v || !reflect.DeepEqual(err, nvc.err) {
			t.Errorf(
				"%v s.NameVar(%v) = %q, %q, %v WANT %q, %q, %v",
				caseIndex,
				nvc.line,
				name,
				v,
				err,
				nvc.name,
				nvc.v,
				nvc.err,
			)
		}
	}
}

type nameVarCase struct {
	line string
	name string
	v    string
	err  error
}
