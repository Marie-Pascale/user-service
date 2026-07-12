-- Le service "gestionnaire-usagers" lui-meme, pour proteger ses propres
-- routes avec le meme systeme RBAC que les services tiers de l'ecosysteme.
-- ID fixe -- doit rester synchronise avec application.SelfServiceID.
-- api_key_hash est vide : ce "service" n'est jamais appele de l'exterieur
-- via X-API-Key, il sert uniquement de point d'ancrage pour ses propres
-- permissions internes.
INSERT INTO services (id, name, description, api_key_hash, created_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'gestionnaire-usagers',
    'Service interne : represente l''API elle-meme dans le systeme de permissions',
    '',
    now()
);

-- Permissions reconnues par ce service.
-- Voir application.Perm* pour les constantes Go correspondantes.
INSERT INTO permissions (id, service_id, key, description) VALUES
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'users:read:any', 'Voir n''importe quel usager (pas seulement soi-meme)'),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'users:write:any', 'Modifier n''importe quel usager'),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'users:delete:any', 'Supprimer n''importe quel usager'),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'services:create', 'Creer un nouveau service tiers');

-- Le role admin, avec toutes les permissions ci-dessus.
-- ID fixe -- doit rester synchronise avec application.AdminRoleID.
INSERT INTO roles (id, name, description)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    'admin',
    'Acces administrateur complet a gestionnaire-usagers'
);

INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000002', id
FROM permissions
WHERE service_id = '00000000-0000-0000-0000-000000000001';
