// Code generated by mockery v2.39.1. DO NOT EDIT.

package discovery

import (
	mock "github.com/stretchr/testify/mock"
	discovery "k8s.io/client-go/discovery"

	openapi "k8s.io/client-go/openapi"

	openapi_v2 "github.com/google/gnostic-models/openapiv2"

	rest "k8s.io/client-go/rest"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	version "k8s.io/apimachinery/pkg/version"
)

// MockDiscoveryInterface is an autogenerated mock type for the DiscoveryInterface type
type MockDiscoveryInterface struct {
	mock.Mock
}

type MockDiscoveryInterface_Expecter struct {
	mock *mock.Mock
}

func (_m *MockDiscoveryInterface) EXPECT() *MockDiscoveryInterface_Expecter {
	return &MockDiscoveryInterface_Expecter{mock: &_m.Mock}
}

// OpenAPISchema provides a mock function with given fields:
func (_m *MockDiscoveryInterface) OpenAPISchema() (*openapi_v2.Document, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for OpenAPISchema")
	}

	var r0 *openapi_v2.Document
	var r1 error
	if rf, ok := ret.Get(0).(func() (*openapi_v2.Document, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *openapi_v2.Document); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*openapi_v2.Document)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDiscoveryInterface_OpenAPISchema_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OpenAPISchema'
type MockDiscoveryInterface_OpenAPISchema_Call struct {
	*mock.Call
}

// OpenAPISchema is a helper method to define mock.On call
func (_e *MockDiscoveryInterface_Expecter) OpenAPISchema() *MockDiscoveryInterface_OpenAPISchema_Call {
	return &MockDiscoveryInterface_OpenAPISchema_Call{Call: _e.mock.On("OpenAPISchema")}
}

func (_c *MockDiscoveryInterface_OpenAPISchema_Call) Run(run func()) *MockDiscoveryInterface_OpenAPISchema_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockDiscoveryInterface_OpenAPISchema_Call) Return(_a0 *openapi_v2.Document, _a1 error) *MockDiscoveryInterface_OpenAPISchema_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockDiscoveryInterface_OpenAPISchema_Call) RunAndReturn(run func() (*openapi_v2.Document, error)) *MockDiscoveryInterface_OpenAPISchema_Call {
	_c.Call.Return(run)
	return _c
}

// OpenAPIV3 provides a mock function with given fields:
func (_m *MockDiscoveryInterface) OpenAPIV3() openapi.Client {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for OpenAPIV3")
	}

	var r0 openapi.Client
	if rf, ok := ret.Get(0).(func() openapi.Client); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(openapi.Client)
		}
	}

	return r0
}

// MockDiscoveryInterface_OpenAPIV3_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OpenAPIV3'
type MockDiscoveryInterface_OpenAPIV3_Call struct {
	*mock.Call
}

// OpenAPIV3 is a helper method to define mock.On call
func (_e *MockDiscoveryInterface_Expecter) OpenAPIV3() *MockDiscoveryInterface_OpenAPIV3_Call {
	return &MockDiscoveryInterface_OpenAPIV3_Call{Call: _e.mock.On("OpenAPIV3")}
}

func (_c *MockDiscoveryInterface_OpenAPIV3_Call) Run(run func()) *MockDiscoveryInterface_OpenAPIV3_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockDiscoveryInterface_OpenAPIV3_Call) Return(_a0 openapi.Client) *MockDiscoveryInterface_OpenAPIV3_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockDiscoveryInterface_OpenAPIV3_Call) RunAndReturn(run func() openapi.Client) *MockDiscoveryInterface_OpenAPIV3_Call {
	_c.Call.Return(run)
	return _c
}

// RESTClient provides a mock function with given fields:
func (_m *MockDiscoveryInterface) RESTClient() rest.Interface {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for RESTClient")
	}

	var r0 rest.Interface
	if rf, ok := ret.Get(0).(func() rest.Interface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(rest.Interface)
		}
	}

	return r0
}

// MockDiscoveryInterface_RESTClient_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RESTClient'
type MockDiscoveryInterface_RESTClient_Call struct {
	*mock.Call
}

// RESTClient is a helper method to define mock.On call
func (_e *MockDiscoveryInterface_Expecter) RESTClient() *MockDiscoveryInterface_RESTClient_Call {
	return &MockDiscoveryInterface_RESTClient_Call{Call: _e.mock.On("RESTClient")}
}

func (_c *MockDiscoveryInterface_RESTClient_Call) Run(run func()) *MockDiscoveryInterface_RESTClient_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockDiscoveryInterface_RESTClient_Call) Return(_a0 rest.Interface) *MockDiscoveryInterface_RESTClient_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockDiscoveryInterface_RESTClient_Call) RunAndReturn(run func() rest.Interface) *MockDiscoveryInterface_RESTClient_Call {
	_c.Call.Return(run)
	return _c
}

// ServerGroups provides a mock function with given fields:
func (_m *MockDiscoveryInterface) ServerGroups() (*v1.APIGroupList, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ServerGroups")
	}

	var r0 *v1.APIGroupList
	var r1 error
	if rf, ok := ret.Get(0).(func() (*v1.APIGroupList, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *v1.APIGroupList); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.APIGroupList)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDiscoveryInterface_ServerGroups_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ServerGroups'
type MockDiscoveryInterface_ServerGroups_Call struct {
	*mock.Call
}

// ServerGroups is a helper method to define mock.On call
func (_e *MockDiscoveryInterface_Expecter) ServerGroups() *MockDiscoveryInterface_ServerGroups_Call {
	return &MockDiscoveryInterface_ServerGroups_Call{Call: _e.mock.On("ServerGroups")}
}

func (_c *MockDiscoveryInterface_ServerGroups_Call) Run(run func()) *MockDiscoveryInterface_ServerGroups_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockDiscoveryInterface_ServerGroups_Call) Return(_a0 *v1.APIGroupList, _a1 error) *MockDiscoveryInterface_ServerGroups_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockDiscoveryInterface_ServerGroups_Call) RunAndReturn(run func() (*v1.APIGroupList, error)) *MockDiscoveryInterface_ServerGroups_Call {
	_c.Call.Return(run)
	return _c
}

// ServerGroupsAndResources provides a mock function with given fields:
func (_m *MockDiscoveryInterface) ServerGroupsAndResources() ([]*v1.APIGroup, []*v1.APIResourceList, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ServerGroupsAndResources")
	}

	var r0 []*v1.APIGroup
	var r1 []*v1.APIResourceList
	var r2 error
	if rf, ok := ret.Get(0).(func() ([]*v1.APIGroup, []*v1.APIResourceList, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []*v1.APIGroup); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v1.APIGroup)
		}
	}

	if rf, ok := ret.Get(1).(func() []*v1.APIResourceList); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]*v1.APIResourceList)
		}
	}

	if rf, ok := ret.Get(2).(func() error); ok {
		r2 = rf()
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// MockDiscoveryInterface_ServerGroupsAndResources_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ServerGroupsAndResources'
type MockDiscoveryInterface_ServerGroupsAndResources_Call struct {
	*mock.Call
}

// ServerGroupsAndResources is a helper method to define mock.On call
func (_e *MockDiscoveryInterface_Expecter) ServerGroupsAndResources() *MockDiscoveryInterface_ServerGroupsAndResources_Call {
	return &MockDiscoveryInterface_ServerGroupsAndResources_Call{Call: _e.mock.On("ServerGroupsAndResources")}
}

func (_c *MockDiscoveryInterface_ServerGroupsAndResources_Call) Run(run func()) *MockDiscoveryInterface_ServerGroupsAndResources_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockDiscoveryInterface_ServerGroupsAndResources_Call) Return(_a0 []*v1.APIGroup, _a1 []*v1.APIResourceList, _a2 error) *MockDiscoveryInterface_ServerGroupsAndResources_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *MockDiscoveryInterface_ServerGroupsAndResources_Call) RunAndReturn(run func() ([]*v1.APIGroup, []*v1.APIResourceList, error)) *MockDiscoveryInterface_ServerGroupsAndResources_Call {
	_c.Call.Return(run)
	return _c
}

// ServerPreferredNamespacedResources provides a mock function with given fields:
func (_m *MockDiscoveryInterface) ServerPreferredNamespacedResources() ([]*v1.APIResourceList, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ServerPreferredNamespacedResources")
	}

	var r0 []*v1.APIResourceList
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]*v1.APIResourceList, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []*v1.APIResourceList); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v1.APIResourceList)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDiscoveryInterface_ServerPreferredNamespacedResources_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ServerPreferredNamespacedResources'
type MockDiscoveryInterface_ServerPreferredNamespacedResources_Call struct {
	*mock.Call
}

// ServerPreferredNamespacedResources is a helper method to define mock.On call
func (_e *MockDiscoveryInterface_Expecter) ServerPreferredNamespacedResources() *MockDiscoveryInterface_ServerPreferredNamespacedResources_Call {
	return &MockDiscoveryInterface_ServerPreferredNamespacedResources_Call{Call: _e.mock.On("ServerPreferredNamespacedResources")}
}

func (_c *MockDiscoveryInterface_ServerPreferredNamespacedResources_Call) Run(run func()) *MockDiscoveryInterface_ServerPreferredNamespacedResources_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockDiscoveryInterface_ServerPreferredNamespacedResources_Call) Return(_a0 []*v1.APIResourceList, _a1 error) *MockDiscoveryInterface_ServerPreferredNamespacedResources_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockDiscoveryInterface_ServerPreferredNamespacedResources_Call) RunAndReturn(run func() ([]*v1.APIResourceList, error)) *MockDiscoveryInterface_ServerPreferredNamespacedResources_Call {
	_c.Call.Return(run)
	return _c
}

// ServerPreferredResources provides a mock function with given fields:
func (_m *MockDiscoveryInterface) ServerPreferredResources() ([]*v1.APIResourceList, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ServerPreferredResources")
	}

	var r0 []*v1.APIResourceList
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]*v1.APIResourceList, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []*v1.APIResourceList); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*v1.APIResourceList)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDiscoveryInterface_ServerPreferredResources_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ServerPreferredResources'
type MockDiscoveryInterface_ServerPreferredResources_Call struct {
	*mock.Call
}

// ServerPreferredResources is a helper method to define mock.On call
func (_e *MockDiscoveryInterface_Expecter) ServerPreferredResources() *MockDiscoveryInterface_ServerPreferredResources_Call {
	return &MockDiscoveryInterface_ServerPreferredResources_Call{Call: _e.mock.On("ServerPreferredResources")}
}

func (_c *MockDiscoveryInterface_ServerPreferredResources_Call) Run(run func()) *MockDiscoveryInterface_ServerPreferredResources_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockDiscoveryInterface_ServerPreferredResources_Call) Return(_a0 []*v1.APIResourceList, _a1 error) *MockDiscoveryInterface_ServerPreferredResources_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockDiscoveryInterface_ServerPreferredResources_Call) RunAndReturn(run func() ([]*v1.APIResourceList, error)) *MockDiscoveryInterface_ServerPreferredResources_Call {
	_c.Call.Return(run)
	return _c
}

// ServerResourcesForGroupVersion provides a mock function with given fields: groupVersion
func (_m *MockDiscoveryInterface) ServerResourcesForGroupVersion(groupVersion string) (*v1.APIResourceList, error) {
	ret := _m.Called(groupVersion)

	if len(ret) == 0 {
		panic("no return value specified for ServerResourcesForGroupVersion")
	}

	var r0 *v1.APIResourceList
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*v1.APIResourceList, error)); ok {
		return rf(groupVersion)
	}
	if rf, ok := ret.Get(0).(func(string) *v1.APIResourceList); ok {
		r0 = rf(groupVersion)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.APIResourceList)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(groupVersion)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDiscoveryInterface_ServerResourcesForGroupVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ServerResourcesForGroupVersion'
type MockDiscoveryInterface_ServerResourcesForGroupVersion_Call struct {
	*mock.Call
}

// ServerResourcesForGroupVersion is a helper method to define mock.On call
//   - groupVersion string
func (_e *MockDiscoveryInterface_Expecter) ServerResourcesForGroupVersion(groupVersion interface{}) *MockDiscoveryInterface_ServerResourcesForGroupVersion_Call {
	return &MockDiscoveryInterface_ServerResourcesForGroupVersion_Call{Call: _e.mock.On("ServerResourcesForGroupVersion", groupVersion)}
}

func (_c *MockDiscoveryInterface_ServerResourcesForGroupVersion_Call) Run(run func(groupVersion string)) *MockDiscoveryInterface_ServerResourcesForGroupVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockDiscoveryInterface_ServerResourcesForGroupVersion_Call) Return(_a0 *v1.APIResourceList, _a1 error) *MockDiscoveryInterface_ServerResourcesForGroupVersion_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockDiscoveryInterface_ServerResourcesForGroupVersion_Call) RunAndReturn(run func(string) (*v1.APIResourceList, error)) *MockDiscoveryInterface_ServerResourcesForGroupVersion_Call {
	_c.Call.Return(run)
	return _c
}

// ServerVersion provides a mock function with given fields:
func (_m *MockDiscoveryInterface) ServerVersion() (*version.Info, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ServerVersion")
	}

	var r0 *version.Info
	var r1 error
	if rf, ok := ret.Get(0).(func() (*version.Info, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() *version.Info); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*version.Info)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDiscoveryInterface_ServerVersion_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ServerVersion'
type MockDiscoveryInterface_ServerVersion_Call struct {
	*mock.Call
}

// ServerVersion is a helper method to define mock.On call
func (_e *MockDiscoveryInterface_Expecter) ServerVersion() *MockDiscoveryInterface_ServerVersion_Call {
	return &MockDiscoveryInterface_ServerVersion_Call{Call: _e.mock.On("ServerVersion")}
}

func (_c *MockDiscoveryInterface_ServerVersion_Call) Run(run func()) *MockDiscoveryInterface_ServerVersion_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockDiscoveryInterface_ServerVersion_Call) Return(_a0 *version.Info, _a1 error) *MockDiscoveryInterface_ServerVersion_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockDiscoveryInterface_ServerVersion_Call) RunAndReturn(run func() (*version.Info, error)) *MockDiscoveryInterface_ServerVersion_Call {
	_c.Call.Return(run)
	return _c
}

// WithLegacy provides a mock function with given fields:
func (_m *MockDiscoveryInterface) WithLegacy() discovery.DiscoveryInterface {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for WithLegacy")
	}

	var r0 discovery.DiscoveryInterface
	if rf, ok := ret.Get(0).(func() discovery.DiscoveryInterface); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(discovery.DiscoveryInterface)
		}
	}

	return r0
}

// MockDiscoveryInterface_WithLegacy_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WithLegacy'
type MockDiscoveryInterface_WithLegacy_Call struct {
	*mock.Call
}

// WithLegacy is a helper method to define mock.On call
func (_e *MockDiscoveryInterface_Expecter) WithLegacy() *MockDiscoveryInterface_WithLegacy_Call {
	return &MockDiscoveryInterface_WithLegacy_Call{Call: _e.mock.On("WithLegacy")}
}

func (_c *MockDiscoveryInterface_WithLegacy_Call) Run(run func()) *MockDiscoveryInterface_WithLegacy_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockDiscoveryInterface_WithLegacy_Call) Return(_a0 discovery.DiscoveryInterface) *MockDiscoveryInterface_WithLegacy_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockDiscoveryInterface_WithLegacy_Call) RunAndReturn(run func() discovery.DiscoveryInterface) *MockDiscoveryInterface_WithLegacy_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockDiscoveryInterface creates a new instance of MockDiscoveryInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockDiscoveryInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockDiscoveryInterface {
	mock := &MockDiscoveryInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
