package httpapi

import (
	"time"

	"gestionnaire-usagers/internal/domain"
)

type userResponse struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	Status        string     `json:"status"`
	RoleIDs       []string   `json:"role_ids"`
	PermissionIDs []string   `json:"permission_ids"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastLoginAt   *time.Time `json:"last_login_at"`
}

func newUserResponse(u *domain.User) userResponse {
	return userResponse{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		Status:        string(u.Status),
		RoleIDs:       u.RoleIDs,
		PermissionIDs: u.PermissionIDs,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
		LastLoginAt:   u.LastLoginAt,
	}
}

type userBasicInfoResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

func newUserBasicInfoResponse(u *domain.User) userBasicInfoResponse {
	return userBasicInfoResponse{ID: u.ID, Name: u.Name, Email: u.Email, Status: string(u.Status)}
}

type authMethodResponse struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Provider   string    `json:"provider"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
}

func newAuthMethodResponse(am *domain.AuthMethod) authMethodResponse {
	return authMethodResponse{
		ID:         am.ID,
		UserID:     am.UserID,
		Provider:   string(am.Provider),
		CreatedAt:  am.CreatedAt,
		LastUsedAt: am.LastUsedAt,
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

type createUserRequest struct {
	Name       string            `json:"name"`
	Email      string            `json:"email"`
	AuthMethod authMethodRequest `json:"auth_method"`
}

type authMethodRequest struct {
	Provider       string `json:"provider,omitempty"`
	Password       string `json:"password,omitempty"`
	ProviderUserID string `json:"provider_user_id,omitempty"`
}

type updateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type setPasswordRequest struct {
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	User      userResponse `json:"user"`
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
}

type serviceResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func newServiceResponse(s *domain.Service) serviceResponse {
	return serviceResponse{ID: s.ID, Name: s.Name, Description: s.Description, CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt}
}

type createServiceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type createServiceResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	APIKey      string    `json:"api_key"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type permissionResponse struct {
	ID          string    `json:"id"`
	ServiceID   string    `json:"service_id"`
	Key         string    `json:"key"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func newPermissionResponse(p *domain.Permission) permissionResponse {
	return permissionResponse{
		ID:          p.ID,
		ServiceID:   p.ServiceID,
		Key:         p.Key,
		Description: p.Description,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

type createPermissionRequest struct {
	ServiceID   string `json:"service_id"`
	Key         string `json:"key"`
	Description string `json:"description"`
}

type registerPermissionRequest struct {
	Key         string `json:"key"`
	Description string `json:"description"`
}

type roleResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	PermissionIDs []string  `json:"permission_ids"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func newRoleResponse(r *domain.Role) roleResponse {
	return roleResponse{
		ID:            r.ID,
		Name:          r.Name,
		Description:   r.Description,
		PermissionIDs: r.PermissionIDs,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

type createRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type roleGrantPermissionRequest struct {
	PermissionID string `json:"permission_id"`
}

type groupResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	MemberUserIDs []string  `json:"member_user_ids"`
	RoleIDs       []string  `json:"role_ids"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func newGroupResponse(g *domain.Group) groupResponse {
	return groupResponse{
		ID:            g.ID,
		Name:          g.Name,
		Description:   g.Description,
		MemberUserIDs: g.MemberUserIDs,
		RoleIDs:       g.RoleIDs,
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
	}
}

type createGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type groupAddMemberRequest struct {
	UserID string `json:"user_id"`
}

type groupAssignRoleRequest struct {
	RoleID string `json:"role_id"`
}
