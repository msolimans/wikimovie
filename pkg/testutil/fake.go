package testutil

import (
	"encoding/json"
	"errors"
	"sync"
)

// FakeMethodArgs map of method to its arguments
type FakeMethodArgs map[string]interface{}

// FakeMethodCalls slice of FakeMethodArgs to keep track of method calls
type FakeMethodCalls []FakeMethodArgs

//methods to allow sorting on FakeMethodCalls

func (c FakeMethodCalls) Len() int { return len(c) }

//Less Arbitrary way to sort slice of interface
func (c FakeMethodCalls) Less(i, j int) bool {
	a, _ := json.MarshalIndent(c[i], "", "  ")
	b, _ := json.MarshalIndent(c[j], "", "  ")
	return string(a) < string(b)
}
func (c FakeMethodCalls) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

//FakeMethodCallsByMethodName is an object that contains all FakeMethodCalls made by method name
//and has an added mutex for concurrency purposes
type FakeMethodCallsByMethodName struct {
	callsMap map[string]FakeMethodCalls
	//mutex needed for getting/setting via go-routines
	mutex *sync.RWMutex
}

func NewFakeMethodCallsByMethodName() *FakeMethodCallsByMethodName {
	return &FakeMethodCallsByMethodName{
		callsMap: make(map[string]FakeMethodCalls),
		mutex:    &sync.RWMutex{},
	}
}

func (f *FakeMethodCallsByMethodName) Reset() {
	for k := range f.callsMap {
		delete(f.callsMap, k)
	}
}

func (f *FakeMethodCallsByMethodName) AddCall(methodName string, call FakeMethodArgs) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.callsMap[methodName] = append(f.callsMap[methodName], call)
}

func (f *FakeMethodCallsByMethodName) GetCalls(methodName string) FakeMethodCalls {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.callsMap[methodName]
}

func (f *FakeMethodCallsByMethodName) GetCallsCount(methodName string) int {
	return len(f.GetCalls(methodName))
}

// ResultParameters map of method name to result parameter represented as interface
//with added support for array of interfaces to represent different calls
type ResultParameters map[string]interface{}

type ResultParametersPerCall []interface{}

func (r ResultParameters) GetMockedResultParameter(methodName string, callCount int) (parameter interface{}, isMocked bool) {
	paramMock, paramMocked := r[methodName]
	if paramMocked {
		if paramMocks, ok := paramMock.(ResultParametersPerCall); ok {
			paramMock = paramMocks[len(paramMocks)-1]
			if callCount <= len(paramMocks) {
				paramMock = paramMocks[callCount-1]
			}
		}
	}
	return paramMock, paramMocked
}

// IsMockedResponse is a utility method that can be used in mocked methods to check if the current method
// has mocked errors or return values
func IsMockedResponse(methodName string, callCount int, errors ResultParameters, values ResultParameters) (val interface{}, err error, isMocked bool) {
	errMock, errMocked := errors.GetMockedResultParameter(methodName, callCount)
	err, _ = errMock.(error)
	val, valMocked := values.GetMockedResultParameter(methodName, callCount)
	isMocked = errMocked || valMocked
	return
}

func GenericError() error {
	return errors.New("generic error")
}
