package core

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/benbjohnson/immutable"
	"github.com/dball/glimpse/printer"
	"github.com/dball/glimpse/reader"
	"github.com/dball/glimpse/runtime"
	"github.com/dball/glimpse/types"
)

func intList(items []types.MalType) ([]int64, error) {
	var ints []int64
	for _, item := range items {
		i, valid := item.(types.Integer)
		if !valid {
			return ints, errors.New("non-integer found")
		}
		ints = append(ints, int64(i))
	}
	return ints, nil
}

// BuildEnv builds and returns a new environment with core vars
func BuildEnv() *types.Env {
	var env = types.BuildEnv()
	env.Set("+", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			ints, err := intList(args)
			if err != nil {
				return nil, err
			}
			var sum int64 = 0
			for _, i := range ints {
				sum += i
			}
			return types.Integer(sum), nil
		},
	})
	env.Set("-", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			ints, err := intList(args)
			if err != nil {
				return nil, err
			}
			if len(ints) == 1 {
				return -ints[0], nil
			}
			var sum int64 = ints[0]
			for _, i := range ints[1:] {
				sum -= i
			}
			return types.Integer(sum), nil
		},
	})
	env.Set("*", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			ints, err := intList(args)
			if err != nil {
				return nil, err
			}
			var sum int64 = 1
			for _, i := range ints {
				sum *= i
			}
			return types.Integer(sum), nil
		},
	})
	env.Set("/", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			ints, err := intList(args)
			if err != nil {
				return nil, err
			}
			if len(ints) == 1 {
				return types.Integer(1 / ints[0]), nil
			}
			var sum int64 = ints[0]
			for _, i := range ints[1:] {
				sum /= i
			}
			return types.Integer(sum), nil
		},
	})
	env.Set("list", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			return types.NewList(args...), nil
		},
	})
	env.Set("list?", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			if len(args) != 1 {
				return nil, errors.New("list? requires 1 arg")
			}
			_, valid := args[0].(types.List)
			return types.Boolean(valid), nil
		},
	})
	env.Set("empty?", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			if len(args) != 1 {
				return nil, errors.New("empty? requires 1 arg")
			}
			return runtime.Empty(args[0])
		},
	})
	env.Set("count", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			if len(args) != 1 {
				return nil, errors.New("count requires 1 arg")
			}
			var count int64
			// TODO We could count seqs and seqables
			switch coll := args[0].(type) {
			case types.Counted:
				count = int64(coll.Count())
			default:
				return nil, errors.New("count requires a countable collection")
			}
			return types.Integer(count), nil
		},
	})
	env.Set("=", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			if len(args) == 0 {
				return types.Boolean(true), nil
			}
			this := args[0]
			for _, that := range args[1:] {
				if !types.Equals(this, that) {
					return types.Boolean(false), nil
				}
			}
			return types.Boolean(true), nil
		},
	})
	env.Set(">=", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			if len(args) == 0 {
				return nil, errors.New(">= requires at least one arg")
			}
			this := args[0]
			for _, that := range args[1:] {
				comp, err := types.Compare(this, that)
				if err != nil {
					return nil, err
				}
				switch comp {
				case -1:
					return types.Boolean(false), nil
				}
				this = that
			}
			return types.Boolean(true), nil
		},
	})
	env.Set(">", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			if len(args) == 0 {
				return nil, errors.New("> requires at least one arg")
			}
			this := args[0]
			for _, that := range args[1:] {
				comp, err := types.Compare(this, that)
				if err != nil {
					return nil, err
				}
				switch comp {
				case -1:
					return types.Boolean(false), nil
				case 0:
					return types.Boolean(false), nil
				}
				this = that
			}
			return types.Boolean(true), nil
		},
	})
	env.Set("<=", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			if len(args) == 0 {
				return nil, errors.New("<= requires at least one arg")
			}
			this := args[0]
			for _, that := range args[1:] {
				comp, err := types.Compare(this, that)
				if err != nil {
					return nil, err
				}
				switch comp {
				case 1:
					return types.Boolean(false), nil
				}
				this = that
			}
			return types.Boolean(true), nil
		},
	})
	env.Set("<", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			if len(args) == 0 {
				return nil, errors.New("< requires at least one arg")
			}
			this := args[0]
			for _, that := range args[1:] {
				comp, err := types.Compare(this, that)
				if err != nil {
					return nil, err
				}
				switch comp {
				case 1:
					return types.Boolean(false), nil
				case 0:
					return types.Boolean(false), nil
				}
				this = that
			}
			return types.Boolean(true), nil
		},
	})
	env.Set("pr-str", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			var sb strings.Builder
			for i, arg := range args {
				if i > 0 {
					sb.WriteRune(' ')
				}
				sb.WriteString(printer.PrintStr(printer.Config{Readably: true}, arg))
			}
			return types.String(sb.String()), nil
		},
	})
	env.Set("str", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			var sb strings.Builder
			for _, arg := range args {
				sb.WriteString(printer.PrintStr(printer.Config{Readably: false}, arg))
			}
			return types.String(sb.String()), nil
		},
	})
	env.Set("prn", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			var sb strings.Builder
			for i, arg := range args {
				if i > 0 {
					sb.WriteRune(' ')
				}
				sb.WriteString(printer.PrintStr(printer.Config{Readably: true}, arg))
			}
			sb.WriteRune('\n')
			os.Stdout.WriteString(sb.String())
			return types.Nil{}, nil
		},
	})
	env.Set("println", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			var sb strings.Builder
			for i, arg := range args {
				if i > 0 {
					sb.WriteRune(' ')
				}
				sb.WriteString(printer.PrintStr(printer.Config{Readably: false}, arg))
			}
			sb.WriteRune('\n')
			os.Stdout.WriteString(sb.String())
			return types.Nil{}, nil
		},
	})
	env.Set("read-string", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			if len(args) != 1 {
				return nil, errors.New("read-string requires one arg")
			}
			s, valid := args[0].(types.String)
			if !valid {
				return nil, errors.New("read-string requires a string arg")
			}
			return reader.ReadStr(string(s))
		},
	})
	env.Set("slurp", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			if len(args) != 1 {
				return nil, errors.New("read-string requires one arg")
			}
			s, valid := args[0].(types.String)
			if !valid {
				return nil, errors.New("read-string requires a string arg")
			}
			bytes, err := ioutil.ReadFile(string(s))
			if err != nil {
				return nil, err
			}
			return types.String(string(bytes)), nil
		},
	})
	env.Set("atom", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			// TODO validate one value
			return &types.Atom{Value: args[0]}, nil
		},
	})
	env.Set("atom?", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			// TODO validate one value
			_, valid := args[0].(*types.Atom)
			return types.Boolean(valid), nil
		},
	})
	env.Set("deref", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			// TODO validate one value
			atom, valid := args[0].(*types.Atom)
			if !valid {
				return nil, errors.New("deref requires an atom value")
			}
			return atom.Value, nil
		},
	})
	env.Set("reset!", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			// TODO validate two value
			atom, valid := args[0].(*types.Atom)
			if !valid {
				return nil, errors.New("deref requires an atom value")
			}
			value := args[1]
			atom.Set(value)
			return value, nil
		},
	})
	env.Set("swap!", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			// TODO validate
			atom, valid := args[0].(*types.Atom)
			if !valid {
				return nil, errors.New("swap! requires atom value")
			}
			fn, valid := args[1].(types.Function)
			swapArgs := []types.MalType{atom.Value}
			if len(args) > 2 {
				swapArgs = append(swapArgs, args[2:]...)
			}
			value, error := fn.Fn(swapArgs...)
			if error != nil {
				return nil, error
			}
			atom.Set(value)
			return value, nil
		},
	})
	env.Set("seq", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			// TODO validate one value
			return runtime.Seq(args[0])
		},
	})
	env.Set("empty?", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			// TODO validate one value
			return runtime.Empty(args[0])
		},
	})
	env.Set("first", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			seq, err := runtime.Seq(args[0])
			if err != nil {
				return nil, err
			}
			empty, head, _ := seq.Next()
			if empty {
				return types.Nil{}, nil
			}
			return head, nil
		},
	})
	env.Set("rest", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			seq, err := runtime.Seq(args[0])
			if err != nil {
				return nil, err
			}
			empty, _, tail := seq.Next()
			if empty {
				return types.List{}, nil
			}
			return tail, nil
		},
	})
	env.Set("take", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			seq, _, err := runtime.TakeDrop(args[0], args[1])
			if err != nil {
				return nil, err
			}
			return seq, nil
		},
	})
	env.Set("cons", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			seq, err := runtime.Seq(args[1])
			if err != nil {
				return nil, err
			}
			return types.ConsCell{Head: args[0], Tail: seq}, nil
		},
	})
	env.Set("concat", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			seq, err := runtime.Concat(args...)
			if err != nil {
				return nil, err
			}
			return seq, nil
		},
	})
	env.Set("conj", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			conjed, err := runtime.Conj(args[0], args[1:]...)
			if err != nil {
				return nil, err
			}
			return conjed, nil
		},
	})
	env.Set("into", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			conjed, err := runtime.Into(args[0], args[1])
			if err != nil {
				return nil, err
			}
			return conjed, nil
		},
	})
	env.Set("nth", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			return runtime.Nth(args[0], args[1])
		},
	})
	env.Set("throw", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			return nil, types.MalError{Reason: args[0]}
		},
	})
	env.Set("symbol?", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			_, valid := args[0].(types.Symbol)
			return types.Boolean(valid), nil
		},
	})
	env.Set("symbol", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			name, valid := args[0].(types.String)
			if !valid {
				return nil, errors.New("Invalid arg")
			}
			return types.NewSymbol(string(name)), nil
		},
	})
	env.Set("keyword?", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			_, valid := args[0].(types.Keyword)
			return types.Boolean(valid), nil
		},
	})
	env.Set("keyword", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			var name string
			switch v := args[0].(type) {
			case types.String:
				name = string(v)
			case types.Keyword:
				name = v.Name
			case types.Symbol:
				name = v.Name
			default:
				return nil, errors.New("invalid")
			}
			return types.NewKeyword(name), nil
		},
	})
	env.Set("nil?", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			_, valid := args[0].(types.Nil)
			return types.Boolean(valid), nil
		},
	})
	env.Set("true?", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			b, valid := args[0].(types.Boolean)
			if !valid {
				return types.Boolean(false), nil
			}
			return b, nil
		},
	})
	env.Set("false?", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			b, valid := args[0].(types.Boolean)
			if !valid {
				return types.Boolean(false), nil
			}
			return types.Boolean(!bool(b)), nil
		},
	})
	env.Set("sequential?", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			_, valid := args[0].(types.Sequential)
			return types.Boolean(valid), nil
		},
	})
	env.Set("vector?", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			_, valid := args[0].(types.Vector)
			return types.Boolean(valid), nil
		},
	})
	env.Set("map?", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			_, valid := args[0].(types.Map)
			return types.Boolean(valid), nil
		},
	})
	// TODO lazy seq
	env.Set("map", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			fn, valid := args[0].(types.Function)
			if !valid {
				return nil, errors.New("Invalid")
			}
			seq, err := runtime.Seq(args[1])
			if err != nil {
				return nil, err
			}
			var items []types.MalType
			for {
				empty, head, tail := seq.Next()
				if empty {
					return types.NewList(items...), nil
				}
				item, err := fn.Fn(head)
				if err != nil {
					return nil, err
				}
				items = append(items, item)
				seq = tail
			}
		},
	})
	env.Set("apply", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			total := len(args)
			if total < 2 {
				return nil, errors.New("invalid")
			}
			fn, valid := args[0].(types.Function)
			if !valid {
				return nil, errors.New("Invalid")
			}
			fnargs := args[1:(total - 1)]
			seq, err := runtime.Seq(args[total-1])
			if err != nil {
				return nil, err
			}
			for {
				empty, head, tail := seq.Next()
				if empty {
					break
				}
				fnargs = append(fnargs, head)
				seq = tail
			}
			if err != nil {
				return nil, err
			}
			return fn.Fn(fnargs...)
		},
	})
	env.Set("vector", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			return types.NewVector(args), nil
		},
	})
	env.Set("hash-map", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			if len(args)%2 != 0 {
				return nil, errors.New("invalid")
			}
			return types.NewMap(args...), nil
		},
	})
	env.Set("assoc", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			m, valid := args[0].(types.Map)
			if !valid {
				return nil, errors.New("invalid")
			}
			if len(args)%2 != 1 {
				return nil, errors.New("invalid")
			}
			b := immutable.NewMapBuilder(m.Imm)
			for i := 1; i < len(args); i += 2 {
				b.Set(args[1], args[i+1])
			}
			return types.Map{Imm: b.Map()}, nil
		},
	})
	env.Set("dissoc", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			m, valid := args[0].(types.Map)
			if !valid {
				return nil, errors.New("invalid")
			}
			b := immutable.NewMapBuilder(m.Imm)
			for _, k := range args[1:] {
				b.Delete(k)
			}
			return types.Map{Imm: b.Map()}, nil
		},
	})
	env.Set("get", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			var notfound types.MalType
			if len(args) == 2 {
				notfound = types.Nil{}
			} else {
				notfound = args[2]
			}
			return runtime.Get(args[0], args[1], notfound), nil
		},
	})
	env.Set("contains?", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			return runtime.Contains(args[0], args[1]), nil
		},
	})
	env.Set("keys", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			return runtime.Keys(args[0])
		},
	})
	env.Set("vals", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			return runtime.Vals(args[0])
		},
	})
	env.Set("hash", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			return types.Integer(types.Hash(args[0])), nil
		},
	})
	env.Set("with-meta", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			return runtime.WithMeta(args[0], args[1])
		},
	})
	env.Set("meta", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			return runtime.Meta(args[0])
		},
	})
	env.Set("time-ms", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			return types.Integer(time.Now().Unix()), nil
		},
	})
	env.Set("string?", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			_, valid := args[0].(types.String)
			return types.Boolean(valid), nil
		},
	})
	env.Set("number?", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			_, valid := args[0].(types.Integer)
			return types.Boolean(valid), nil
		},
	})
	env.Set("fn?", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			fn, valid := args[0].(types.Function)
			return types.Boolean(valid && !fn.IsMacro), nil
		},
	})
	env.Set("macro?", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			fn, valid := args[0].(types.Function)
			return types.Boolean(valid && fn.IsMacro), nil
		},
	})
	env.Set("range", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			return runtime.Range(args...)
		},
	})

	/*
		env.Set(types.Symbol{Name: ""}, types.Function{
			Fn: func(args ...types.MalType) (types.MalType, error) {
			},
		})
	*/
	return env
}
