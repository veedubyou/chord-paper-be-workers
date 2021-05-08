// Code generated by counterfeiter. DO NOT EDIT.
package executorfakes

import (
	"chord-paper-be-workers/src/application/executor"
	"sync"
)

type FakeCommand struct {
	CombinedOutputStub        func() ([]byte, error)
	combinedOutputMutex       sync.RWMutex
	combinedOutputArgsForCall []struct {
	}
	combinedOutputReturns struct {
		result1 []byte
		result2 error
	}
	combinedOutputReturnsOnCall map[int]struct {
		result1 []byte
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeCommand) CombinedOutput() ([]byte, error) {
	fake.combinedOutputMutex.Lock()
	ret, specificReturn := fake.combinedOutputReturnsOnCall[len(fake.combinedOutputArgsForCall)]
	fake.combinedOutputArgsForCall = append(fake.combinedOutputArgsForCall, struct {
	}{})
	stub := fake.CombinedOutputStub
	fakeReturns := fake.combinedOutputReturns
	fake.recordInvocation("CombinedOutput", []interface{}{})
	fake.combinedOutputMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeCommand) CombinedOutputCallCount() int {
	fake.combinedOutputMutex.RLock()
	defer fake.combinedOutputMutex.RUnlock()
	return len(fake.combinedOutputArgsForCall)
}

func (fake *FakeCommand) CombinedOutputCalls(stub func() ([]byte, error)) {
	fake.combinedOutputMutex.Lock()
	defer fake.combinedOutputMutex.Unlock()
	fake.CombinedOutputStub = stub
}

func (fake *FakeCommand) CombinedOutputReturns(result1 []byte, result2 error) {
	fake.combinedOutputMutex.Lock()
	defer fake.combinedOutputMutex.Unlock()
	fake.CombinedOutputStub = nil
	fake.combinedOutputReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *FakeCommand) CombinedOutputReturnsOnCall(i int, result1 []byte, result2 error) {
	fake.combinedOutputMutex.Lock()
	defer fake.combinedOutputMutex.Unlock()
	fake.CombinedOutputStub = nil
	if fake.combinedOutputReturnsOnCall == nil {
		fake.combinedOutputReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 error
		})
	}
	fake.combinedOutputReturnsOnCall[i] = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *FakeCommand) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.combinedOutputMutex.RLock()
	defer fake.combinedOutputMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeCommand) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ executor.Command = new(FakeCommand)