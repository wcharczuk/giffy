package viewmodel

// CreateTagArgs is the post body the POST /api/tag method accepts.
type CreateTagArgs struct {
	TagValue  string   `json:"tag_value"`
	TagValues []string `json:"tag_values"`
}
