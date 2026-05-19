package errors

import (
	"errors"
	"testing"
)

func TestErrCode_String(t *testing.T) {
	tests := []struct {
		code ErrCode
		want string
	}{
		{Success, "Success"},
		{Canceled, "Canceled"},
		{Unknown, "Unknown"},
		{NotFound, "NotFound"},
		{Internal, "Internal"},
		{ErrCode(9999), "Unknown Error, Code:9999"},
	}
	for _, tt := range tests {
		if got := tt.code.String(); got != tt.want {
			t.Errorf("ErrCode(%d).String() = %q, want %q", tt.code, got, tt.want)
		}
	}
}

func TestErrCode_ErrResp(t *testing.T) {
	resp := NotFound.ErrResp()
	if resp.Code != NotFound {
		t.Errorf("ErrResp().Code = %v, want %v", resp.Code, NotFound)
	}
	if resp.Msg != "NotFound" {
		t.Errorf("ErrResp().Msg = %q, want %q", resp.Msg, "NotFound")
	}
}

func TestErrCode_Msg(t *testing.T) {
	resp := NotFound.Msg("custom message")
	if resp.Code != NotFound {
		t.Errorf("Msg().Code = %v, want %v", resp.Code, NotFound)
	}
	if resp.Msg != "custom message" {
		t.Errorf("Msg().Msg = %q, want %q", resp.Msg, "custom message")
	}
}

func TestErrCode_Wrap(t *testing.T) {
	err := errors.New("some error")
	resp := Internal.Wrap(err)
	if resp.Code != Internal {
		t.Errorf("Wrap().Code = %v, want %v", resp.Code, Internal)
	}
	if resp.Msg != "some error" {
		t.Errorf("Wrap().Msg = %q, want %q", resp.Msg, "some error")
	}
}

func TestErrCode_Error(t *testing.T) {
	if Success.Error() != "Success" {
		t.Errorf("Success.Error() = %q, want %q", Success.Error(), "Success")
	}
}

func TestNewErrResp(t *testing.T) {
	resp := NewErrResp(NotFound, "not found")
	if resp.Code != NotFound {
		t.Errorf("NewErrResp().Code = %v, want %v", resp.Code, NotFound)
	}
	if resp.Msg != "not found" {
		t.Errorf("NewErrResp().Msg = %q, want %q", resp.Msg, "not found")
	}
}

func TestErrResp_Error(t *testing.T) {
	resp := NewErrResp(NotFound, "not found")
	if resp.Error() != "not found" {
		t.Errorf("ErrResp.Error() = %q, want %q", resp.Error(), "not found")
	}
}

func TestErrRespFrom(t *testing.T) {
	// nil error
	if resp := ErrRespFrom(nil); resp != nil {
		t.Errorf("ErrRespFrom(nil) = %v, want nil", resp)
	}
	// *ErrResp
	orig := NewErrResp(NotFound, "not found")
	if resp := ErrRespFrom(orig); resp != orig {
		t.Errorf("ErrRespFrom(*ErrResp) should return same *ErrResp")
	}
	// regular error
	resp := ErrRespFrom(errors.New("test"))
	if resp.Code != Unknown {
		t.Errorf("ErrRespFrom(error).Code = %v, want %v", resp.Code, Unknown)
	}
	if resp.Msg != "test" {
		t.Errorf("ErrRespFrom(error).Msg = %q, want %q", resp.Msg, "test")
	}
}

func TestRegister(t *testing.T) {
	code := ErrCode(100)
	Register(code, "CustomError")
	if code.String() != "CustomError" {
		t.Errorf("After Register: ErrCode(100).String() = %q, want %q", code.String(), "CustomError")
	}
}
