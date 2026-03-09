package cert

import "fmt"

type pemError struct{ file string }

func (e *pemError) Error() string { return fmt.Sprintf("invalid PEM data in %s", e.file) }

func errInvalidPEM(file string) error { return &pemError{file} }
