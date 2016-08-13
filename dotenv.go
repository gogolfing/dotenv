//Package dotenv provides a Sourcer type that allows client code to source
//environment variable inputs and set the values in the process via os.Setenv().
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
	//DefaultComment is the Comment string set to Sourcer.Comment in NewSourcer().
	DefaultComment = "#"

	//DefaultQuote is the Quote string set to Sourcer.Quote in NewSourcer().
	DefaultQuote = `"`

	//DefaultExport is the export string set to Sourcer.Export in NewSourcer().
	DefaultExport = "export"

	//SpaceTab is used in various ways to trim and test certain strings throughout
	//parsing.
	SpaceTab = " \t"
)

//ErrSourcing is an error type that indicates something went wrong while trying
//to source an input.
//See Sourcer's exported methods for use of this type.
type ErrSourcing struct {
	//Line is the line number (1-based) that the error occurred on.
	Line int

	//LineError is an error (of any other error type in this package) that occurred
	//on the specific line.
	LineError error
}

//Error is the error implementation for ErrSourcing. It describes both the e.Line
//and e.LineError.
func (e *ErrSourcing) Error() string {
	return fmt.Sprintf("dotenv: line %v %v", e.Line, e.LineError.Error())
}

//ErrInvalidWhitespaceValuePrefix is a line error that occurs when there is
//whitespace between the equal sign and beginning of the value definition.
type ErrInvalidWhitespaceValuePrefix string

//Error is the error implementation for ErrInvalidWhitespaceValuePrefix.
func (e ErrInvalidWhitespaceValuePrefix) Error() string {
	return fmt.Sprintf("invalid whitespace at beginning of value %q", string(e))
}

//ErrValueUnclosedQuote is a line error that occurs when a value definition starts
//with but does not end with a Quote.
type ErrValueUnclosedQuote struct {
	Variable string
	Quote    string
}

//Error is the error implementation for ErrValueUnclosedQuote.
func (e *ErrValueUnclosedQuote) Error() string {
	return fmt.Sprintf("value %q cannot start with unclosed quote %q", e.Variable, e.Quote)
}

//ErrNonVariableLine is a line error that occurs when a line does not contain or
//resemble a variable definition.
//E.g. "export", "cat in.csv > out.csv", or "name".
type ErrNonVariableLine string

//Error is the error implementation for ErrNonVariableLine.
func (e ErrNonVariableLine) Error() string {
	return fmt.Sprintf("line does not contain a variable definition %q", string(e))
}

//ErrInvalidName is a line error that occurs when a name in a variable definition
//is invalid. Names must not contain a whitespace character, nor contain a Quote
//or Comment string.
type ErrInvalidName string

//Error is the error implementation for ErrInvalidName.
func (e ErrInvalidName) Error() string {
	return fmt.Sprintf("name %q is invalid", string(e))
}

//ErrEmptyLine is a sentinel error value that is returned from Sourcer.NameVar()
//that tells a Sourcer that a line is effectively empty (contains only whitespace
//or whitespace and a comment).
//Note that this is not a semantic error and will never be returned from any methods
//in this package. It is simply used internally for parsing purposes.
var ErrEmptyLine = errors.New("empty line")

//Sourcer is a container for parsing parameters relevant to sourcing environment
//variable inputs.
//A Sourcer is able to take in an io.Reader (or file path) and set the environment
//variables defined in the input on the process via os.Setenv().
type Sourcer struct {
	//Comment denotes the beginning of a comment on a line.
	//An empty Comment value means that all commenting is disallowed.
	//Comment is set to DefaultComment by NewSourcer().
	Comment string

	//Quote denotes the quote string that is allowed to surround a variable's
	//value deinition to allow for whitespace, comment, and escaped values.
	//An empty Quote value means that value quoting is disallowed.
	//Quote is set to DefaultQuote by NewSourcer().
	Quote string

	//Export denotes the possible export keyword that can appear at the beginning
	//of a line without changing the semantics of the line within this package.
	//This is provided so that a valid Bash file with export lines can be sourced
	//normally within a terminal and parsed correctly by this package.
	//An empty Export value means that no keyword prefix is allowed.
	//Export is set to DefaultExport by NewSourcer().
	Export string

	//Unquote is a function that is called to unquote a variable's value definition
	//if the value starts and ends with Quote.
	//It must not be nil if any variables have the surrounding Quotes.
	//Unquote is set to strconv.Unquote by NewSourcer().
	Unquote func(s string) (t string, err error)
}

//NewSourcer returns a Sourcer with Comment, Quote, Export, and Unquote set to
//DefaultComment, DefaultQuote, DefaultExport, and strconv.Unquote respectively.
func NewSourcer() *Sourcer {
	return &Sourcer{
		Comment: DefaultComment,
		Quote:   DefaultQuote,
		Export:  DefaultExport,
		Unquote: strconv.Unquote,
	}
}

//SourceFile attempts to parse and set all variable definitions in the file at path.
//If os.Open() errors, then that error is returned immediately.
//If an error occurs while parsing or setting values, then an *ErrSourcing is returned.
//The opened file is then closed and that possible error returned.
//SourceFile uses s.Source() to do the work on the file.
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

//Source attempts to parse and set all variable definitions from in.
//As soon as an error occurs while parsing or setting values, then that
//*ErrSourcing is returned and reading stops.
//Therefore, Source is not guaranteed to read all of in.
//Upon completion with a nil return value, all parsed name, value associations
//will have been called in os.Setenv().
func (s *Sourcer) Source(in io.Reader) error {
	return s.sourceVisitor(in, os.Setenv)
}

//not guaranteed to read all of in.

//NameVars attempts parse and return all variable definitions from in.
//As soon as an error occurs while parsing or setting values, then that
//*ErrSourcing is returned and reading stops.
//Therefore, NameVars is not guaranteed to read all of in.
//The return value nameVars will contain all name, value associations found from
//in with name at array index 0 and value at index 1.
func (s *Sourcer) NameVars(in io.Reader) (nameVars [][2]string, err error) {
	result := [][2]string{}
	err = s.sourceVisitor(in, func(name, v string) error {
		result = append(result, [2]string{name, v})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

//sourceVisitor actually does the work of reading from in using a bufio.Scanner
//to read, parse, and visit all lines from in.
func (s *Sourcer) sourceVisitor(in io.Reader, visit func(name, v string) error) error {
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
		if err := visit(name, v); err != nil {
			return &ErrSourcing{lineNumber, err}
		}
	}
	return scanner.Err()
}

//NameVar attempts to parse a single line and return the name, value association
//found.
//NameVar will return one of the errors in this package if a parsing error occurs.
//Note that ErrSourcing will never be returned from this method since this method
//simply parses and does not know about the purpose of the return values.
//The error ErrEmptyLine will be returned with empty name and v if line contains
//only whitespace or whitespace and a comment.
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
		if len(line) == 0 || (strings.HasPrefix(line, s.Comment) && s.Comment != "") {
			return "", "", ErrEmptyLine
		}
		return "", "", ErrNonVariableLine(origLine)
	}

	//get name and varible parts of the line. trim the name.
	name, v = strings.TrimLeft(line[:equalIndex], SpaceTab), line[equalIndex+1:]

	//if a comment appears at the beginning name (before Equal) then it is a comment line.
	if strings.HasPrefix(strings.TrimLeft(line, SpaceTab), s.Comment) && s.Comment != "" {
		return "", "", ErrEmptyLine
	}

	//evaluate name for errors.
	if s.isNameInvalid(name) {
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
	if strings.HasPrefix(v, s.Quote) && s.Quote != "" {
		//if starts and ends with quote but not equal to quote.
		if strings.HasSuffix(v, s.Quote) && v != s.Quote {
			return s.Unquote(v)
		}
		return "", &ErrValueUnclosedQuote{origV, s.Quote}
	}

	//if there is a comment, then get rid of it.
	commentIndex := strings.Index(v, s.Comment)
	if commentIndex >= 0 && s.Comment != "" {
		v = v[:commentIndex]
	}
	//trim any right whitespace.
	v = strings.TrimRight(v, SpaceTab)

	if v != strings.TrimLeft(v, SpaceTab) {
		return "", ErrInvalidWhitespaceValuePrefix(origV)
	}

	return v, nil
}
