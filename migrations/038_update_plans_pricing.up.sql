-- Insert Básico plan first (needed before reassigning Beta Tester users)
INSERT INTO plans (name, price, max_sites, max_ai_analyses, features)
VALUES ('Básico', 9.90, 3, 15, ARRAY[
  'Monitoramento de até 3 sites',
  '15 análises IA/mês',
  'Notificações por email'
]);

-- Reassign Beta Tester users to Básico + give 30-day courtesy
UPDATE users
SET plan_id = (SELECT id FROM plans WHERE name = 'Básico'),
    expires_at = GREATEST(COALESCE(expires_at, NOW()), NOW()) + INTERVAL '30 days'
WHERE plan_id = (SELECT id FROM plans WHERE name = 'Beta Tester');

-- Delete Beta Tester plan (no more FK references)
DELETE FROM plans WHERE name = 'Beta Tester';

-- Update Profissional plan
UPDATE plans SET
  price = 16.90,
  max_sites = 8,
  max_ai_analyses = 40,
  features = ARRAY[
    'Monitoramento de até 8 sites',
    '40 análises IA/mês',
    'Notificações por email',
    'Suporte por email'
  ]
WHERE name = 'Profissional';

-- Update Premium plan
UPDATE plans SET
  price = 19.90,
  max_sites = 30,
  max_ai_analyses = -1,
  features = ARRAY[
    'Monitoramento de até 30 sites',
    'Análises IA ilimitadas',
    'Notificações por email',
    'Suporte prioritário via Email e WhatsApp'
  ]
WHERE name = 'Premium';
