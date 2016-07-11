package spruce

import (
	"fmt"
	"github.com/starkandwayne/goutils/ansi"

	"github.com/starkandwayne/goutils/tree"

	. "github.com/geofffranks/spruce/log"
)

// CartesianProductOperator ...
type CartesianProductOperator struct{}

// Setup ...
func (CartesianProductOperator) Setup() error {
	return nil
}

// Phase ...
func (CartesianProductOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies ...
func (CartesianProductOperator) Dependencies(_ *Evaluator, args []*Expr, locs []*tree.Cursor) []*tree.Cursor {
	l := []*tree.Cursor{}

	for _, arg := range args {
		if arg.Type != Reference {
			continue
		}

		for _, other := range locs {
			if other.Under(arg.Reference) {
				l = append(l, other)
			}
		}
	}

	return l
}

// Run ...
func (CartesianProductOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	DEBUG("running (( cartesian-product ... )) operation at $.%s", ev.Here)
	defer DEBUG("done with (( cartesian-product ... )) operation at $%s\n", ev.Here)

	var vals [][]string

	for i, arg := range args {
		v, err := arg.Resolve(ev.Tree)
		if err != nil {
			DEBUG("     [%d]: resolution failed\n    error: %s", i, err)
			return nil, err
		}

		switch v.Type {
		case Literal:
			DEBUG("  arg[%d]: found string literal '%s'", i, v.Literal)
			vals = append(vals, []string{v.Literal.(string)})

		case Reference:
			DEBUG("  arg[%d]: trying to resolve reference $.%s", i, v.Reference)
			s, err := v.Reference.Resolve(ev.Tree)
			if err != nil {
				DEBUG("     [%d]: resolution failed\n    error: %s", i, err)
				return nil, ansi.Errorf("Unable to resolve `@m{%s}`: %s", v.Reference, err)
			}
			switch s.(type) {
			case []interface{}:
				var strs []string

				DEBUG("     [%d]: resolved to a list; verifying", i)
				for j, sub := range s.([]interface{}) {
					if _, ok := sub.([]interface{}); ok {
						DEBUG("       list[%d]: list item is itself a list; error!", j)
						return nil, fmt.Errorf("cartesian-product operator can only operate on lists of scalar values")

					} else if _, ok := sub.(map[interface{}]interface{}); ok {
						DEBUG("       list[%d]: list item is a map; error!", j)
						return nil, fmt.Errorf("cartesian-product operator can only operate on lists of scalar values")

					}
					DEBUG("       list[%d]: list item is a scalar: %v", j, sub)
					strs = append(strs, fmt.Sprintf("%v", sub))
				}
				vals = append(vals, strs)

			case map[interface{}]interface{}:
				DEBUG("     [%d]: resolved to a map; error!", i)
				return nil, fmt.Errorf("cartesian-product operator only accepts arrays and string values")

			default:
				DEBUG("     [%d]: resolved to a scalar; appending", i)
				vals = append(vals, []string{fmt.Sprintf("%v", s)})
			}

		default:
			DEBUG("  arg[%d]: I don't know what to do with '%v'", i, arg)
			return nil, fmt.Errorf("cartesian-product operator only accepts key reference arguments")
		}
		DEBUG("")
	}

	switch len(args) {
	case 0:
		DEBUG("  no arguments supplied to (( cartesian-product ... )) operation.  oops.")
		return nil, ansi.Errorf("no arguments specified to @c{(( cartesian-product ... ))}")

	case 1:
		DEBUG("  called with only one argument; returning value as-is")
		return &Response{
			Type:  Replace,
			Value: vals[0],
		}, nil

	default:
		DEBUG("  called with more than one arguments; combining into a single list of strings")

		lst := vals[0]
		for _, l := range vals[1:] {
			lst = cartesian(lst, l)
		}

		return &Response{
			Type:  Replace,
			Value: lst,
		}, nil
	}
}

func cartesian(a, b []string) []string {
	if len(a) == 0 {
		return a
	}
	if len(b) == 0 {
		return b
	}

	l := make([]string, len(a)*len(b))
	n := 0
	for _, x := range a {
		for _, y := range b {
			l[n] = fmt.Sprintf("%s%s", x, y)
			n++
		}
	}

	return l
}

func init() {
	RegisterOp("cartesian-product", CartesianProductOperator{})
}
