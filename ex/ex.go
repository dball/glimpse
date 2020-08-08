package ex

import "fmt"

// Ex is an error with a code, possibly an error, and a context map
type Ex struct {
	Code    string
	Err     error
	Context map[string]interface{}
}

func (ex Ex) Unwrap() error { return ex.Err }

func (ex Ex) String() string {
	return fmt.Sprintf("error: %v: %v: %v", ex.Code, ex.Err, ex.Context)
}

func (ex Ex) Error() string {
	return ex.String()
}
