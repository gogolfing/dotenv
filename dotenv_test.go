package dotenv

import (
	"fmt"
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

func TestSourcer_NameVar(t *testing.T) {
	tests := []struct {
		sourcer *Sourcer
		cases   []*nameVarCase
	}{
		{
			//default sourcer.
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

				{"=", "", "", ErrEmptyName("=")},
				{" = ", "", "", ErrEmptyName(" = ")},
				{"=a", "", "", ErrEmptyName("=a")},
				{"a= b", "a", "", ErrInvalidWhitespaceVariablePrefix(" b")},
				{`a="`, "a", "", &ErrVariableUnclosedQuote{`"`, `"`}},
				{`a="  b`, "a", "", &ErrVariableUnclosedQuote{`"  b`, `"`}},
				{"a#b=value", "", "", &ErrCommentInName{"a#b", "#"}},

				{"export =", "", "", ErrEmptyName("export =")},
				{"export  = ", "", "", ErrEmptyName("export  = ")},
				{"export =a", "", "", ErrEmptyName("export =a")},
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
		},
		{
			func() *Sourcer {
				s := New()
				s.Export = ""
				return s
			}(),
			[]*nameVarCase{
				/*
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

					{"=", "", "", ErrEmptyName("=")},
					{" = ", "", "", ErrEmptyName(" = ")},
					{"=a", "", "", ErrEmptyName("=a")},
					{"a= b", "a", "", ErrInvalidWhitespaceVariablePrefix(" b")},
					{`a="`, "a", "", &ErrVariableUnclosedQuote{`"`, `"`}},
					{`a="  b`, "a", "", &ErrVariableUnclosedQuote{`"  b`, `"`}},
					{"a#b=value", "", "", &ErrCommentInName{"a#b", "#"}},
				*/

				{"export =", "", "", ErrEmptyName("export =")},
				/*
					{"export  = ", "", "", ErrEmptyName("export  = ")},
					{"export =a", "", "", ErrEmptyName("export =a")},
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
				*/
			},
		},
	}
	for testIndex, test := range tests {
		for caseIndex, nvc := range test.cases {
			name, v, err := test.sourcer.NameVar(nvc.line)
			fmt.Println(testIndex, caseIndex)
			if name != nvc.name || v != nvc.v || !reflect.DeepEqual(err, nvc.err) {
				t.Logf("%v %q %q %q %v", "+++++++++++++++++++++++++++", nvc.line, name, v, len(name))
				t.Errorf(
					"%v, %v test.sourcer.NameVar(%v) = %q, %q, %v WANT %q, %q, %v",
					testIndex,
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
}

type nameVarCase struct {
	line string
	name string
	v    string
	err  error
}
