package list

func (l *List[T]) RemoveAllAfter(e *Element[T]) {
	if e.list != l || l.len == 0 {
		return
	}

	// first element
	if e.prev == &l.root {
		l.Init()
		return
	}

	node, length := &l.root, 0
	for node.next != e {
		node = node.next
		length++
	}

	e.prev.next = &l.root
	l.root.prev = e.prev
	l.len = length

	if node.next == &l.root {
		return
	}

	// cleanup
	node = node.next
	for node.next != &l.root {
		oldnode := node
		node = node.next

		// prevent memory leaks and other shenanigans
		oldnode.list = nil
		oldnode.next = nil
		oldnode.prev = nil
	}
}
