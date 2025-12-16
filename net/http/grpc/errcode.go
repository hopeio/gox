package grpc

import (
	"strconv"

	"github.com/hopeio/gox/errors"
	stringsx "github.com/hopeio/gox/strings"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCStatus interface {
	GRPCStatus() *status.Status
}

const (
	SysErr             = ErrCode(-1)
	Success            = ErrCode(codes.OK)
	Canceled           = ErrCode(codes.Canceled)
	Unknown            = ErrCode(codes.Unknown)
	InvalidArgument    = ErrCode(codes.InvalidArgument)
	DeadlineExceeded   = ErrCode(codes.DeadlineExceeded)
	NotFound           = ErrCode(codes.NotFound)
	AlreadyExists      = ErrCode(codes.AlreadyExists)
	PermissionDenied   = ErrCode(codes.PermissionDenied)
	ResourceExhausted  = ErrCode(codes.ResourceExhausted)
	FailedPrecondition = ErrCode(codes.FailedPrecondition)
	Aborted            = ErrCode(codes.Aborted)
	OutOfRange         = ErrCode(codes.OutOfRange)
	Unimplemented      = ErrCode(codes.Unimplemented)
	Internal           = ErrCode(codes.Internal)
	Unavailable        = ErrCode(codes.Unavailable)
	DataLoss           = ErrCode(codes.DataLoss)
	Unauthenticated    = ErrCode(codes.Unauthenticated)
)

func Register(code ErrCode, msg string) {
	errors.Register(errors.ErrCode(code), msg)
}

type ErrCode errors.ErrCode

func (x ErrCode) String() string {
	return errors.ErrCode(x).String()
}

func (x ErrCode) Error() string {
	return errors.ErrCode(x).Error()
}

func (x ErrCode) ErrResp() *ErrResp {
	return &ErrResp{Code: errors.ErrCode(x), Msg: x.String()}
}

func (x ErrCode) Msg(msg string) *ErrResp {
	return &ErrResp{Code: errors.ErrCode(x), Msg: msg}
}

func (x ErrCode) Wrap(err error) *ErrResp {
	return &ErrResp{Code: errors.ErrCode(x), Msg: err.Error()}
}

func (x ErrCode) GRPCStatus() *status.Status {
	return status.New(codes.Code(x), x.String())
}

type ErrResp errors.ErrResp

func NewErrResp(code ErrCode, msg string) *ErrResp {
	return &ErrResp{
		Code: errors.ErrCode(code),
		Msg:  msg,
	}
}

func ErrRespFrom(err error) *ErrResp {
	return (*ErrResp)(errors.ErrRespFrom(err))
}

func (x *ErrResp) GRPCStatus() *status.Status {
	return status.New(codes.Code(x.Code), x.Msg)
}

func (x *ErrResp) Error() string {
	return x.Msg
}

func (x *ErrResp) MarshalJSON() ([]byte, error) {
	return stringsx.ToBytes(`{"code":` + strconv.Itoa(int(x.Code)) + `,"msg":` + strconv.Quote(x.Msg) + `}`), nil
}

func (e *ErrResp) ErrResp() *errors.ErrResp {
	return (*errors.ErrResp)(e)
}
