package dotenv

import "testing"

func TestNewSourcer(t *testing.T) {
	s := NewSourcer()
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

				{"=", "", "", ErrEmptyName("=")},
				{" = ", "", "", ErrEmptyName(" = ")},
				{"=a", "", "", ErrEmptyName("=a")},
				//other errors here.
				{"a= ", "a", "", ErrInvalidWhitespaceVariablePrefix(" ")},

				//space before export
			},
		},
	}
	for testIndex, test := range tests {
		for caseIndex, nvc := range test.cases {
			name, v, err := test.sourcer.NameVar(nvc.line)
			if name != nvc.name || v != nvc.v || err != nvc.err {
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
