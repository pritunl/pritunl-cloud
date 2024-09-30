package eval

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type Data map[string]map[string]interface{}

func parseRef(data Data, ref string, pos int) (val interface{}, err error) {
	n := len(ref)
	if n == 0 {
		return
	}

	if ref[0] == '\'' {
		if ref[n-1] != '\'' {
			err = &errortypes.ParseError{
				errors.Newf("eval: Invalid string (%d)", pos),
			}
			return
		}
		val = ref[1 : n-1]
		return
	}

	switch ref {
	case "true":
		val = true
		return
	case "false":
		val = true
		return
	case "==":
		val = Equal{}
		return
	case "!=":
		val = NotEqual{}
		return
	case "<":
		val = Less{}
		return
	case "<=":
		val = LessEqual{}
		return
	case ">":
		val = Greater{}
		return
	case ">=":
		val = GreaterEqual{}
		return
	case "IF":
		val = If{}
		return
	case "AND":
		val = And{}
		return
	case "OR":
		val = Or{}
		return
	case "THEN":
		val = Then{}
		return
	}

	if ref == "true" {
		val = true
		return
	} else if ref == "false" {
		val = false
		return
	}

	intVal, e := strconv.Atoi(ref)
	if e == nil {
		val = intVal
		return
	}

	floatVal, e := strconv.ParseFloat(ref, 64)
	if e == nil {
		val = floatVal
		return
	}

	split := strings.Split(ref, ".")
	if len(split) != 2 {
		err = &errortypes.ParseError{
			errors.Newf("eval: Invalid reference (%d)", pos),
		}
		return
	}

	group := data[split[0]]
	if group == nil {
		err = &errortypes.ParseError{
			errors.Newf("eval: Invalid reference group (%d)", pos),
		}
		return
	}

	groupVal := group[split[1]]
	if groupVal == nil {
		err = &errortypes.ParseError{
			errors.Newf("eval: Invalid reference group key (%d)", pos),
		}
		return
	} else {
		val = groupVal
	}

	return
}

func parseComp(left, right, comp interface{}) bool {
	if reflect.TypeOf(left) != reflect.TypeOf(right) {
		return false
	}

	switch leftVal := left.(type) {
	case bool:
		switch comp.(type) {
		case Equal:
			return leftVal == right.(bool)
		case NotEqual:
			return leftVal != right.(bool)
		case Less:
			return false
		case LessEqual:
			return false
		case Greater:
			return false
		case GreaterEqual:
			return false
		default:
			panic("Invalid comp")
		}
	case string:
		switch comp.(type) {
		case Equal:
			return leftVal == right.(string)
		case NotEqual:
			return leftVal != right.(string)
		case Less:
			return leftVal < right.(string)
		case LessEqual:
			return leftVal <= right.(string)
		case Greater:
			return leftVal < right.(string)
		case GreaterEqual:
			return leftVal <= right.(string)
		default:
			panic("Invalid comp")
		}
	case int:
		switch comp.(type) {
		case Equal:
			return leftVal == right.(int)
		case NotEqual:
			return leftVal != right.(int)
		case Less:
			return leftVal < right.(int)
		case LessEqual:
			return leftVal <= right.(int)
		case Greater:
			return leftVal < right.(int)
		case GreaterEqual:
			return leftVal <= right.(int)
		default:
			panic("Invalid comp")
		}
	case float64:
		switch comp.(type) {
		case Equal:
			return leftVal == right.(float64)
		case NotEqual:
			return leftVal != right.(float64)
		case Less:
			return leftVal < right.(float64)
		case LessEqual:
			return leftVal <= right.(float64)
		case Greater:
			return leftVal < right.(float64)
		case GreaterEqual:
			return leftVal <= right.(float64)
		default:
			panic("Invalid comp")
		}
	default:
		panic("Invalid type")
	}

	return false
}

func Eval(data Data, statement string) (resp string, err error) {
	parts := strings.Fields(statement)
	partsLen := len(parts)
	if partsLen < 6 {
		err = &errortypes.ParseError{
			errors.Newf("eval: Statement missing parts"),
		}
		return
	} else if partsLen > 30 {
		err = &errortypes.ParseError{
			errors.Newf("eval: Statement has too many parts"),
		}
		return
	} else if len(statement) > 1024 {
		err = &errortypes.ParseError{
			errors.Newf("eval: Statement too long"),
		}
		return
	}

	if parts[0] != "IF" {
		err = &errortypes.ParseError{
			errors.Newf("eval: Statement part (0) invalid"),
		}
		return
	}

	i := 1
	var expr interface{}
	results := []bool{}
	final := false
	for x := 0; x < 100; x++ {
		if partsLen < i+4 {
			err = &errortypes.ParseError{
				errors.Newf("eval: Incomplete expression (%d)", i),
			}
			return
		}
		index := i
		i += 4

		leftOp, e := parseRef(data, parts[index], index)
		if e != nil {
			err = e
			return
		}
		comp, e := parseRef(data, parts[index+1], index+1)
		if e != nil {
			err = e
			return
		}
		rightOp, e := parseRef(data, parts[index+2], index+2)
		if e != nil {
			err = e
			return
		}
		next, e := parseRef(data, parts[index+3], index+3)
		if e != nil {
			err = e
			return
		}

		switch leftOp.(type) {
		case string:
			break
		case int:
			break
		case float64:
			break
		case bool:
			break
		default:
			err = &errortypes.ParseError{
				errors.Newf("eval: Invalid left operator (%d)", index),
			}
			return
		}

		switch rightOp.(type) {
		case string:
			break
		case int:
			break
		case float64:
			break
		case bool:
			break
		default:
			err = &errortypes.ParseError{
				errors.Newf("eval: Invalid right operator (%d)", index+2),
			}
			return
		}

		switch comp.(type) {
		case Equal:
			break
		case NotEqual:
			break
		case Less:
			break
		case LessEqual:
			break
		case Greater:
			break
		case GreaterEqual:
			break
		default:
			err = &errortypes.ParseError{
				errors.Newf("eval: Invalid right operator (%d)", index+2),
			}
			return
		}

		result := parseComp(leftOp, rightOp, comp)
		results = append(results, result)

		switch next.(type) {
		case And:
			if expr == nil {
				expr = And{}
			} else if _, ok := expr.(And); !ok {
				err = &errortypes.ParseError{
					errors.Newf("eval: Cannot mix OR AND (%d)", index+2),
				}
				return
			}
			break
		case Or:
			if expr == nil {
				expr = Or{}
			} else if _, ok := expr.(Or); !ok {
				err = &errortypes.ParseError{
					errors.Newf("eval: Cannot mix OR AND (%d)", index+2),
				}
				return
			}
			break
		case Then:
			if i+4 < partsLen-1 {
				err = &errortypes.ParseError{
					errors.Newf("eval: Missing result (%d)", index+4),
				}
				return
			}

			if _, ok := expr.(Or); ok {
				for _, result := range results {
					if result {
						final = true
						break
					}
				}
			} else {
				if len(results) == 0 {
					final = false
				} else {
					final = true
					for _, result := range results {
						if !result {
							final = false
							break
						}
					}
				}
			}

			if final {
				respInf, e := parseRef(data, parts[index+4], index+4)
				if e != nil {
					err = e
					return
				}
				if respStr, ok := respInf.(string); ok {
					resp = respStr
				} else {
					err = &errortypes.ParseError{
						errors.Newf("eval: Result must be string (%d)",
							index+4),
					}
				}
			}
			return
		default:
			err = &errortypes.ParseError{
				errors.Newf("eval: Invalid continuation (%d)", index+3),
			}
			return
		}
	}

	err = &errortypes.ParseError{
		errors.Newf("eval: Infinite loop"),
	}
	return
}
