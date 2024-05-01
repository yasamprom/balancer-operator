package model

// Triggers is a struct for triggers rules.
//
//	Operator can trigger on spec.label or spec.name
type Triggers struct {
	Labels []string
	Names  []string
}
