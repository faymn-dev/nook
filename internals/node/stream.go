package node

type Stream[T comparable] struct {
	data []T
}

func (s *Stream[T]) Next() T {
	if s.HasData() {
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
	if offset < len(s.data) {
		return s.data[offset]
	}

	var zero T
	return zero
}
