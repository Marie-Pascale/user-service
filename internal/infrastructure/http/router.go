package httpapi

import (
	"net/http"

	"gestionnaire-usagers/internal/application"
)

// NewRouter cable toutes les routes de l'API sur un ServeMux (Go 1.22+,
// avec le support natif des methodes HTTP et des variables de chemin --
// voir la discussion sur ServeMux plus tot dans le projet).
//
// Chaque middleware n'enveloppe QUE les routes qui en ont besoin -- pas le
// mux entier (voir la discussion sur le placement des middlewares plus tot
// dans le projet). Pour les routes protegees, l'ordre de chainage est
// TOUJOURS : session (authentification) -> permission (autorisation) ->
// handler. On ne peut pas verifier une permission sans savoir d'abord QUI
// fait la requete.
func NewRouter(
	userHandler *UserHandler,
	authHandler *AuthHandler,
	serviceHandler *ServiceHandler,
	permissionsHandler *PermissionsHandler,
	permissionHandler *PermissionHandler,
	roleHandler *RoleHandler,
	groupHandler *GroupHandler,
	serviceAuth *ServiceAuthMiddleware,
	userAuth *UserAuthMiddleware,
	authz *AuthorizationMiddleware,
) *http.ServeMux {
	mux := http.NewServeMux()

	// protected() combine session + permission -- raccourci pour eviter de
	// repeter userAuth.Authenticate(authz.RequirePermission(...)(...)) sur
	// chaque route de gestion RBAC ci-dessous.
	protected := func(permissionKey string, h http.HandlerFunc) http.Handler {
		return userAuth.Authenticate(authz.RequirePermission(permissionKey)(h))
	}

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /docs", docsHandler)
	mux.HandleFunc("GET /openapi.yaml", specHandler)

	// Ouverte : auto-inscription, aucune authentification requise.
	mux.HandleFunc("POST /users", userHandler.Create)

	// Ouverte : il faut bien pouvoir se connecter avant d'avoir un jeton.
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	// Protegees par session seulement : usager authentifie requis.
	mux.Handle("GET /me", userAuth.Authenticate(http.HandlerFunc(userHandler.Me)))
	mux.Handle("POST /auth/logout", userAuth.Authenticate(http.HandlerFunc(authHandler.Logout)))

	// Usagers -- protection "self-ou-permission" (voir AuthorizationMiddleware).
	mux.Handle("GET /users", protected(application.PermUsersReadAny, userHandler.List))
	mux.Handle("GET /users/{id}",
		userAuth.Authenticate(authz.RequireSelfOrPermission(application.PermUsersReadAny)(http.HandlerFunc(userHandler.Get))))
	mux.Handle("PUT /users/{id}",
		userAuth.Authenticate(authz.RequireSelfOrPermission(application.PermUsersWriteAny)(http.HandlerFunc(userHandler.Update))))
	mux.Handle("DELETE /users/{id}",
		userAuth.Authenticate(authz.RequireSelfOrPermission(application.PermUsersDeleteAny)(http.HandlerFunc(userHandler.Delete))))
	mux.Handle("POST /users/{id}/password",
		userAuth.Authenticate(authz.RequireSelfOrPermission(application.PermUsersWriteAny)(http.HandlerFunc(authHandler.SetPassword))))

	// Services tiers -- creation reservee a "services:create", le reste
	// (consultation, suppression, rotation de cle) a "services:manage".
	mux.Handle("POST /services", protected(application.PermServicesCreate, serviceHandler.Create))
	mux.Handle("GET /services", protected(application.PermServicesManage, serviceHandler.List))
	mux.Handle("GET /services/{id}", protected(application.PermServicesManage, serviceHandler.Get))
	mux.Handle("PUT /services/{id}", protected(application.PermServicesManage, serviceHandler.Update))
	mux.Handle("DELETE /services/{id}", protected(application.PermServicesManage, serviceHandler.Delete))
	mux.Handle("POST /services/{id}/rotate-key", protected(application.PermServicesManage, serviceHandler.RotateAPIKey))

	// Permissions -- gestion RBAC, aucune notion de "soi-meme".
	mux.Handle("POST /permissions", protected(application.PermPermissionsManage, permissionHandler.Create))
	mux.Handle("GET /permissions", protected(application.PermPermissionsManage, permissionHandler.ListByService))
	mux.Handle("GET /permissions/{id}", protected(application.PermPermissionsManage, permissionHandler.Get))
	mux.Handle("DELETE /permissions/{id}", protected(application.PermPermissionsManage, permissionHandler.Delete))

	// Auto-enregistrement par un service tiers -- protegee par cle API
	// (pas par session usager), comme GET /users/{id}/permissions plus bas.
	mux.Handle("POST /permissions/self", serviceAuth.Authenticate(http.HandlerFunc(permissionHandler.RegisterSelf)))

	// Roles
	mux.Handle("POST /roles", protected(application.PermRolesManage, roleHandler.Create))
	mux.Handle("GET /roles", protected(application.PermRolesManage, roleHandler.List))
	mux.Handle("GET /roles/{id}", protected(application.PermRolesManage, roleHandler.Get))
	mux.Handle("PUT /roles/{id}", protected(application.PermRolesManage, roleHandler.Update))
	mux.Handle("DELETE /roles/{id}", protected(application.PermRolesManage, roleHandler.Delete))
	mux.Handle("POST /roles/{id}/permissions", protected(application.PermRolesManage, roleHandler.GrantPermission))
	mux.Handle("DELETE /roles/{id}/permissions/{permissionId}", protected(application.PermRolesManage, roleHandler.RevokePermission))

	// Groupes
	mux.Handle("POST /groups", protected(application.PermGroupsManage, groupHandler.Create))
	mux.Handle("GET /groups", protected(application.PermGroupsManage, groupHandler.List))
	mux.Handle("GET /groups/{id}", protected(application.PermGroupsManage, groupHandler.Get))
	mux.Handle("PUT /groups/{id}", protected(application.PermGroupsManage, groupHandler.Update))
	mux.Handle("DELETE /groups/{id}", protected(application.PermGroupsManage, groupHandler.Delete))
	mux.Handle("POST /groups/{id}/members", protected(application.PermGroupsManage, groupHandler.AddMember))
	mux.Handle("DELETE /groups/{id}/members/{userId}", protected(application.PermGroupsManage, groupHandler.RemoveMember))
	mux.Handle("POST /groups/{id}/roles", protected(application.PermGroupsManage, groupHandler.AssignRole))

	// Protegee par cle API (services tiers, pas des usagers).
	mux.Handle("GET /users/{id}/profile", serviceAuth.Authenticate(http.HandlerFunc(userHandler.GetBasicInfo)))
	mux.Handle("GET /users/{id}/permissions", serviceAuth.Authenticate(http.HandlerFunc(permissionsHandler.ListForUser)))
	mux.Handle("GET /users/{id}/permissions/check", serviceAuth.Authenticate(http.HandlerFunc(permissionsHandler.CheckForUser)))

	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
