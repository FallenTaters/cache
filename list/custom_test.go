package list

import (
	"testing"

	"github.com/FallenTaters/cache/assert"
)

func TestRemoveAllAfter(t *testing.T) {
	t.Run(`remove successfully`, func(t *testing.T) {
		l := New[int]()
		l.PushFront(3)
		l.PushFront(2)
		l.PushFront(1)

		l.RemoveAllAfter(l.root.next.next) // 2 and 3 should be removed

		assert.Equal(t, 1, l.len)
		assert.Equal(t, 1, l.root.next.Value)
		assert.Equal(t, &l.root, l.root.next.next, `loop not closed`)
	})

	t.Run(`remove first element`, func(t *testing.T) {
		l := New[int]()
		l.PushFront(3)
		l.PushFront(2)
		l.PushFront(1)

		l.RemoveAllAfter(l.root.next) // all should be removed

		assert.Equal(t, 0, l.len)
		assert.Equal(t, &l.root, l.root.next)
		assert.Equal(t, &l.root, l.root.prev)
	})

	t.Run(`invalid removal`, func(t *testing.T) {
		l := New[int]()
		l.PushFront(3)
		l.PushFront(2)
		l.PushFront(1)

		l.RemoveAllAfter(&Element[int]{}) // none should be removed

		assert.Equal(t, 3, l.len)
	})
}
