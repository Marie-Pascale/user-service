INSERT INTO permissions (id, service_id, key, description) VALUES
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'permissions:manage', 'Create or delete permissions'),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'roles:manage', 'Create or delete roles and manage their permissions'),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'groups:manage', 'Create or delete groups and manage their members and roles');

INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000002', id
FROM permissions
WHERE service_id = '00000000-0000-0000-0000-000000000001'
  AND key IN ('permissions:manage', 'roles:manage', 'groups:manage');
