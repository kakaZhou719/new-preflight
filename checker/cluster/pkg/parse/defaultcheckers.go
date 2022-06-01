package parse

// Checker validates the state of the system to ensure trident will be
// successful as often as possible.
type Checker interface {
	Name() string
	Script() string
	Params() string
}

// DefaultCheckers register your Checkers here.
var DefaultCheckers = []Checker{
	parseinstance{},
}

var ParseInstance = parseinstance{}
