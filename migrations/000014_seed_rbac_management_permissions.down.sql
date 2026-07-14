-- role_permissions correspondantes supprimees en cascade.
DELETE FROM permissions
WHERE service_id = '00000000-0000-0000-0000-000000000001'
  AND key IN ('permissions:manage', 'roles:manage', 'groups:manage');
