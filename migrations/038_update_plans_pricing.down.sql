-- Re-insert Beta Tester (was formerly Iniciante)
INSERT INTO plans (name, price, max_sites, max_ai_analyses, features)
VALUES ('Beta Tester', 9.90, 3, 10, ARRAY[
  'Monitoramento de até 3 sites',
  '10 análises de IA por mês',
  'Notificações por e-mail',
  'Preço especial para Beta Testers'
]);

-- Reassign Básico users back to Beta Tester
UPDATE users
SET plan_id = (SELECT id FROM plans WHERE name = 'Beta Tester')
WHERE plan_id = (SELECT id FROM plans WHERE name = 'Básico');

-- Remove Básico
DELETE FROM plans WHERE name = 'Básico';

-- Revert Profissional
UPDATE plans SET
  price = 29.90, max_sites = 15, max_ai_analyses = 100,
  features = ARRAY['Monitoramento de até 15 sites','100 análises IA/mês','Notificações por email','Suporte por email']
WHERE name = 'Profissional';

-- Revert Premium
UPDATE plans SET
  price = 79.90, max_sites = 50, max_ai_analyses = -1,
  features = ARRAY['Monitoramento de até 50 sites','Análises IA ilimitadas','Notificações por email','Suporte prioritário via Email e WhatsApp']
WHERE name = 'Premium';
