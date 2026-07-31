package language

type Stream[T comparable] struct {
	data     []T
	previous T
}

func (s *Stream[T]) Next() T {
	if s.HasData() {
		s.previous = s.data[0]
		s.data = s.data[1:]
	}
	return s.Current()
}

func (s *Stream[T]) HasData() bool {
	return len(s.data) > 0
}

func (s *Stream[T]) HasNext() bool {
	return len(s.data) > 1
}

func (s *Stream[T]) Previous() T {
	return s.previous
}

func (s *Stream[T]) Current() T {
	if s.HasData() {
		return s.data[0]
	}

	var zero T
	return zero
}

func (s *Stream[T]) Peek() T {
	return s.PeekOffset(1)
}

func (s *Stream[T]) PeekOffset(offset int) T {
	if offset == -1 {
		return s.Previous()
	}

	if offset < len(s.data) && offset >= 0 {
		return s.data[offset]
	}

	var zero T
	return zero
}
