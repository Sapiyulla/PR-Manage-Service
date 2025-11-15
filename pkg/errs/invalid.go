package errs

type InvalidError struct {
	Domain string
	Desc   string
}

func (e *InvalidError) Error() string {
	if e.Desc == "" {
		return e.Domain + " invalid"
	}
	return e.Domain + " invalid: " + e.Desc
}
