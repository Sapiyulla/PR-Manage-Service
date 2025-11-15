package errs

type DublicateError struct {
	Domain string
}

func (e *DublicateError) Error() string {
	return e.Domain + " dublicate error"
}
