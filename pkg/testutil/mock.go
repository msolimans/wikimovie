package testutil

import (
	"testing"
	"time"
)

type Mock struct {
	FakeMethodCallsByMethodName *FakeMethodCallsByMethodName
	ReturnErrors                ResultParameters
	ReturnValues                ResultParameters
	now                         *time.Time
}

func NewMock(t *testing.T) *Mock {
	client := &Mock{
		FakeMethodCallsByMethodName: NewFakeMethodCallsByMethodName(),
		ReturnErrors:                make(ResultParameters),
		ReturnValues:                make(ResultParameters),
	}
	t.Cleanup(func() {
		client.Reset()
	})
	return client
}

// Reset resets the ReturnsErrors and MethodArgs to be empty
func (m *Mock) Reset() {
	for k := range m.ReturnErrors {
		delete(m.ReturnErrors, k)
	}
	for k := range m.ReturnValues {
		delete(m.ReturnValues, k)
	}
	m.FakeMethodCallsByMethodName.Reset()
	m.now = nil
}

func (m *Mock) FreezeNow(t time.Time) {
	m.now = &t
}

func (m *Mock) GetNow() time.Time {
	if m.now != nil {
		return *m.now
	}
	return time.Now().UTC()
}
