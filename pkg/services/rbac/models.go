package rbac

import (
	"fmt"
	"time"
)

var (
	errPolicyNotFound         = fmt.Errorf("could not find policy")
	errTeamPolicyAlreadyAdded = fmt.Errorf("policy is already added to this team")
	errTeamMemberNotFound     = fmt.Errorf("team policy not found")
	errTeamNotFound           = fmt.Errorf("team not found")
)

// Policy is the model for Policy in RBAC.
type Policy struct {
	Id          int64
	OrgId       int64
	Name        string
	Description string

	Updated time.Time
	Created time.Time
}

type PolicyDTO struct {
	Id          int64
	OrgId       int64
	Name        string
	Description string
	Permissions []Permission

	Updated time.Time
	Created time.Time
}

// Policy is the model for Permission in RBAC.
type Permission struct {
	Id           int64
	OrgId        int64
	PolicyId     int64
	Resource     string
	ResourceType string
	Action       string

	Updated time.Time
	Created time.Time
}

type TeamPolicy struct {
	Id       int64
	OrgId    int64
	PolicyId int64
	TeamId   int64

	Updated time.Time
	Created time.Time
}

type ListPoliciesQuery struct {
	OrgId int64 `json:"-"`

	Result []*Policy
}

type GetPolicyQuery struct {
	OrgId    int64 `json:"-"`
	PolicyId int64

	Result *PolicyDTO
}

type GetPolicyPermissionsQuery struct {
	OrgId    int64 `json:"-"`
	PolicyId int64

	Result []Permission
}

type GetTeamPoliciesQuery struct {
	OrgId  int64 `json:"-"`
	TeamId int64

	Result []*PolicyDTO
}

type CreatePermissionCommand struct {
	OrgId        int64
	PolicyId     int64
	Resource     string
	ResourceType string
	Action       string

	Result *Permission
}

type DeletePermissionCommand struct {
	Id    int64
	OrgId int64
}

type CreatePolicyCommand struct {
	OrgId       int64
	Name        string
	Description string

	Result *Policy
}

type DeletePolicyCommand struct {
	Id    int64
	OrgId int64
}

type AddTeamPolicyCommand struct {
	OrgId    int64
	PolicyId int64
	TeamId   int64
}

type RemoveTeamPolicyCommand struct {
	OrgId    int64
	PolicyId int64
	TeamId   int64
}
