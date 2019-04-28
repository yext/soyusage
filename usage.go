package soyusage

type Usage int

const (
	UsageFull Usage = iota
	UsageUnknown
	UsageReference
)

type UsageByTemplate map[string][]Usage

type Params map[string]*Param

type Param struct {
	Children Params
	Usage    UsageByTemplate
}

func newParam() *Param {
	return &Param{
		Children: make(Params),
		Usage:    make(UsageByTemplate),
	}
}
