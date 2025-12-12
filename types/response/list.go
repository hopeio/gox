package response

type List[T any] struct {
	List    []T  `json:"list"`
	Total   uint `json:"total,omitempty"`
	HasMore bool `json:"hasMore"`
}
