package engine

type Predicate interface {
	Apply(Record) bool
}

var (
	_ = Predicate(F{})
	_ = Predicate(OR{})
	_ = Predicate(AND{})
)

func (f F) Apply(r Record) bool {
	field := r.Field(f.ColName)
	if field == nil {
		return false
	}
	switch f.Comparator {
	case EQ:
		return field.Value.Eq(f.Comparando)
	case NEQ:
		return !field.Value.Eq(f.Comparando)
	case GT:
		return field.Value.Gt(f.Comparando)
	case GTE:
		return field.Value.Gt(f.Comparando) || field.Value.Eq(f.Comparando)
	case LT:
		return field.Value.Lt(f.Comparando)
	case LTE:
		return field.Value.Lt(f.Comparando) || field.Value.Eq(f.Comparando)
	case IN:
		return field.Value.In(f.Comparando)
	case NOT_IN:
		return !field.Value.In(f.Comparando)
	case CompFunc:
		if fn, ok := f.Comparando.(func(any) bool); ok {
			return field.Func(fn)
		}
	}
	return false
}

type OR struct {
	l Predicate
	r Predicate
}

func PredicateOR(l, r Predicate) Predicate {
	return OR{l: l, r: r}
}

func (of OR) Apply(r Record) bool {
	if of.l != nil && of.l.Apply(r) {
		return true
	}
	if of.r != nil && of.r.Apply(r) {
		return true
	}
	return false
}

type AND struct {
	l Predicate
	r Predicate
}

func PredicateAND(l, r Predicate) Predicate {
	return AND{l: l, r: r}
}

func (of AND) Apply(r Record) bool {
	if of.l == nil || !of.l.Apply(r) {
		return false
	}
	if of.r == nil || !of.r.Apply(r) {
		return false
	}
	return true
}

type F struct {
	ColName    string
	Comparator Comparator
	Comparando any
}

type Comparator string

const (
	EQ       Comparator = "=="
	NEQ      Comparator = "!="
	GT       Comparator = ">"
	GTE      Comparator = ">="
	LT       Comparator = "<"
	LTE      Comparator = "<="
	IN       Comparator = "in"
	NOT_IN   Comparator = "not in"
	CompFunc Comparator = "func"
)
