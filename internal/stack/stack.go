package stack

// Stack is a generic stack for any type T
type Stack[T any] struct {
	items []T
}

// Push adds an item on top of the stack
func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

// TryPop removes and returns the top item and true if ok.
// Returns zero value and false if the stack is empty.
func (s *Stack[T]) TryPop() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	last := len(s.items) - 1
	item := s.items[last]
	s.items = s.items[:last]
	return item, true
}

// Pop removes and returns the top item from the stack.
// Panics if the stack is empty. It calls TryPop internally.
func (s *Stack[T]) Pop() T {
	item, ok := s.TryPop()
	if !ok {
		panic("Pop from empty stack")
	}
	return item
}

// TryPeek returns the top item and true if ok.
// Returns zero value and false if the stack is empty.
func (s *Stack[T]) TryPeek() (T, bool) {
	var zero T
	if len(s.items) == 0 {
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

// Peek returns the top item without removing it.
// Panics if the stack is empty. It calls TryPeek internally.
func (s *Stack[T]) Peek() T {
	item, ok := s.TryPeek()
	if !ok {
		panic("Peek from empty stack")
	}
	return item
}

// Size returns the number of items in the stack
func (s *Stack[T]) Size() int {
	return len(s.items)
}

// Empty checks if the stack has no elements
func (s *Stack[T]) Empty() bool {
	return len(s.items) == 0
}
