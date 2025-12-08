package result

type List[T any] struct {
	List  []T  `json:"list"`
	Total uint `json:"total,omitempty"`
}
