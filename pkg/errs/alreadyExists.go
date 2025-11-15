package errs

type AlreadyExistsError struct {
	Domain string
	Desc   string
}

func (e *AlreadyExistsError) Error() string {
	if e.Desc != "" {
		return e.Domain + " already exists: " + e.Desc
	}
	return e.Domain + " already exists"
}
