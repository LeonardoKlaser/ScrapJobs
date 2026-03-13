-- Remove unlimited analyses from Premium plan: change -1 to 50
UPDATE plans SET
  max_ai_analyses = 50,
  features = ARRAY[
    'Monitoramento de até 30 sites',
    '50 análises IA/mês',
    'Notificações por email',
    'Suporte prioritário via Email e WhatsApp'
  ]
WHERE name = 'Premium';
