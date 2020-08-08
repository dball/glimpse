package printer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dball/mal/glimpse/runtime"
	"github.com/dball/mal/glimpse/types"
)

// Config controls printing behavior
type Config struct {
	Readably     bool
	MaxSeqLength int
}

// PrintStr prints values
func PrintStr(config Config, value types.MalType) string {
	switch v := value.(type) {
	case types.Integer:
		return strconv.FormatInt(int64(v), 10)
	case types.Symbol:
		return v.Name
	case types.List:
		return printSeq(config, v.Seq(), "(", ")")
	case types.Vector:
		return printSeq(config, v.Seq(), "[", "]")
	case types.Map:
		return printMap(config, v)
	case types.String:
		return printString(config, v)
	case types.Function:
		return "#FN"
	case types.Keyword:
		return ":" + v.Name
	case types.Boolean:
		if v {
			return "true"
		}
		return "false"
	case types.Nil:
		return "nil"
	case *types.Atom:
		return "(atom " + PrintStr(config, v.Value) + ")"
	case types.Seq:
		// TODO config length
		seq, rest, _ := runtime.TakeDrop(types.Integer(10), v)
		empty, _ := runtime.Empty(rest)
		var last = ")"
		if !empty {
			last = " ... )"
		}
		return printSeq(config, seq, "(", last)
	case types.MalError:
		return PrintStr(config, v.Reason)
	case error:
		return printString(config, types.String(v.Error()))
	default:
		return fmt.Sprintf("#UNKNOWN: %v", value)
	}
}

func printSeq(config Config, seq types.Seq, first string, last string) string {
	var sb strings.Builder
	sb.WriteString(first)
	i := 0
	for {
		empty, head, tail := seq.Next()
		if empty {
			break
		}
		if i > 0 {
			sb.WriteRune(' ')
		}
		i++
		sb.WriteString(PrintStr(config, head))
		seq = tail
	}
	sb.WriteString(last)
	return sb.String()
}

func printMap(config Config, m types.Map) string {
	var sb strings.Builder
	sb.WriteRune('{')
	var i int
	imm := m.Imm
	itr := imm.Iterator()
	for !itr.Done() {
		if i > 0 {
			sb.WriteRune(' ')
		}
		i++
		k, v := itr.Next()
		sb.WriteString(PrintStr(config, k))
		sb.WriteRune(' ')
		sb.WriteString(PrintStr(config, v))
	}
	sb.WriteRune('}')
	return sb.String()
}

// When print_readably is true, doublequotes, newlines, and backslashes are translated into their printed representations (the reverse of the reader)
func printString(config Config, s types.String) string {
	if !config.Readably {
		return string(s)
	}
	var sb strings.Builder
	sb.WriteRune('"')
	runes := []rune(s)
	for _, r := range runes {
		switch r {
		case '"':
			sb.WriteString(`\"`)
		case '\\':
			sb.WriteString(`\\`)
		case '\n':
			sb.WriteString(`\n`)
		default:
			sb.WriteRune(r)
		}
	}
	sb.WriteRune('"')
	return sb.String()
}
