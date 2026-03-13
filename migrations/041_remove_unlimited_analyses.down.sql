-- Revert Premium plan back to unlimited analyses
UPDATE plans SET
  max_ai_analyses = -1,
  features = ARRAY[
    'Monitoramento de até 30 sites',
    'Análises IA ilimitadas',
    'Notificações por email',
    'Suporte prioritário via Email e WhatsApp'
  ]
WHERE name = 'Premium';
