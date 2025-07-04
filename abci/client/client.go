package abcicli

import (
	"context"
	"sync"

	"github.com/cometbft/cometbft/v2/abci/types"
	"github.com/cometbft/cometbft/v2/libs/service"
	cmtsync "github.com/cometbft/cometbft/v2/libs/sync"
)

const (
	dialRetryIntervalSeconds = 3
	echoRetryIntervalSeconds = 1
)

//go:generate ../../scripts/mockery_generate.sh Client

// Client defines the interface for an ABCI client.
//
// NOTE these are client errors, eg. ABCI socket connectivity issues.
// Application-related errors are reflected in response via ABCI error codes
// and (potentially) error response.
type Client interface {
	service.Service
	types.Application

	// TODO: remove as each method now returns an error
	Error() error
	// TODO: remove as this is not implemented
	Flush(ctx context.Context) error
	Echo(ctx context.Context, echo string) (*types.EchoResponse, error)

	// FIXME: All other operations are run synchronously and rely
	// on the caller to dictate concurrency (i.e. run a go routine),
	// with the exception of `CheckTxAsync` which we maintain
	// for the v0 mempool. We should explore refactoring the
	// mempool to remove this vestige behavior.
	//
	// SetResponseCallback is not used anymore. The callback was invoked only by the mempool on
	// CheckTx responses, only during rechecking. Now the responses are handled by the callback of
	// the *ReqRes struct returned by CheckTxAsync. This callback is more flexible as it allows to
	// pass other information such as the sender.
	//
	// Deprecated: Do not use.
	SetResponseCallback(cb Callback)
	CheckTxAsync(ctx context.Context, req *types.CheckTxRequest) (*ReqRes, error)
}

// ----------------------------------------

// NewClient returns a new ABCI client of the specified transport type.
// It returns an error if the transport is not "socket" or "grpc".
func NewClient(addr, transport string, mustConnect bool) (client Client, err error) {
	switch transport {
	case "socket":
		client = NewSocketClient(addr, mustConnect)
	case "grpc":
		client = NewGRPCClient(addr, mustConnect)
	default:
		err = ErrUnknownAbciTransport{Transport: transport}
	}
	return client, err
}

type Callback func(*types.Request, *types.Response)

type ReqRes struct {
	*types.Request
	*sync.WaitGroup
	*types.Response // Not set atomically, so be sure to use WaitGroup.

	mtx cmtsync.Mutex

	// callbackInvoked is a variable to track if the callback was already
	// invoked during the regular execution of the request. This variable
	// allows clients to set the callback simultaneously without potentially
	// invoking the callback twice by accident, once when 'SetCallback' is
	// called and once during the normal request.
	callbackInvoked bool
	cb              func(*types.Response) error // A single callback that may be set.
	cbErr           error
}

func NewReqRes(req *types.Request) *ReqRes {
	return &ReqRes{
		Request:   req,
		WaitGroup: waitGroup1(),
		Response:  nil,

		callbackInvoked: false,
		cb:              nil,
	}
}

// SetCallback sets the callback. If reqRes is already done, it will call the cb
// immediately. Note, reqRes.cb should not change if reqRes.done and only one
// callback is supported.
func (r *ReqRes) SetCallback(cb func(res *types.Response) error) {
	r.mtx.Lock()

	if r.callbackInvoked {
		r.mtx.Unlock()
		r.cbErr = cb(r.Response)
		return
	}

	r.cb = cb
	r.mtx.Unlock()
}

// InvokeCallback invokes a thread-safe execution of the configured callback
// if non-nil.
func (r *ReqRes) InvokeCallback() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if r.cb != nil && r.Response != nil {
		r.cbErr = r.cb(r.Response)
	}
	r.callbackInvoked = true
}

// Error returns the error returned by the callback, if any.
func (r *ReqRes) Error() error {
	return r.cbErr
}

func waitGroup1() (wg *sync.WaitGroup) {
	wg = &sync.WaitGroup{}
	wg.Add(1)
	return wg
}
