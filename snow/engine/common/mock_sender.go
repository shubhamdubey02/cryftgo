// Code generated by MockGen. DO NOT EDIT.
// Source: snow/engine/common/sender.go
//
// Generated by this command:
//
//	mockgen -source=snow/engine/common/sender.go -destination=snow/engine/common/mock_sender.go -package=common -exclude_interfaces=StateSummarySender,AcceptedStateSummarySender,FrontierSender,AcceptedSender,FetchSender,AppSender,QuerySender,CrossChainAppSender,NetworkAppSender,Gossiper
//

// Package common is a generated GoMock package.
package common

import (
	context "context"
	reflect "reflect"

	ids "github.com/shubhamdubey02/cryftgo/ids"
	set "github.com/shubhamdubey02/cryftgo/utils/set"
	gomock "go.uber.org/mock/gomock"
)

// MockSender is a mock of Sender interface.
type MockSender struct {
	ctrl     *gomock.Controller
	recorder *MockSenderMockRecorder
}

// MockSenderMockRecorder is the mock recorder for MockSender.
type MockSenderMockRecorder struct {
	mock *MockSender
}

// NewMockSender creates a new mock instance.
func NewMockSender(ctrl *gomock.Controller) *MockSender {
	mock := &MockSender{ctrl: ctrl}
	mock.recorder = &MockSenderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSender) EXPECT() *MockSenderMockRecorder {
	return m.recorder
}

// SendAccepted mocks base method.
func (m *MockSender) SendAccepted(ctx context.Context, nodeID ids.NodeID, requestID uint32, containerIDs []ids.ID) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendAccepted", ctx, nodeID, requestID, containerIDs)
}

// SendAccepted indicates an expected call of SendAccepted.
func (mr *MockSenderMockRecorder) SendAccepted(ctx, nodeID, requestID, containerIDs any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendAccepted", reflect.TypeOf((*MockSender)(nil).SendAccepted), ctx, nodeID, requestID, containerIDs)
}

// SendAcceptedFrontier mocks base method.
func (m *MockSender) SendAcceptedFrontier(ctx context.Context, nodeID ids.NodeID, requestID uint32, containerID ids.ID) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendAcceptedFrontier", ctx, nodeID, requestID, containerID)
}

// SendAcceptedFrontier indicates an expected call of SendAcceptedFrontier.
func (mr *MockSenderMockRecorder) SendAcceptedFrontier(ctx, nodeID, requestID, containerID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendAcceptedFrontier", reflect.TypeOf((*MockSender)(nil).SendAcceptedFrontier), ctx, nodeID, requestID, containerID)
}

// SendAcceptedStateSummary mocks base method.
func (m *MockSender) SendAcceptedStateSummary(ctx context.Context, nodeID ids.NodeID, requestID uint32, summaryIDs []ids.ID) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendAcceptedStateSummary", ctx, nodeID, requestID, summaryIDs)
}

// SendAcceptedStateSummary indicates an expected call of SendAcceptedStateSummary.
func (mr *MockSenderMockRecorder) SendAcceptedStateSummary(ctx, nodeID, requestID, summaryIDs any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendAcceptedStateSummary", reflect.TypeOf((*MockSender)(nil).SendAcceptedStateSummary), ctx, nodeID, requestID, summaryIDs)
}

// SendAncestors mocks base method.
func (m *MockSender) SendAncestors(ctx context.Context, nodeID ids.NodeID, requestID uint32, containers [][]byte) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendAncestors", ctx, nodeID, requestID, containers)
}

// SendAncestors indicates an expected call of SendAncestors.
func (mr *MockSenderMockRecorder) SendAncestors(ctx, nodeID, requestID, containers any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendAncestors", reflect.TypeOf((*MockSender)(nil).SendAncestors), ctx, nodeID, requestID, containers)
}

// SendAppError mocks base method.
func (m *MockSender) SendAppError(ctx context.Context, nodeID ids.NodeID, requestID uint32, errorCode int32, errorMessage string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendAppError", ctx, nodeID, requestID, errorCode, errorMessage)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendAppError indicates an expected call of SendAppError.
func (mr *MockSenderMockRecorder) SendAppError(ctx, nodeID, requestID, errorCode, errorMessage any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendAppError", reflect.TypeOf((*MockSender)(nil).SendAppError), ctx, nodeID, requestID, errorCode, errorMessage)
}

// SendAppGossip mocks base method.
func (m *MockSender) SendAppGossip(ctx context.Context, config SendConfig, appGossipBytes []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendAppGossip", ctx, config, appGossipBytes)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendAppGossip indicates an expected call of SendAppGossip.
func (mr *MockSenderMockRecorder) SendAppGossip(ctx, config, appGossipBytes any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendAppGossip", reflect.TypeOf((*MockSender)(nil).SendAppGossip), ctx, config, appGossipBytes)
}

// SendAppRequest mocks base method.
func (m *MockSender) SendAppRequest(ctx context.Context, nodeIDs set.Set[ids.NodeID], requestID uint32, appRequestBytes []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendAppRequest", ctx, nodeIDs, requestID, appRequestBytes)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendAppRequest indicates an expected call of SendAppRequest.
func (mr *MockSenderMockRecorder) SendAppRequest(ctx, nodeIDs, requestID, appRequestBytes any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendAppRequest", reflect.TypeOf((*MockSender)(nil).SendAppRequest), ctx, nodeIDs, requestID, appRequestBytes)
}

// SendAppResponse mocks base method.
func (m *MockSender) SendAppResponse(ctx context.Context, nodeID ids.NodeID, requestID uint32, appResponseBytes []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendAppResponse", ctx, nodeID, requestID, appResponseBytes)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendAppResponse indicates an expected call of SendAppResponse.
func (mr *MockSenderMockRecorder) SendAppResponse(ctx, nodeID, requestID, appResponseBytes any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendAppResponse", reflect.TypeOf((*MockSender)(nil).SendAppResponse), ctx, nodeID, requestID, appResponseBytes)
}

// SendChits mocks base method.
func (m *MockSender) SendChits(ctx context.Context, nodeID ids.NodeID, requestID uint32, preferredID, preferredIDAtHeight, acceptedID ids.ID) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendChits", ctx, nodeID, requestID, preferredID, preferredIDAtHeight, acceptedID)
}

// SendChits indicates an expected call of SendChits.
func (mr *MockSenderMockRecorder) SendChits(ctx, nodeID, requestID, preferredID, preferredIDAtHeight, acceptedID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendChits", reflect.TypeOf((*MockSender)(nil).SendChits), ctx, nodeID, requestID, preferredID, preferredIDAtHeight, acceptedID)
}

// SendCrossChainAppError mocks base method.
func (m *MockSender) SendCrossChainAppError(ctx context.Context, chainID ids.ID, requestID uint32, errorCode int32, errorMessage string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendCrossChainAppError", ctx, chainID, requestID, errorCode, errorMessage)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendCrossChainAppError indicates an expected call of SendCrossChainAppError.
func (mr *MockSenderMockRecorder) SendCrossChainAppError(ctx, chainID, requestID, errorCode, errorMessage any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendCrossChainAppError", reflect.TypeOf((*MockSender)(nil).SendCrossChainAppError), ctx, chainID, requestID, errorCode, errorMessage)
}

// SendCrossChainAppRequest mocks base method.
func (m *MockSender) SendCrossChainAppRequest(ctx context.Context, chainID ids.ID, requestID uint32, appRequestBytes []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendCrossChainAppRequest", ctx, chainID, requestID, appRequestBytes)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendCrossChainAppRequest indicates an expected call of SendCrossChainAppRequest.
func (mr *MockSenderMockRecorder) SendCrossChainAppRequest(ctx, chainID, requestID, appRequestBytes any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendCrossChainAppRequest", reflect.TypeOf((*MockSender)(nil).SendCrossChainAppRequest), ctx, chainID, requestID, appRequestBytes)
}

// SendCrossChainAppResponse mocks base method.
func (m *MockSender) SendCrossChainAppResponse(ctx context.Context, chainID ids.ID, requestID uint32, appResponseBytes []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendCrossChainAppResponse", ctx, chainID, requestID, appResponseBytes)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendCrossChainAppResponse indicates an expected call of SendCrossChainAppResponse.
func (mr *MockSenderMockRecorder) SendCrossChainAppResponse(ctx, chainID, requestID, appResponseBytes any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendCrossChainAppResponse", reflect.TypeOf((*MockSender)(nil).SendCrossChainAppResponse), ctx, chainID, requestID, appResponseBytes)
}

// SendGet mocks base method.
func (m *MockSender) SendGet(ctx context.Context, nodeID ids.NodeID, requestID uint32, containerID ids.ID) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendGet", ctx, nodeID, requestID, containerID)
}

// SendGet indicates an expected call of SendGet.
func (mr *MockSenderMockRecorder) SendGet(ctx, nodeID, requestID, containerID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendGet", reflect.TypeOf((*MockSender)(nil).SendGet), ctx, nodeID, requestID, containerID)
}

// SendGetAccepted mocks base method.
func (m *MockSender) SendGetAccepted(ctx context.Context, nodeIDs set.Set[ids.NodeID], requestID uint32, containerIDs []ids.ID) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendGetAccepted", ctx, nodeIDs, requestID, containerIDs)
}

// SendGetAccepted indicates an expected call of SendGetAccepted.
func (mr *MockSenderMockRecorder) SendGetAccepted(ctx, nodeIDs, requestID, containerIDs any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendGetAccepted", reflect.TypeOf((*MockSender)(nil).SendGetAccepted), ctx, nodeIDs, requestID, containerIDs)
}

// SendGetAcceptedFrontier mocks base method.
func (m *MockSender) SendGetAcceptedFrontier(ctx context.Context, nodeIDs set.Set[ids.NodeID], requestID uint32) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendGetAcceptedFrontier", ctx, nodeIDs, requestID)
}

// SendGetAcceptedFrontier indicates an expected call of SendGetAcceptedFrontier.
func (mr *MockSenderMockRecorder) SendGetAcceptedFrontier(ctx, nodeIDs, requestID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendGetAcceptedFrontier", reflect.TypeOf((*MockSender)(nil).SendGetAcceptedFrontier), ctx, nodeIDs, requestID)
}

// SendGetAcceptedStateSummary mocks base method.
func (m *MockSender) SendGetAcceptedStateSummary(ctx context.Context, nodeIDs set.Set[ids.NodeID], requestID uint32, heights []uint64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendGetAcceptedStateSummary", ctx, nodeIDs, requestID, heights)
}

// SendGetAcceptedStateSummary indicates an expected call of SendGetAcceptedStateSummary.
func (mr *MockSenderMockRecorder) SendGetAcceptedStateSummary(ctx, nodeIDs, requestID, heights any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendGetAcceptedStateSummary", reflect.TypeOf((*MockSender)(nil).SendGetAcceptedStateSummary), ctx, nodeIDs, requestID, heights)
}

// SendGetAncestors mocks base method.
func (m *MockSender) SendGetAncestors(ctx context.Context, nodeID ids.NodeID, requestID uint32, containerID ids.ID) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendGetAncestors", ctx, nodeID, requestID, containerID)
}

// SendGetAncestors indicates an expected call of SendGetAncestors.
func (mr *MockSenderMockRecorder) SendGetAncestors(ctx, nodeID, requestID, containerID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendGetAncestors", reflect.TypeOf((*MockSender)(nil).SendGetAncestors), ctx, nodeID, requestID, containerID)
}

// SendGetStateSummaryFrontier mocks base method.
func (m *MockSender) SendGetStateSummaryFrontier(ctx context.Context, nodeIDs set.Set[ids.NodeID], requestID uint32) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendGetStateSummaryFrontier", ctx, nodeIDs, requestID)
}

// SendGetStateSummaryFrontier indicates an expected call of SendGetStateSummaryFrontier.
func (mr *MockSenderMockRecorder) SendGetStateSummaryFrontier(ctx, nodeIDs, requestID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendGetStateSummaryFrontier", reflect.TypeOf((*MockSender)(nil).SendGetStateSummaryFrontier), ctx, nodeIDs, requestID)
}

// SendPullQuery mocks base method.
func (m *MockSender) SendPullQuery(ctx context.Context, nodeIDs set.Set[ids.NodeID], requestID uint32, containerID ids.ID, requestedHeight uint64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendPullQuery", ctx, nodeIDs, requestID, containerID, requestedHeight)
}

// SendPullQuery indicates an expected call of SendPullQuery.
func (mr *MockSenderMockRecorder) SendPullQuery(ctx, nodeIDs, requestID, containerID, requestedHeight any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendPullQuery", reflect.TypeOf((*MockSender)(nil).SendPullQuery), ctx, nodeIDs, requestID, containerID, requestedHeight)
}

// SendPushQuery mocks base method.
func (m *MockSender) SendPushQuery(ctx context.Context, nodeIDs set.Set[ids.NodeID], requestID uint32, container []byte, requestedHeight uint64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendPushQuery", ctx, nodeIDs, requestID, container, requestedHeight)
}

// SendPushQuery indicates an expected call of SendPushQuery.
func (mr *MockSenderMockRecorder) SendPushQuery(ctx, nodeIDs, requestID, container, requestedHeight any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendPushQuery", reflect.TypeOf((*MockSender)(nil).SendPushQuery), ctx, nodeIDs, requestID, container, requestedHeight)
}

// SendPut mocks base method.
func (m *MockSender) SendPut(ctx context.Context, nodeID ids.NodeID, requestID uint32, container []byte) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendPut", ctx, nodeID, requestID, container)
}

// SendPut indicates an expected call of SendPut.
func (mr *MockSenderMockRecorder) SendPut(ctx, nodeID, requestID, container any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendPut", reflect.TypeOf((*MockSender)(nil).SendPut), ctx, nodeID, requestID, container)
}

// SendStateSummaryFrontier mocks base method.
func (m *MockSender) SendStateSummaryFrontier(ctx context.Context, nodeID ids.NodeID, requestID uint32, summary []byte) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendStateSummaryFrontier", ctx, nodeID, requestID, summary)
}

// SendStateSummaryFrontier indicates an expected call of SendStateSummaryFrontier.
func (mr *MockSenderMockRecorder) SendStateSummaryFrontier(ctx, nodeID, requestID, summary any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendStateSummaryFrontier", reflect.TypeOf((*MockSender)(nil).SendStateSummaryFrontier), ctx, nodeID, requestID, summary)
}
