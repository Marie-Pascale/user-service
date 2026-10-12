INSERT INTO permissions (id, service_id, key, description) VALUES
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'services:manage', 'View, delete, or rotate a service''s API key');

INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000002', id
FROM permissions
WHERE service_id = '00000000-0000-0000-0000-000000000001'
  AND key = 'services:manage';
