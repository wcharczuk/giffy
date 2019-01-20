package model

// ImagePredicate is used in linq queries
type ImagePredicate func(i Image) bool

// NewImagePredicate creates a new ImagePredicate that resolves to a linq.Predicate
func NewImagePredicate(predicate ImagePredicate) func(item interface{}) bool {
	return func(item interface{}) bool {
		if typed, isTyped := item.(Image); isTyped {
			return predicate(typed)
		}
		return false
	}
}
