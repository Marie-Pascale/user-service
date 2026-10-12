INSERT INTO services (id, name, description, api_key_hash, created_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'gestionnaire-usagers',
    'Internal service: represents the API itself in the permissions system',
    '',
    now()
);

INSERT INTO permissions (id, service_id, key, description) VALUES
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'users:read:any', 'View any user (not just oneself)'),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'users:write:any', 'Modify any user'),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'users:delete:any', 'Delete any user'),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'services:create', 'Create a new third-party service');

INSERT INTO roles (id, name, description)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    'admin',
    'Full administrator access to the user service'
);

INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000002', id
FROM permissions
WHERE service_id = '00000000-0000-0000-0000-000000000001';
