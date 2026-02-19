UPDATE plans
SET name = 'Beta Tester',
    price = 9.90,
    features = ARRAY[
        'Monitoramento de até 3 sites',
        '10 análises de IA por mês',
        'Notificações por e-mail',
        'Preço especial para Beta Testers'
    ],
    updated_at = NOW()
WHERE name = 'Iniciante';
