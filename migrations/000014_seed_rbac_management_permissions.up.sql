-- Nouvelles permissions internes pour la gestion du RBAC lui-meme
-- (Permission, Role, Group) -- voir application.PermPermissionsManage,
-- PermRolesManage, PermGroupsManage.
INSERT INTO permissions (id, service_id, key, description) VALUES
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'permissions:manage', 'Creer/supprimer des permissions'),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'roles:manage', 'Creer/supprimer des roles, gerer leurs permissions'),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'groups:manage', 'Creer/supprimer des groupes, gerer membres et roles');

-- Accordees automatiquement au role admin (meme role que celui seme par
-- la migration 000013).
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000002', id
FROM permissions
WHERE service_id = '00000000-0000-0000-0000-000000000001'
  AND key IN ('permissions:manage', 'roles:manage', 'groups:manage');
