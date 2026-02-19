UPDATE plans
SET name = 'Iniciante',
    price = 0.00,
    features = ARRAY[
        'Monitoramento de até 3 sites',
        '10 análises de IA por mês',
        'Notificações por e-mail'
    ],
    updated_at = NOW()
WHERE name = 'Beta Tester';
