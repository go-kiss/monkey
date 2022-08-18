package monkey

type Option interface {
	apply(*opt)
}

var OptGlobal = optGlobal{}
var OptGeneric = optGeneric{}

type opt struct {
	global  bool
	generic bool
}

type optGlobal struct{}

func (optGlobal) apply(o *opt) { o.global = true }

type optGeneric struct{}

func (optGeneric) apply(o *opt) { o.generic = true }
