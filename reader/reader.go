package reader

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/dball/mal/glimpse/types"
)

var tokenRegexp = regexp.MustCompile(`[\s,]*(~@|[\[\]{}()'` + "`" +
	`~^@]|"(?:\\.|[^\\"])*"?|;.*|[^\s\[\]{}('"` + "`" +
	`,;)]*)`)

var integerRegexp = regexp.MustCompile(`^-?\d+$`)

// Reader reads tokens
type Reader struct {
	tokens []string
	offset int
}

// Error is a reader error
type Error struct {
	Message string
	Err     error
}

func (err Error) Unwrap() error { return err.Err }

func (err Error) String() string {
	return fmt.Sprintf("reader error: %v: %v", err.Message, err.Err)
}
func (err Error) Error() string {
	return fmt.Sprintf("reader error: %v: %v", err.Message, err.Err)
}

// Comment is an error indicating no token
type Comment struct{}

func (Comment) Error() string {
	return "Comment token"
}

func (reader *Reader) peek() *string {
	if reader.offset == len(reader.tokens) {
		return nil
	}
	return &reader.tokens[reader.offset]
}

func (reader *Reader) next() *string {
	token := reader.peek()
	if token != nil {
		reader.offset++
	}
	return token
}

func tokenize(s string) []string {
	matches := tokenRegexp.FindAllStringSubmatch(s, -1)
	tokens := make([]string, len(matches))
	for i, match := range matches {
		tokens[i] = match[1]
	}
	return tokens
}

// ReadStr reads strings
func ReadStr(s string) (types.MalType, error) {
	return readForm(&Reader{tokenize(s), 0})
}

func readForm(reader *Reader) (types.MalType, error) {
Loop:
	for {
		token := reader.peek()
		if token == nil {
			return nil, Error{"Unexpected end of input reading form", nil}
		}
		switch *token {
		case "(":
			reader.next()
			return readList(reader, ")", types.NewList())
		case "[":
			reader.next()
			return readList(reader, "]", types.NewVector())
		case "{":
			reader.next()
			return readList(reader, "}", types.NewMap())
		case "'":
			return readQuotedForm(reader, "quote")
		case "`":
			return readQuotedForm(reader, "quasiquote")
		case "~":
			return readQuotedForm(reader, "unquote")
		case "~@":
			return readQuotedForm(reader, "splice-unquote")
		case "@":
			return readQuotedForm(reader, "deref")
		default:
			val, err := readAtom(reader)
			if err != nil {
				_, comment := err.(Comment)
				if comment {
					continue Loop
				}
				return nil, err
			}
			return val, err
		}
	}
}

func readQuotedForm(reader *Reader, name string) (types.MalType, error) {
	reader.next()
	form, err := readForm(reader)
	if err != nil {
		return nil, Error{"Unexpected end of quoted form: " + name, err}
	}
	return types.NewList(types.NewSymbol(name), form), nil
}

func readList(reader *Reader, end string, coll types.MalType) (types.MalType, error) {
	var items []types.MalType
Loop:
	for {
		value, err := readForm(reader)
		if err != nil {
			return coll, Error{"Error reading list", err}
		}
		switch value {
		case types.Symbol{Name: end}:
			break Loop
		case nil:
			return coll, Error{"Unexpected end of input reading list", nil}
		default:
			items = append(items, value)
		}
	}
	switch coll.(type) {
	case types.List:
		return types.NewList(items...), nil
	case types.Vector:
		return types.NewVector(items...), nil
	case types.Map:
		if len(items)%2 != 0 {
			return coll, Error{"Unbalanced map input", nil}
		}
		return types.NewMap(items...), nil
	default:
		return nil, Error{"Invalid list type", nil}
	}
}

func readAtom(reader *Reader) (types.MalType, error) {
	token := *reader.next()
	if integerRegexp.MatchString(token) {
		value, err := strconv.ParseInt(token, 10, 64)
		if err != nil {
			return nil, Error{"Unparseable integer", err}
		}
		return types.Integer(value), nil
	}
	runes := []rune(token)
	switch runes[0] {
	case ';':
		return nil, Comment{}
	case '"':
		return parseString(runes)
	case ':':
		return types.NewKeyword(string(runes[1:])), nil
	default:
		switch token {
		case "true":
			return types.Boolean(true), nil
		case "false":
			return types.Boolean(false), nil
		case "nil":
			return types.Nil{}, nil
		default:
			return types.NewSymbol(token), nil
		}
	}
}

func parseString(runes []rune) (types.MalType, error) {
	last := len(runes) - 1
	if last == 0 || runes[last] != '"' {
		return nil, Error{"String quotes are unbalanced", nil}
	}
	var result []rune
	var escaping bool
	for _, r := range runes[1:last] {
		if !escaping {
			if r == '\\' {
				escaping = true
			} else {
				result = append(result, r)
			}
		} else {
			switch r {
			case '\\':
				result = append(result, r)
			case '"':
				result = append(result, r)
			case 'n':
				result = append(result, '\n')
			default:
				return nil, Error{"String escape sequence is invalid", nil}
			}
			escaping = false
		}
	}
	if escaping {
		return nil, Error{"String slashes are unbalanced", nil}
	}
	return types.String(string(result)), nil
}
