-- Permission pour consulter/supprimer/faire pivoter la cle d'un service
-- existant -- separee de "services:create" (voir application.PermServicesManage).
INSERT INTO permissions (id, service_id, key, description) VALUES
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'services:manage', 'Consulter, supprimer ou faire pivoter la cle API d''un service');

-- Accordee automatiquement au role admin.
INSERT INTO role_permissions (role_id, permission_id)
SELECT '00000000-0000-0000-0000-000000000002', id
FROM permissions
WHERE service_id = '00000000-0000-0000-0000-000000000001'
  AND key = 'services:manage';
