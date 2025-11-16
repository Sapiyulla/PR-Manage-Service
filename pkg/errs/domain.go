package errs

import "pr-manage-service/pkg/codes"

type DomainError struct {
	Code codes.CODE
}

func (e *DomainError) Error() string {
	switch e.Code {
	case codes.NO_CANDIDATE:
		return "no active replacement candidate in team"
	case codes.PR_MERGED:
		return "cannot reassign on merged PR"
	default: // codes.NOT_ASSIGNED
		return "reviewer is not assigned to this PR"
	}
}
