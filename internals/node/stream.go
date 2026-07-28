package node

type Stream[T comparable] struct {
	data []T
}

func (s *Stream[T]) Next() {
	if s.HasData() {
		s.data = s.data[1:]
	}
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

// when using this method, you probably don't want to call .Next()
// because the last token you'll be on is the token AFTER target
func (s *Stream[T]) CollectWhile(target T) []T {
	result := []T{}
	for s.HasData() && s.Current() == target {
		result = append(result, target)
		s.Next()
	}
	return result
}

// when using this method, you probably don't want to call .Next()
// because the last token you'll be on is the token AFTER target
func (s *Stream[T]) CollectUntil(target T) []T {
	result := []T{}
	for s.HasData() && s.Current() != target {
		result = append(result, target)
		s.Next()
	}
	return result
}

func (s *Stream[T]) MatchAhead(targets ...T) bool {
	for i, target := range targets {
		if target != s.PeekOffset(i+1) {
			return false
		}
	}
	return true
}
