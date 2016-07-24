package dotenv

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	DefaultComment = "#"
	DefaultQuote   = `"`

	Export = "export"

	SpaceTab = " \t"
)

type ErrSourcing struct {
	Line      int
	LineError error
}

func (e *ErrSourcing) Error() string {
	return fmt.Sprintf("dotenv: line %v %v", e.Line, e.LineError.Error())
}

type ErrInvalidWhitespaceVariablePrefix string

func (e ErrInvalidWhitespaceVariablePrefix) Error() string {
	return fmt.Sprintf("invalid whitespace at beginning of variable %q", string(e))
}

type ErrVariableIsQuote struct {
	Variable string
	Quote    string
}

func (e *ErrVariableIsQuote) Error() string {
	return fmt.Sprintf("variable %q cannot be the quote %q", e.Variable, e.Quote)
}

type ErrNonVariableLine string

func (e ErrNonVariableLine) Error() string {
	return fmt.Sprintf("line does not contain a variable definition %q", string(e))
}

type ErrEmptyName string

func (e ErrEmptyName) Error() string {
	return fmt.Sprintf("line contains an empty name %q", string(e))
}

var ErrEmptyLine = errors.New("empty line")

type Sourcer struct {
	Comment string
	Quote   string
	Export  string
	Unquote func(s string) (t string, err error)
}

func NewSourcer() *Sourcer {
	return &Sourcer{
		Comment: DefaultComment,
		Quote:   DefaultQuote,
		Export:  Export,
		Unquote: strconv.Unquote,
	}
}

func (s *Sourcer) SourceFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	if err := s.Source(file); err != nil {
		return err
	}
	return file.Close()
}

//not guaranteed to read all of in.
func (s *Sourcer) Source(in io.Reader) error {
	lineNumber := 0
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++
		name, v, err := s.NameVar(line)
		if err == ErrEmptyLine {
			continue
		}
		if err != nil {
			return &ErrSourcing{lineNumber, err}
		}
		if err := os.Setenv(name, v); err != nil {
			return &ErrSourcing{lineNumber, err}
		}
	}
	return scanner.Err()
}

func (s *Sourcer) NameVar(line string) (name, v string, err error) {
	//check for s.Export at beginning of line.
	origLine := line
	if strings.HasPrefix(line, s.Export) {
		line = strings.TrimPrefix(line, s.Export)
		line = strings.TrimLeft(line, SpaceTab)
		if len(line) == 0 || strings.HasPrefix(line, s.Comment) {
			return "", "", ErrNonVariableLine(origLine)
		}
	}

	//check for Equal in the line.
	equalIndex := strings.Index(line, "=")
	if equalIndex < 0 {
		line = strings.TrimLeft(line, SpaceTab)
		if len(line) == 0 || strings.HasPrefix(line, s.Comment) {
			return "", "", ErrEmptyLine
		}
		return "", "", ErrNonVariableLine(origLine)
	}

	//get name and varible parts of the line. trim the name.
	name, v = strings.Trim(line[:equalIndex], SpaceTab), line[:equalIndex+1]
	if len(name) == 0 {
		return "", "", ErrEmptyName(origLine)
	}
	//if a comment appears in name (before Equal) then it is a comment line.
	if strings.Contains(name, s.Comment) {
		return "", "", ErrEmptyLine
	}

	//fix and return variable part with possible error.
	v, err = s.fixVariable(v)
	return name, v, err
}

func (s *Sourcer) fixVariable(v string) (string, error) {
	trimmed := strings.TrimLeft(v, SpaceTab)
	if trimmed != v {
		return "", ErrInvalidWhitespaceVariablePrefix(v)
	}
	if len(v) == 0 {
		return v, nil
	}
	if v == s.Quote {
		return "", &ErrVariableIsQuote{v, s.Quote}
	}
	if strings.HasPrefix(v, s.Quote) {
		if strings.HasSuffix(v, s.Quote) {
			return s.Unquote(v[len(s.Quote) : len(v)-len(s.Quote)])
		}
	}
	commentIndex := strings.Index(v, s.Comment)
	if commentIndex < 0 {
		return v, nil
	}
	v = strings.TrimRight(v[:commentIndex], SpaceTab)
	return v, nil
}
