package groups

// CreateGroupRequest represents the request body for creating a group
type CreateGroupRequest struct {
	Name string `json:"name"`
}

// MembersRequest represents the request body for adding or removing users from a group
type MembersRequest struct {
	Usernames []string `json:"usernames"`
}

// GroupMemberResponse represents a group member in the response
type GroupMemberResponse struct {
	Username string `json:"username"`
}

// GroupRoleRequest represents the request body for adding or removing a role from a group
type GroupRoleRequest struct {
	Application string `json:"application"`
	Role        string `json:"role"`
}

// GroupRoleResponse represents a group role assignment in the response
type GroupRoleResponse struct {
	Application string `json:"application"`
	Role        string `json:"role"`
}
