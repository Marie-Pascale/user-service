package application

const SelfServiceID = "00000000-0000-0000-0000-000000000001"

const AdminRoleID = "00000000-0000-0000-0000-000000000002"

const (
	PermUsersReadAny   = "users:read:any"
	PermUsersWriteAny  = "users:write:any"
	PermUsersDeleteAny = "users:delete:any"
	PermServicesCreate = "services:create"

	PermServicesManage = "services:manage"

	PermPermissionsManage = "permissions:manage"
	PermRolesManage       = "roles:manage"
	PermGroupsManage      = "groups:manage"
)

var AllPermissionKeys = []string{
	PermUsersReadAny,
	PermUsersWriteAny,
	PermUsersDeleteAny,
	PermServicesCreate,
	PermServicesManage,
	PermPermissionsManage,
	PermRolesManage,
	PermGroupsManage,
}
