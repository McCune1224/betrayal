package alliance

import "errors"

// Already exists errors
var (
	ErrAllianceAlreadyExists      = errors.New("alliance already exists")
	ErrCreateRequestAlreadyExists = errors.New("alliance request already exists")
	ErrInviteAlreadyExists        = errors.New("alliance invite already exists")
	ErrMemberAlreadyExists        = errors.New("alliance member already exists")
	ErrChannelAlreadyExists       = errors.New("alliance channel already exists")
	ErrAlreadyAllianceMember      = errors.New("user is already a member of an alliance")
)

// Not found errors
var (
	ErrAllianceNotFound = errors.New("alliance not found")
	ErrRequestNotFound  = errors.New("alliance request not found")
	ErrInviteNotFound   = errors.New("alliance invite not found")
	ErrMemberNotFound   = errors.New("alliance member not found")
)

// Misc errors
var (
	ErrAllianceMemberLimitExceeded = errors.New("alliance member limit exceeded")
	ErrOverrideRequired            = errors.New("alliance override required")
)
