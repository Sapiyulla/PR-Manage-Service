package errs

type NotFoundError struct {
	Domain string
	Desc   string
}

func (e *NotFoundError) Error() string {
	if e.Desc != "" {
		return e.Domain + " not found: " + e.Desc
	}
	return e.Domain + " not found"
}
