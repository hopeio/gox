package response

import "github.com/hopeio/gox/net/http"

type List[T any] struct {
	List  []T  `json:"list"`
	Total uint `json:"total,omitempty"`
}

type CursorList[T, ID any] struct {
	List    []T  `json:"list"`
	Total   uint `json:"total,omitempty"`
	Cursor  ID
	HasMore bool `json:"hasMore"`
}


type CommonResp[T any] = http.CommonResp[T]