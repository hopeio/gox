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
	// SysErr ErrCode = -1
	Success            ErrCode = 0
	Canceled           ErrCode = 1
	Unknown            ErrCode = 2
	InvalidArgument    ErrCode = 3
	DeadlineExceeded   ErrCode = 4
	NotFound           ErrCode = 5
	AlreadyExists      ErrCode = 6
	PermissionDenied   ErrCode = 7
	ResourceExhausted  ErrCode = 8
	FailedPrecondition ErrCode = 9
	Aborted            ErrCode = 10
	OutOfRange         ErrCode = 11
	Unimplemented      ErrCode = 12
	Internal           ErrCode = 13
	Unavailable        ErrCode = 14
	DataLoss           ErrCode = 15
	Unauthenticated    ErrCode = 16
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

func (x ErrCode) ErrRep() *ErrRep {
	return &ErrRep{Code: errors.ErrCode(x), Msg: x.String()}
}

func (x ErrCode) Msg(msg string) *ErrRep {
	return &ErrRep{Code: errors.ErrCode(x), Msg: msg}
}

func (x ErrCode) Wrap(err error) *ErrRep {
	return &ErrRep{Code: errors.ErrCode(x), Msg: err.Error()}
}

func (x ErrCode) GRPCStatus() *status.Status {
	return status.New(codes.Code(x), x.String())
}

type ErrRep errors.ErrRep

func NewErrRep(code ErrCode, msg string) *ErrRep {
	return &ErrRep{
		Code: errors.ErrCode(code),
		Msg:  msg,
	}
}

func ErrRepFrom(err error) *ErrRep {
	return (*ErrRep)(errors.ErrRepFrom(err))
}

func (x *ErrRep) GRPCStatus() *status.Status {
	return status.New(codes.Code(x.Code), x.Msg)
}

func (x *ErrRep) Error() string {
	return x.Msg
}

func (x *ErrRep) MarshalJSON() ([]byte, error) {
	return stringsx.ToBytes(`{"code":` + strconv.Itoa(int(x.Code)) + `,"msg":` + strconv.Quote(x.Msg) + `}`), nil
}
