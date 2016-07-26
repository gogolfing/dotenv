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

type ErrVariableUnclosedQuote struct {
	Variable string
	Quote    string
}

func (e *ErrVariableUnclosedQuote) Error() string {
	return fmt.Sprintf("variable %q cannot start with unclosed quote %q", e.Variable, e.Quote)
}

type ErrNonVariableLine string

func (e ErrNonVariableLine) Error() string {
	return fmt.Sprintf("line does not contain a variable definition %q", string(e))
}

type ErrInvalidName string

func (e ErrInvalidName) Error() string {
	return fmt.Sprintf("name %q is invalid", string(e))
}

var ErrEmptyLine = errors.New("empty line")

type Sourcer struct {
	Comment string
	Quote   string
	Export  string
	Unquote func(s string) (t string, err error)
}

func New() *Sourcer {
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
	origLine := line

	//get rid of any whitespace at the start of the line. doesn't really matter.
	line = strings.TrimLeft(line, SpaceTab)

	//check for s.Export at beginning of line.
	if strings.HasPrefix(line, s.Export) && s.Export != "" {
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
	name, v = strings.TrimLeft(line[:equalIndex], SpaceTab), line[equalIndex+1:]

	//evaluate name for errors.
	if s.isNameInvalid(name) {
		return "", "", ErrInvalidName(name)
	}

	//if a comment appears in name (before Equal) then it is a comment line.
	if strings.Contains(name, s.Comment) && s.Comment != "" {
		return "", "", ErrInvalidName(name)
	}

	//fix and return variable part with possible error.
	v, err = s.fixVariable(v)
	return name, v, err
}

//isNameInvalid determines whether or not name is valid in s.
func (s *Sourcer) isNameInvalid(name string) bool {
	return len(name) == 0 ||
		strings.ContainsAny(name, SpaceTab) ||
		(strings.Contains(name, s.Comment) && s.Comment != "")
}

//fixVariable returns the actual variable value to set parsed from v.
//v should be the remainder of a line after the first equal sign.
//It may contain a comment.
func (s *Sourcer) fixVariable(v string) (string, error) {
	origV := v

	//if v is empty, then just return the empty string and no error.
	if len(v) == 0 {
		return v, nil
	}

	//if v starts with s.Quote, then assume it either ends with one and unquote
	//or v should be returned literally.
	if strings.HasPrefix(v, s.Quote) {
		//if starts and ends with quote but not equal to quote.
		if strings.HasSuffix(v, s.Quote) && v != s.Quote {
			return s.Unquote(v)
		}
		return "", &ErrVariableUnclosedQuote{origV, s.Quote}
	}

	//if there is a comment, then get rid of it.
	commentIndex := strings.Index(v, s.Comment)
	if commentIndex >= 0 {
		v = v[:commentIndex]
	}
	//trim any right whitespace.
	v = strings.TrimRight(v, SpaceTab)

	if v != strings.TrimLeft(v, SpaceTab) {
		return "", ErrInvalidWhitespaceVariablePrefix(origV)
	}

	return v, nil
}
