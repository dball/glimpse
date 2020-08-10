package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/benbjohnson/immutable"
	"github.com/dball/glimpse/core"
	"github.com/dball/glimpse/printer"
	"github.com/dball/glimpse/reader"
	"github.com/dball/glimpse/runtime"
	"github.com/dball/glimpse/types"
	"github.com/peterh/liner"
)

func evalAst(evalEnv *types.Env, form types.MalType) (types.MalType, error) {
	switch value := form.(type) {
	case types.Symbol:
		v, err := evalEnv.Get(value.Name)
		if err != nil {
			return nil, err
		}
		return v, nil
	case types.List:
		items := make([]types.MalType, value.Imm.Len())
		itr := value.Imm.Iterator()
		for !itr.Done() {
			i, v := itr.Next()
			item, err := EVAL(evalEnv, v)
			if err != nil {
				return nil, err
			}
			items[i] = item
		}
		return types.NewList(items...), nil
	case types.Vector:
		items := make([]types.MalType, value.Imm.Len())
		itr := value.Imm.Iterator()
		for !itr.Done() {
			i, v := itr.Next()
			item, err := EVAL(evalEnv, v)
			if err != nil {
				return nil, err
			}
			items[i] = item
		}
		return types.NewVector(items...), nil
	case types.Map:
		itr := value.Imm.Iterator()
		m2 := types.NewMap()
		b := immutable.NewMapBuilder(m2.Imm)
		for !itr.Done() {
			k, v := itr.Next()
			k2, err := EVAL(evalEnv, k)
			if err != nil {
				return nil, err
			}
			v2, err := EVAL(evalEnv, v)
			if err != nil {
				return nil, err
			}
			b.Set(k2, v2)
		}
		return types.Map{Imm: b.Map()}, nil
	default:
		return value, nil
	}
}

// READ reads
func READ(s string) (types.MalType, error) {
	return reader.ReadStr(s)
}

func isPair(form types.MalType) bool {
	_, seq := form.(types.Seq)
	if !seq {
		_, sequential := form.(types.Sequential)
		if !sequential {
			return false
		}
	}
	empty, err := runtime.Empty(form)
	return (err == nil) && !bool(empty)
}

func quasiquote(form types.MalType) types.MalType {
	if !isPair(form) {
		return types.NewList(types.NewSymbol("quote"), form)
	}
	seq, _ := runtime.Seq(form)
	_, head, tail := seq.Next()
	symbol, valid := head.(types.Symbol)
	if valid && symbol.Name == "unquote" {
		_, ihead, _ := tail.Next()
		return ihead
	}
	if isPair(head) {
		iseq, _ := runtime.Seq(head)
		_, ihead, itail := iseq.Next()
		isymbol, valid := ihead.(types.Symbol)
		if valid && isymbol.Name == "splice-unquote" {
			_, iihead, _ := itail.Next()
			return types.NewList(types.NewSymbol("concat"), iihead, quasiquote(tail))
		}
	}
	return types.NewList(types.NewSymbol("cons"), quasiquote(head), quasiquote(tail))
}

func isMacroCall(evalEnv *types.Env, form types.MalType) (types.Function, types.Seq, bool) {
	var fn types.Function
	var args types.Seq
	if !isPair(form) {
		return fn, args, false
	}
	seq, err := runtime.Seq(form)
	if err != nil {
		return fn, args, false
	}
	_, head, tail := seq.Next()
	symbol, valid := head.(types.Symbol)
	if !valid {
		return fn, args, false
	}
	val, err := evalEnv.Get(symbol.Name)
	if err != nil {
		return fn, args, false
	}
	fn, valid = val.(types.Function)
	if !valid {
		return fn, args, false
	}
	return fn, tail, fn.IsMacro
}

func macroexpand(evalEnv *types.Env, form types.MalType) (types.MalType, error) {
	for {
		macro, args, valid := isMacroCall(evalEnv, form)
		if !valid {
			return form, nil
		}
		items, err := runtime.IntoSlice(args)
		if err != nil {
			return nil, err
		}
		expanded, err := macro.Fn(items...)
		if err != nil {
			return nil, err
		}
		form = expanded
	}
}

// EVAL evals
func EVAL(evalEnv *types.Env, form types.MalType) (types.MalType, error) {
	for {
		applicable, isApplicable := form.(types.Applicable)
		if isApplicable {
			items, err := runtime.IntoSlice(applicable.Seq())
			if err != nil {
				return nil, err
			}
			if len(items) == 0 {
				return types.NewList(), nil
			}
			expanded, err := macroexpand(evalEnv, types.NewList(items...))
			if err != nil {
				return nil, err
			}
			form = expanded
			_, stillApplicable := form.(types.Applicable)
			if !stillApplicable {
				return evalAst(evalEnv, form)
			}
		}
		switch value := form.(type) {
		case types.Applicable:
			items, err := runtime.IntoSlice(value.Seq())
			if err != nil {
				return nil, err
			}
			if len(items) == 0 {
				return value, nil
			}
			switch items[0] {
			case types.Symbol{Name: "def!"}:
				if len(items) != 3 {
					return nil, errors.New("def! requires 2 args")
				}
				symbol, valid := items[1].(types.Symbol)
				if !valid {
					return nil, errors.New("def! requires a symbol arg")
				}
				val, err := EVAL(evalEnv, items[2])
				if err != nil {
					return nil, err
				}
				evalEnv.Set(symbol.Name, val)
				return val, nil
			case types.Symbol{Name: "defmacro!"}:
				if len(items) != 3 {
					return nil, errors.New("defmacro! requires 2 args")
				}
				symbol, valid := items[1].(types.Symbol)
				if !valid {
					return nil, errors.New("defmacro! requires a symbol arg")
				}
				val, err := EVAL(evalEnv, items[2])
				if err != nil {
					return nil, err
				}
				fn, valid := val.(types.Function)
				if !valid {
					return nil, errors.New("defmacro! requires a macro arg")
				}
				fn.IsMacro = true
				evalEnv.Set(symbol.Name, fn)
				return fn, nil
			case types.Symbol{Name: "let*"}:
				if len(items) != 3 {
					return nil, errors.New("let* requires 2 args")
				}
				sequential, valid := items[1].(types.Sequential)
				if !valid {
					return nil, errors.New("let* requires a binding sequential arg")
				}
				bindings, err := runtime.IntoSlice(sequential)
				if len(bindings)%2 != 0 {
					return nil, errors.New("let* requires an even list of bindings")
				}
				inner, err := types.DeriveEnv(evalEnv, nil, nil)
				if err != nil {
					return nil, err
				}
				for i := 0; i < len(bindings); i += 2 {
					symbol, valid := bindings[i].(types.Symbol)
					if !valid {
						return nil, errors.New("let* binding arg requires a symbol")
					}
					val, err := EVAL(inner, bindings[i+1])
					if err != nil {
						return nil, err
					}
					inner, err = types.DeriveEnv(inner, []types.MalType{symbol}, []types.MalType{val})
					if err != nil {
						return nil, err
					}
				}
				evalEnv = inner
				form = items[2]
				continue
			case types.Symbol{Name: "do"}:
				forms := len(items) - 1
				if forms == 0 {
					return types.Nil{}, nil
				}
				for _, item := range items[1:forms] {
					_, err := EVAL(evalEnv, item)
					if err != nil {
						return nil, err
					}
				}
				form = items[forms]
				continue
			case types.Symbol{Name: "if"}:
				argl := len(items)
				if argl < 3 || argl > 4 {
					return nil, errors.New("if requires 2 or 3 args")
				}
				test, err := EVAL(evalEnv, items[1])
				if err != nil {
					return nil, err
				}
				var cond bool
				switch test {
				case types.Boolean(false):
					cond = false
				case types.Nil{}:
					cond = false
				default:
					cond = true
				}
				if cond {
					form = items[2]
				} else if argl == 4 {
					form = items[3]
				} else {
					return types.Nil{}, nil
				}
				continue
			case types.Symbol{Name: "fn*"}:
				if len(items) != 3 {
					return nil, errors.New("fn* requires 2 args")
				}
				sequential, valid := items[1].(types.Sequential)
				body := items[2]
				if !valid {
					return nil, errors.New("fn* requires a sequential args arg")
				}
				binds, err := runtime.IntoSlice(sequential)
				if err != nil {
					return nil, err
				}
				return types.Function{
					Fn: func(args ...types.MalType) (types.MalType, error) {
						fnEnv, err := types.DeriveEnv(evalEnv, binds, args)
						if err != nil {
							return nil, err
						}
						return EVAL(fnEnv, body)
					},
					Body:  body,
					Binds: binds,
					Env:   evalEnv,
				}, nil
			case types.Symbol{Name: "quote"}:
				if len(items) != 2 {
					return nil, errors.New("quote requires 1 arg")
				}
				return items[1], nil
			case types.Symbol{Name: "quasiquote"}:
				form = quasiquote(items[1])
				continue
			case types.Symbol{Name: "macroexpand"}:
				return macroexpand(evalEnv, items[1])
			case types.Symbol{Name: "try*"}:
				tryBody := items[1]
				result, err := EVAL(evalEnv, tryBody)
				if err == nil {
					return result, nil
				}
				if len(items) == 2 {
					return nil, err
				}
				catchForm := items[2]
				applicable, valid := catchForm.(types.Applicable)
				if !valid {
					return nil, errors.New("Invalid try* form")
				}
				catchItems, err := runtime.IntoSlice(applicable.Seq())
				if err != nil {
					return nil, err
				}
				symbol, valid := catchItems[0].(types.Symbol)
				if !valid || symbol.Name != "catch*" {
					return nil, errors.New("Invalid try* form")
				}
				catchEnv, err := types.DeriveEnv(evalEnv, catchItems[1:2], []types.MalType{err})
				if err != nil {
					return nil, err
				}
				return EVAL(catchEnv, catchItems[2])
			default:
				evaluated, err := evalAst(evalEnv, value)
				if err != nil {
					return nil, err
				}
				applicable, valid := evaluated.(types.Applicable)
				if !valid {
					return nil, errors.New("List did not eval to list")
				}
				iitems, err := runtime.IntoSlice(applicable.Seq())
				if err != nil {
					return nil, err
				}
				fn, valid := iitems[0].(types.Function)
				if !valid {
					return nil, errors.New("No function found in first position")
				}
				if fn.Body == nil {
					return fn.Fn(iitems[1:]...)
				}
				//a fn* value: set ast to the ast attribute of f. Generate a new
				//environment using the env and params attributes of f as the outer and
				//binds arguments and args as the exprs argument. Set env to the new
				//environment. Continue at the beginning of the loop.
				form = fn.Body
				fnEnv, err := types.DeriveEnv(fn.Env, fn.Binds, iitems[1:])
				if err != nil {
					return nil, err
				}
				evalEnv = fnEnv
				continue
			}
		default:
			return evalAst(evalEnv, form)
		}
	}
}

// PRINT prints
func PRINT(value types.MalType) string {
	return printer.PrintStr(printer.Config{Readably: true}, value)
}

func rep(env *types.Env, s string) string {
	form, err := READ(s)
	if err != nil {
		return err.Error()
	}
	val, err := EVAL(env, form)
	if err != nil {
		return "#ERROR: " + PRINT(err)
	}
	return PRINT(val)
}

func interactiveRepl2(env *types.Env) {
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	historyFile := filepath.Join(os.TempDir(), ".glimpse-history")
	if f, err := os.Open(historyFile); err == nil {
		line.ReadHistory(f)
		f.Close()
	}
	for {
		text, err := line.Prompt("user> ")
		if err == nil {
			line.AppendHistory(text)
			os.Stdout.WriteString(rep(env, text))
			os.Stdout.WriteString("\n")
		} else if err == liner.ErrPromptAborted {
		} else if err == io.EOF {
			break
		} else {
			log.Fatalf("liner err %v", err)
		}
		if f, err := os.Create(historyFile); err == nil {
			line.WriteHistory(f)
			f.Close()
		}
	}
}

func main() {
	env := core.BuildEnv()
	env.Set("*host-language*", types.String("glimpse"))
	env.Set("eval", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			if len(args) != 1 {
				return nil, errors.New("eval requires 1 arg")
			}
			return EVAL(env, args[0])
		},
	})
	env.Set("readline", types.Function{
		Fn: func(args ...types.MalType) (types.MalType, error) {
			if len(args) != 1 {
				return nil, errors.New("sad")
			}
			s, valid := args[0].(types.String)
			if !valid {
				return nil, errors.New("mad")
			}
			os.Stdout.WriteString(string(s))
			scanner := bufio.NewScanner(os.Stdin)
			if !scanner.Scan() {
				return types.Nil{}, nil
			}
			return types.String(scanner.Text()), nil
		},
	})
	rep(env, "(def! not (fn* (a) (if a false true)))")
	rep(env, `(def! load-file (fn* (f) (eval (read-string (str "(do " (slurp f) "\nnil)")))))`)
	rep(env, `(defmacro! cond (fn* (& xs) (if (> (count xs) 0) (list 'if (first xs) (if (> (count xs) 1) (nth xs 1) (throw "odd number of forms to cond")) (cons 'cond (rest (rest xs)))))))`)
	var args []types.MalType
	for _, arg := range os.Args[1:] {
		args = append(args, types.String(arg))
	}
	if len(args) == 0 {
		env.Set("*ARGV*", types.List{})
		interactiveRepl2(env)
	} else {
		env.Set("*ARGV*", types.NewList(args[1:]...))
		var items []types.MalType
		items = append(items, types.Symbol{Name: "load-file"}, args[0])
		form := types.NewList(items...)
		_, err := EVAL(env, form)
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("error: %v", err))
			os.Exit(1)
		}
	}
}
