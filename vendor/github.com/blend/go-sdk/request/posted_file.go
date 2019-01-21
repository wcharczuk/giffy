package request

import "io"

//--------------------------------------------------------------------------------
// PostedFile
//--------------------------------------------------------------------------------

// PostedFile represents a file to post with the request.
type PostedFile struct {
	Key          string
	FileName     string
	FileContents io.Reader
}
