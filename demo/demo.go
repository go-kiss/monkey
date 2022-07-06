package demo

func Add[T int | float64](i, j T) T {
	return i + j
}

type S2[T int | float64] struct{ I T }

func (s *S2[T]) Foo() T { return s.I }
