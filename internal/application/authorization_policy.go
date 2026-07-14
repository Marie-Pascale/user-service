package application

// Ce fichier definit la politique d'autorisation de CETTE API sur ses
// PROPRES routes -- distinct du systeme RBAC generique (Role, Permission,
// Group) qui, lui, sert aussi aux services tiers de l'ecosysteme.
//
// gestionnaire-usagers se traite lui-meme comme un service parmi
// d'autres dans le systeme de permissions -- exactement comme
// "orders-service" ou "billing-service" le seraient. C'est ce qui permet
// de reutiliser AuthorizationService.HasPermission tel quel pour proteger
// ses propres routes, plutot que d'inventer un mecanisme separe.

// SelfServiceID identifie ce service dans les tables services/permissions.
// C'est un ID FIXE (pas genere par gen_random_uuid() a l'execution) parce
// que le code a besoin de le connaitre a la compilation. Seme par une
// migration -- voir migrations/000013_seed_self_service.up.sql, qui doit
// rester synchronisee avec cette constante.
const SelfServiceID = "00000000-0000-0000-0000-000000000001"

// AdminRoleID identifie le role "admin" seme par la meme migration.
// Fixe pour la meme raison : permettre de l'assigner a un usager (via
// `make grant-admin`) sans avoir a le chercher par nom au prealable.
const AdminRoleID = "00000000-0000-0000-0000-000000000002"

// Cles de permission reconnues par ce service, sous SelfServiceID.
// Centralisees ici plutot qu'eparpillees en chaines litterales dans les
// handlers HTTP -- une faute de frappe devient une erreur de compilation
// plutot qu'un bug silencieux (une permission qui ne matche jamais rien).
const (
	PermUsersReadAny   = "users:read:any"
	PermUsersWriteAny  = "users:write:any"
	PermUsersDeleteAny = "users:delete:any"
	PermServicesCreate = "services:create"

	// PermServicesManage couvre consultation, suppression et rotation de
	// cle -- separee de PermServicesCreate parce qu'on pourrait vouloir
	// deleguer la creation de services (ex: a une equipe plateforme) sans
	// donner acces a la gestion (rotation de cle) de services deja
	// existants crees par d'autres.
	PermServicesManage = "services:manage"

	// Gestion du RBAC lui-meme -- creer/supprimer des permissions, roles,
	// groupes, et gerer leurs associations. Un seul niveau de permission
	// par entite (pas de distinction read/write/delete comme pour User) :
	// ces entites n'ont pas de notion de "soi-meme", donc pas de raison
	// de subdiviser davantage pour l'instant.
	PermPermissionsManage = "permissions:manage"
	PermRolesManage       = "roles:manage"
	PermGroupsManage      = "groups:manage"
)
