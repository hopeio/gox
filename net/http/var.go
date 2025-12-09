package http

var (
	DefaultMarshaler Marshaler = &Json{}
	RespSysErr, _              = DefaultMarshaler.Marshal(&ErrResp{Code: -1, Msg: "system error"})
	RespOk, _                  = DefaultMarshaler.Marshal(&ErrResp{})
)
