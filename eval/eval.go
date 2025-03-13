package eval

import (
	"reflect"
	"strconv"
	"strings"
)

type Data map[string]map[string]interface{}

type Parser struct {
	statement string
	parts     []string
	partsLen  int
	data      Data
}

func (p *Parser) parseRef(ref string, pos int) (val interface{}, err error) {
	n := len(ref)
	if n == 0 {
		return
	}

	if ref[0] == '\'' {
		if ref[n-1] != '\'' {
			err = NewEvalError(
				p.statement,
				pos,
				pos,
				p.partsLen,
				"eval: Invalid string {{.ErrIndex}}",
			)
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
	case "FOR":
		val = For{}
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
		err = NewEvalError(
			p.statement,
			pos,
			pos,
			p.partsLen,
			"eval: Invalid reference {{.ErrIndex}}",
		)
		return
	}

	group := p.data[split[0]]
	if group == nil {
		err = NewEvalError(
			p.statement,
			pos,
			pos,
			p.partsLen,
			"eval: Invalid reference group {{.ErrIndex}}",
		)
		return
	}

	groupVal := group[split[1]]
	if groupVal == nil {
		err = NewEvalError(
			p.statement,
			pos,
			pos,
			p.partsLen,
			"eval: Invalid reference group key {{.ErrIndex}}",
		)
		return
	} else {
		val = groupVal
	}

	return
}

func (p *Parser) parseComp(left, right, comp interface{}) bool {
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
			return leftVal > right.(string)
		case GreaterEqual:
			return leftVal >= right.(string)
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
			return leftVal > right.(int)
		case GreaterEqual:
			return leftVal >= right.(int)
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
			return leftVal > right.(float64)
		case GreaterEqual:
			return leftVal >= right.(float64)
		default:
			panic("Invalid comp")
		}
	default:
		panic("Invalid type")
	}

	return false
}

func (p *Parser) Eval() (resp string, threshold int, err error) {
	p.parts = strings.Fields(p.statement)
	p.partsLen = len(p.parts)
	if p.partsLen < 6 {
		err = NewEvalError(
			p.statement,
			0,
			0,
			p.partsLen,
			"eval: Statement under min parts",
		)
		return
	} else if p.partsLen > 30 {
		err = NewEvalError(
			p.statement,
			0,
			0,
			p.partsLen,
			"eval: Statement exceeds max parts",
		)
		return
	} else if len(p.statement) > 1024 {
		err = NewEvalError(
			p.statement,
			0,
			0,
			p.partsLen,
			"eval: Statement exceeds max length",
		)
		return
	}

	if p.parts[0] != "IF" {
		err = NewEvalError(
			p.statement,
			0,
			0,
			p.partsLen,
			"eval: Statement part {{.ErrorIndex}} invalid",
		)
		return
	}

	i := 1
	var expr interface{}
	results := []bool{}
	final := false
	for x := 0; x < 100; x++ {
		if p.partsLen < i+4 {
			err = NewEvalError(
				p.statement,
				i,
				i,
				p.partsLen,
				"eval: Incomplete expression",
			)
			return
		}
		index := i
		i += 4

		leftOp, e := p.parseRef(p.parts[index], index)
		if e != nil {
			err = e
			return
		}
		comp, e := p.parseRef(p.parts[index+1], index+1)
		if e != nil {
			err = e
			return
		}
		rightOp, e := p.parseRef(p.parts[index+2], index+2)
		if e != nil {
			err = e
			return
		}
		next, e := p.parseRef(p.parts[index+3], index+3)
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
			err = NewEvalError(
				p.statement,
				index,
				index,
				p.partsLen,
				"eval: Invalid left operator {{.ErrIndex}}",
			)
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
			err = NewEvalError(
				p.statement,
				index,
				index+2,
				p.partsLen,
				"eval: Invalid right operator {{.ErrIndex}}",
			)
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
			err = NewEvalError(
				p.statement,
				index,
				index+1,
				p.partsLen,
				"eval: Invalid comparison operator {{.ErrIndex}}",
			)
			return
		}

		result := p.parseComp(leftOp, rightOp, comp)
		results = append(results, result)

		if _, ok := next.(For); ok {
			if index+6 != p.partsLen-1 {
				err = NewEvalError(
					p.statement,
					index,
					index+6,
					p.partsLen,
					"eval: Expected %d length", index+6,
				)
				return
			}

			next, e = p.parseRef(p.parts[index+5], index+5)
			if e != nil {
				err = e
				return
			}

			if _, ok := next.(Then); !ok {
				err = NewEvalError(
					p.statement,
					index,
					index+5,
					p.partsLen,
					"eval: Expected THAN at {{.ErrIndex}}",
				)
				return
			}

			forVal, e := p.parseRef(p.parts[index+4], index+4)
			if e != nil {
				err = e
				return
			}

			if forInt, ok := forVal.(int); ok {
				threshold = forInt
			} else {
				err = NewEvalError(
					p.statement,
					index,
					index+4,
					p.partsLen,
					"eval: Expected FOR value to be int",
				)
				return
			}

			index += 2
		}

		switch next.(type) {
		case And:
			if expr == nil {
				expr = And{}
			} else if _, ok := expr.(And); !ok {
				err = NewEvalError(
					p.statement,
					index,
					index+2,
					p.partsLen,
					"eval: Cannot mix OR with AND",
				)
				return
			}
			break
		case Or:
			if expr == nil {
				expr = Or{}
			} else if _, ok := expr.(Or); !ok {
				err = NewEvalError(
					p.statement,
					index,
					index+2,
					p.partsLen,
					"eval: Cannot mix OR with AND",
				)
				return
			}
			break
		case Then:
			if index+4 != p.partsLen-1 {
				err = NewEvalError(
					p.statement,
					index,
					index,
					p.partsLen,
					"eval: Expected %d length", index+4,
				)
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
				respInf, e := p.parseRef(p.parts[index+4], index+4)
				if e != nil {
					err = e
					return
				}
				if respStr, ok := respInf.(string); ok {
					resp = respStr
				} else {
					err = NewEvalError(
						p.statement,
						index,
						index+4,
						p.partsLen,
						"eval: Result must be string",
					)
					return
				}
			}
			return
		default:
			err = NewEvalError(
				p.statement,
				index,
				index+3,
				p.partsLen,
				"eval: Invalid continuation",
			)
			return
		}
	}

	err = NewEvalError(
		p.statement,
		0,
		0,
		p.partsLen,
		"eval: Infinite loop",
	)
	return
}

func Eval(data Data, statement string) (resp string,
	threshold int, err error) {

	parsr := &Parser{
		statement: statement,
		data:      data,
	}

	resp, threshold, err = parsr.Eval()
	if err != nil {
		return
	}

	return
}
