// Code generated by github.com/efritz/go-mockgen 0.1.0; DO NOT EDIT.

package mocks

import (
	"context"
	gitserver "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"io"
	"sync"
)

// MockClient is a mock implementation of the Client interface (from the
// package
// github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver)
// used for unit testing.
type MockClient struct {
	// ArchiveFunc is an instance of a mock function object controlling the
	// behavior of the method Archive.
	ArchiveFunc *ClientArchiveFunc
	// CommitGraphFunc is an instance of a mock function object controlling
	// the behavior of the method CommitGraph.
	CommitGraphFunc *ClientCommitGraphFunc
	// DirectoryChildrenFunc is an instance of a mock function object
	// controlling the behavior of the method DirectoryChildren.
	DirectoryChildrenFunc *ClientDirectoryChildrenFunc
	// FileExistsFunc is an instance of a mock function object controlling
	// the behavior of the method FileExists.
	FileExistsFunc *ClientFileExistsFunc
	// HeadFunc is an instance of a mock function object controlling the
	// behavior of the method Head.
	HeadFunc *ClientHeadFunc
	// TagsFunc is an instance of a mock function object controlling the
	// behavior of the method Tags.
	TagsFunc *ClientTagsFunc
}

// NewMockClient creates a new mock of the Client interface. All methods
// return zero values for all results, unless overwritten.
func NewMockClient() *MockClient {
	return &MockClient{
		ArchiveFunc: &ClientArchiveFunc{
			defaultHook: func(context.Context, store.Store, int, string) (io.Reader, error) {
				return nil, nil
			},
		},
		CommitGraphFunc: &ClientCommitGraphFunc{
			defaultHook: func(context.Context, store.Store, int) (map[string][]string, error) {
				return nil, nil
			},
		},
		DirectoryChildrenFunc: &ClientDirectoryChildrenFunc{
			defaultHook: func(context.Context, store.Store, int, string, []string) (map[string][]string, error) {
				return nil, nil
			},
		},
		FileExistsFunc: &ClientFileExistsFunc{
			defaultHook: func(context.Context, store.Store, int, string, string) (bool, error) {
				return false, nil
			},
		},
		HeadFunc: &ClientHeadFunc{
			defaultHook: func(context.Context, store.Store, int) (string, error) {
				return "", nil
			},
		},
		TagsFunc: &ClientTagsFunc{
			defaultHook: func(context.Context, store.Store, int, string) (string, bool, error) {
				return "", false, nil
			},
		},
	}
}

// NewMockClientFrom creates a new mock of the MockClient interface. All
// methods delegate to the given implementation, unless overwritten.
func NewMockClientFrom(i gitserver.Client) *MockClient {
	return &MockClient{
		ArchiveFunc: &ClientArchiveFunc{
			defaultHook: i.Archive,
		},
		CommitGraphFunc: &ClientCommitGraphFunc{
			defaultHook: i.CommitGraph,
		},
		DirectoryChildrenFunc: &ClientDirectoryChildrenFunc{
			defaultHook: i.DirectoryChildren,
		},
		FileExistsFunc: &ClientFileExistsFunc{
			defaultHook: i.FileExists,
		},
		HeadFunc: &ClientHeadFunc{
			defaultHook: i.Head,
		},
		TagsFunc: &ClientTagsFunc{
			defaultHook: i.Tags,
		},
	}
}

// ClientArchiveFunc describes the behavior when the Archive method of the
// parent MockClient instance is invoked.
type ClientArchiveFunc struct {
	defaultHook func(context.Context, store.Store, int, string) (io.Reader, error)
	hooks       []func(context.Context, store.Store, int, string) (io.Reader, error)
	history     []ClientArchiveFuncCall
	mutex       sync.Mutex
}

// Archive delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockClient) Archive(v0 context.Context, v1 store.Store, v2 int, v3 string) (io.Reader, error) {
	r0, r1 := m.ArchiveFunc.nextHook()(v0, v1, v2, v3)
	m.ArchiveFunc.appendCall(ClientArchiveFuncCall{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the Archive method of
// the parent MockClient instance is invoked and the hook queue is empty.
func (f *ClientArchiveFunc) SetDefaultHook(hook func(context.Context, store.Store, int, string) (io.Reader, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Archive method of the parent MockClient instance inovkes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *ClientArchiveFunc) PushHook(hook func(context.Context, store.Store, int, string) (io.Reader, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *ClientArchiveFunc) SetDefaultReturn(r0 io.Reader, r1 error) {
	f.SetDefaultHook(func(context.Context, store.Store, int, string) (io.Reader, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *ClientArchiveFunc) PushReturn(r0 io.Reader, r1 error) {
	f.PushHook(func(context.Context, store.Store, int, string) (io.Reader, error) {
		return r0, r1
	})
}

func (f *ClientArchiveFunc) nextHook() func(context.Context, store.Store, int, string) (io.Reader, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientArchiveFunc) appendCall(r0 ClientArchiveFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of ClientArchiveFuncCall objects describing
// the invocations of this function.
func (f *ClientArchiveFunc) History() []ClientArchiveFuncCall {
	f.mutex.Lock()
	history := make([]ClientArchiveFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientArchiveFuncCall is an object that describes an invocation of method
// Archive on an instance of MockClient.
type ClientArchiveFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 store.Store
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 int
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 io.Reader
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c ClientArchiveFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c ClientArchiveFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// ClientCommitGraphFunc describes the behavior when the CommitGraph method
// of the parent MockClient instance is invoked.
type ClientCommitGraphFunc struct {
	defaultHook func(context.Context, store.Store, int) (map[string][]string, error)
	hooks       []func(context.Context, store.Store, int) (map[string][]string, error)
	history     []ClientCommitGraphFuncCall
	mutex       sync.Mutex
}

// CommitGraph delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockClient) CommitGraph(v0 context.Context, v1 store.Store, v2 int) (map[string][]string, error) {
	r0, r1 := m.CommitGraphFunc.nextHook()(v0, v1, v2)
	m.CommitGraphFunc.appendCall(ClientCommitGraphFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the CommitGraph method
// of the parent MockClient instance is invoked and the hook queue is empty.
func (f *ClientCommitGraphFunc) SetDefaultHook(hook func(context.Context, store.Store, int) (map[string][]string, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// CommitGraph method of the parent MockClient instance inovkes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *ClientCommitGraphFunc) PushHook(hook func(context.Context, store.Store, int) (map[string][]string, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *ClientCommitGraphFunc) SetDefaultReturn(r0 map[string][]string, r1 error) {
	f.SetDefaultHook(func(context.Context, store.Store, int) (map[string][]string, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *ClientCommitGraphFunc) PushReturn(r0 map[string][]string, r1 error) {
	f.PushHook(func(context.Context, store.Store, int) (map[string][]string, error) {
		return r0, r1
	})
}

func (f *ClientCommitGraphFunc) nextHook() func(context.Context, store.Store, int) (map[string][]string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientCommitGraphFunc) appendCall(r0 ClientCommitGraphFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of ClientCommitGraphFuncCall objects
// describing the invocations of this function.
func (f *ClientCommitGraphFunc) History() []ClientCommitGraphFuncCall {
	f.mutex.Lock()
	history := make([]ClientCommitGraphFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientCommitGraphFuncCall is an object that describes an invocation of
// method CommitGraph on an instance of MockClient.
type ClientCommitGraphFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 store.Store
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 map[string][]string
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c ClientCommitGraphFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c ClientCommitGraphFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// ClientDirectoryChildrenFunc describes the behavior when the
// DirectoryChildren method of the parent MockClient instance is invoked.
type ClientDirectoryChildrenFunc struct {
	defaultHook func(context.Context, store.Store, int, string, []string) (map[string][]string, error)
	hooks       []func(context.Context, store.Store, int, string, []string) (map[string][]string, error)
	history     []ClientDirectoryChildrenFuncCall
	mutex       sync.Mutex
}

// DirectoryChildren delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockClient) DirectoryChildren(v0 context.Context, v1 store.Store, v2 int, v3 string, v4 []string) (map[string][]string, error) {
	r0, r1 := m.DirectoryChildrenFunc.nextHook()(v0, v1, v2, v3, v4)
	m.DirectoryChildrenFunc.appendCall(ClientDirectoryChildrenFuncCall{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the DirectoryChildren
// method of the parent MockClient instance is invoked and the hook queue is
// empty.
func (f *ClientDirectoryChildrenFunc) SetDefaultHook(hook func(context.Context, store.Store, int, string, []string) (map[string][]string, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// DirectoryChildren method of the parent MockClient instance inovkes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *ClientDirectoryChildrenFunc) PushHook(hook func(context.Context, store.Store, int, string, []string) (map[string][]string, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *ClientDirectoryChildrenFunc) SetDefaultReturn(r0 map[string][]string, r1 error) {
	f.SetDefaultHook(func(context.Context, store.Store, int, string, []string) (map[string][]string, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *ClientDirectoryChildrenFunc) PushReturn(r0 map[string][]string, r1 error) {
	f.PushHook(func(context.Context, store.Store, int, string, []string) (map[string][]string, error) {
		return r0, r1
	})
}

func (f *ClientDirectoryChildrenFunc) nextHook() func(context.Context, store.Store, int, string, []string) (map[string][]string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientDirectoryChildrenFunc) appendCall(r0 ClientDirectoryChildrenFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of ClientDirectoryChildrenFuncCall objects
// describing the invocations of this function.
func (f *ClientDirectoryChildrenFunc) History() []ClientDirectoryChildrenFuncCall {
	f.mutex.Lock()
	history := make([]ClientDirectoryChildrenFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientDirectoryChildrenFuncCall is an object that describes an invocation
// of method DirectoryChildren on an instance of MockClient.
type ClientDirectoryChildrenFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 store.Store
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 int
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 string
	// Arg4 is the value of the 5th argument passed to this method
	// invocation.
	Arg4 []string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 map[string][]string
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c ClientDirectoryChildrenFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c ClientDirectoryChildrenFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// ClientFileExistsFunc describes the behavior when the FileExists method of
// the parent MockClient instance is invoked.
type ClientFileExistsFunc struct {
	defaultHook func(context.Context, store.Store, int, string, string) (bool, error)
	hooks       []func(context.Context, store.Store, int, string, string) (bool, error)
	history     []ClientFileExistsFuncCall
	mutex       sync.Mutex
}

// FileExists delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockClient) FileExists(v0 context.Context, v1 store.Store, v2 int, v3 string, v4 string) (bool, error) {
	r0, r1 := m.FileExistsFunc.nextHook()(v0, v1, v2, v3, v4)
	m.FileExistsFunc.appendCall(ClientFileExistsFuncCall{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the FileExists method of
// the parent MockClient instance is invoked and the hook queue is empty.
func (f *ClientFileExistsFunc) SetDefaultHook(hook func(context.Context, store.Store, int, string, string) (bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// FileExists method of the parent MockClient instance inovkes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *ClientFileExistsFunc) PushHook(hook func(context.Context, store.Store, int, string, string) (bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *ClientFileExistsFunc) SetDefaultReturn(r0 bool, r1 error) {
	f.SetDefaultHook(func(context.Context, store.Store, int, string, string) (bool, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *ClientFileExistsFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, store.Store, int, string, string) (bool, error) {
		return r0, r1
	})
}

func (f *ClientFileExistsFunc) nextHook() func(context.Context, store.Store, int, string, string) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientFileExistsFunc) appendCall(r0 ClientFileExistsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of ClientFileExistsFuncCall objects describing
// the invocations of this function.
func (f *ClientFileExistsFunc) History() []ClientFileExistsFuncCall {
	f.mutex.Lock()
	history := make([]ClientFileExistsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientFileExistsFuncCall is an object that describes an invocation of
// method FileExists on an instance of MockClient.
type ClientFileExistsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 store.Store
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 int
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 string
	// Arg4 is the value of the 5th argument passed to this method
	// invocation.
	Arg4 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 bool
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c ClientFileExistsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c ClientFileExistsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// ClientHeadFunc describes the behavior when the Head method of the parent
// MockClient instance is invoked.
type ClientHeadFunc struct {
	defaultHook func(context.Context, store.Store, int) (string, error)
	hooks       []func(context.Context, store.Store, int) (string, error)
	history     []ClientHeadFuncCall
	mutex       sync.Mutex
}

// Head delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockClient) Head(v0 context.Context, v1 store.Store, v2 int) (string, error) {
	r0, r1 := m.HeadFunc.nextHook()(v0, v1, v2)
	m.HeadFunc.appendCall(ClientHeadFuncCall{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the Head method of the
// parent MockClient instance is invoked and the hook queue is empty.
func (f *ClientHeadFunc) SetDefaultHook(hook func(context.Context, store.Store, int) (string, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Head method of the parent MockClient instance inovkes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *ClientHeadFunc) PushHook(hook func(context.Context, store.Store, int) (string, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *ClientHeadFunc) SetDefaultReturn(r0 string, r1 error) {
	f.SetDefaultHook(func(context.Context, store.Store, int) (string, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *ClientHeadFunc) PushReturn(r0 string, r1 error) {
	f.PushHook(func(context.Context, store.Store, int) (string, error) {
		return r0, r1
	})
}

func (f *ClientHeadFunc) nextHook() func(context.Context, store.Store, int) (string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientHeadFunc) appendCall(r0 ClientHeadFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of ClientHeadFuncCall objects describing the
// invocations of this function.
func (f *ClientHeadFunc) History() []ClientHeadFuncCall {
	f.mutex.Lock()
	history := make([]ClientHeadFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientHeadFuncCall is an object that describes an invocation of method
// Head on an instance of MockClient.
type ClientHeadFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 store.Store
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 string
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c ClientHeadFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c ClientHeadFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// ClientTagsFunc describes the behavior when the Tags method of the parent
// MockClient instance is invoked.
type ClientTagsFunc struct {
	defaultHook func(context.Context, store.Store, int, string) (string, bool, error)
	hooks       []func(context.Context, store.Store, int, string) (string, bool, error)
	history     []ClientTagsFuncCall
	mutex       sync.Mutex
}

// Tags delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockClient) Tags(v0 context.Context, v1 store.Store, v2 int, v3 string) (string, bool, error) {
	r0, r1, r2 := m.TagsFunc.nextHook()(v0, v1, v2, v3)
	m.TagsFunc.appendCall(ClientTagsFuncCall{v0, v1, v2, v3, r0, r1, r2})
	return r0, r1, r2
}

// SetDefaultHook sets function that is called when the Tags method of the
// parent MockClient instance is invoked and the hook queue is empty.
func (f *ClientTagsFunc) SetDefaultHook(hook func(context.Context, store.Store, int, string) (string, bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Tags method of the parent MockClient instance inovkes the hook at the
// front of the queue and discards it. After the queue is empty, the default
// hook function is invoked for any future action.
func (f *ClientTagsFunc) PushHook(hook func(context.Context, store.Store, int, string) (string, bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *ClientTagsFunc) SetDefaultReturn(r0 string, r1 bool, r2 error) {
	f.SetDefaultHook(func(context.Context, store.Store, int, string) (string, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *ClientTagsFunc) PushReturn(r0 string, r1 bool, r2 error) {
	f.PushHook(func(context.Context, store.Store, int, string) (string, bool, error) {
		return r0, r1, r2
	})
}

func (f *ClientTagsFunc) nextHook() func(context.Context, store.Store, int, string) (string, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientTagsFunc) appendCall(r0 ClientTagsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of ClientTagsFuncCall objects describing the
// invocations of this function.
func (f *ClientTagsFunc) History() []ClientTagsFuncCall {
	f.mutex.Lock()
	history := make([]ClientTagsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientTagsFuncCall is an object that describes an invocation of method
// Tags on an instance of MockClient.
type ClientTagsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 store.Store
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 int
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 string
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 bool
	// Result2 is the value of the 3rd result returned from this method
	// invocation.
	Result2 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c ClientTagsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c ClientTagsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1, c.Result2}
}
