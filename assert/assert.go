/*
assert should be a standalong go module
*/
package assert

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func Equal[T comparable](t *testing.T, expected, actual T, messages ...string) {
	t.Helper()
	if expected == actual {
		return
	}

	t.Errorf(`%sexpected %v, got %v`, formatMessages(messages), expected, actual)
}

func NotEqual[T comparable](t *testing.T, expected, actual T, messages ...string) {
	t.Helper()
	if expected != actual {
		return
	}

	t.Errorf(`%sexpected %v not to equal %v`, formatMessages(messages), expected, actual)
}

func True(t *testing.T, v bool, messages ...string) {
	t.Helper()
	Equal(t, true, v, messages...)
}

func False(t *testing.T, v bool, messages ...string) {
	t.Helper()
	Equal(t, false, v, messages...)
}

func NoError(t *testing.T, err error, messages ...string) {
	t.Helper()
	if err == nil {
		return
	}

	t.Errorf(`%serror should be nil: %v`, formatMessages(messages), err)
}

func ErrorIs(t *testing.T, target error, err error, messages ...string) {
	t.Helper()

	if errors.Is(err, target) {
		return
	}

	if err == nil {
		t.Errorf(`%serror should have error %q in its chain but is nil`, formatMessages(messages), target)
		return
	}

	t.Errorf(`%serror %q should have error %q in its chain`, formatMessages(messages), err, target)
}

func ErrorAs(t *testing.T, target any, err error, messages ...string) {
	t.Helper()
	if errors.As(err, target) {
		return
	}

	if err == nil {
		t.Errorf(`%serror should have error of type %T in its chain but is nil`, formatMessages(messages), target)
		return
	}

	t.Errorf(`%serror %q should have an error of type %T in its chain`, formatMessages(messages), err, target)
}

func Error(t *testing.T, err error, messages ...string) {
	t.Helper()
	if err != nil {
		return
	}

	t.Errorf(`%serror should be non-nil: %v`, formatMessages(messages), err)
}

func Nil(t *testing.T, v any, messages ...string) {
	t.Helper()
	if v == nil {
		return
	}

	t.Errorf(`%sshould be nil: %v`, formatMessages(messages), v)
}

func NotNil(t *testing.T, v any, messages ...string) {
	t.Helper()
	if v != nil {
		return
	}

	t.Errorf(`%sshould be non-nil: %v`, formatMessages(messages), v)
}

func formatMessages(messages []string) string {
	msg := strings.Join(messages, `, `)
	if len(msg) > 0 {
		return fmt.Sprintf(`(%s): `, msg)
	}

	return ``
}
