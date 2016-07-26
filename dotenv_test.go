package dotenv

import (
	"reflect"
	"testing"
)

func TestNewSourcer(t *testing.T) {
	s := New()
	if s == nil {
		t.Fail()
	}
	if s.Comment != DefaultComment || s.Export != Export || s.Quote != DefaultQuote {
		t.Fail()
	}
	if s.Unquote == nil {
		t.Fail()
	}
}

func TestSourcer_SourceFile(t *testing.T) {
}

func TestSourcer_Source_success(t *testing.T) {
}

func TestSourcer_Source_error(t *testing.T) {
}

func TestSourcer_NameVar_default(t *testing.T) {
	testSourcerNameVarCases(
		t,
		New(),
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

			{"=", "", "", ErrInvalidName("")},
			{" = ", "", "", ErrInvalidName("")},
			{"=a", "", "", ErrInvalidName("")},
			{"a= b", "a", "", ErrInvalidWhitespaceVariablePrefix(" b")},
			{`a="`, "a", "", &ErrVariableUnclosedQuote{`"`, `"`}},
			{`a="  b`, "a", "", &ErrVariableUnclosedQuote{`"  b`, `"`}},
			{"a#b=value", "", "", ErrInvalidName("a#b")},

			{"export =", "", "", ErrInvalidName("")},
			{"export  = ", "", "", ErrInvalidName("")},
			{"export =a", "", "", ErrInvalidName("")},
			{"export a= b", "a", "", ErrInvalidWhitespaceVariablePrefix(" b")},
			{`export a="`, "a", "", &ErrVariableUnclosedQuote{`"`, `"`}},
			{`export a="  b`, "a", "", &ErrVariableUnclosedQuote{`"  b`, `"`}},

			{"a=", "a", "", nil},
			{"a= ", "a", "", nil},
			{"a=#", "a", "", nil},
			{"a= #", "a", "", nil},
			{"a=b", "a", "b", nil},
			{"a=b ", "a", "b", nil},
			{"a=b  c", "a", "b  c", nil},
			{`abcd="foobar"`, "abcd", "foobar", nil},
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
	s := New()
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

			{"=", "", "", ErrInvalidName("")},
			{" = ", "", "", ErrInvalidName("")},
			{"=a", "", "", ErrInvalidName("")},
			{"a= b", "a", "", ErrInvalidWhitespaceVariablePrefix(" b")},
			{`a="`, "a", "", &ErrVariableUnclosedQuote{`"`, `"`}},
			{`a="  b`, "a", "", &ErrVariableUnclosedQuote{`"  b`, `"`}},
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
}

func TestSourcer_NameVar_emptyQuote(t *testing.T) {
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
