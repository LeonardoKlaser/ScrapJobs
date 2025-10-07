CREATE TABLE plans (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    price DECIMAL(10, 2) NOT NULL,
    max_sites INT NOT NULL,
    max_ai_analyses INT NOT NULL,
    features TEXT[] NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE users
ADD COLUMN plan_id INTEGER REFERENCES plans(id);

INSERT INTO plans (name, price, max_sites, max_ai_analyses, features) VALUES
(
    'Iniciante',
    0.00,
    3,
    10,
    ARRAY[
        'Monitoramento de até 3 sites',
        '10 análises de IA por mês',
        'Notificações por e-mail'
    ]
),
(
    'Profissional',
    29.90,
    15,
    100,
    ARRAY[
        'Monitoramento de até 15 sites',
        '100 análises de IA por mês',
        'Notificações por e-mail',
        'Suporte via e-mail'
    ]
),
(
    'Premium',
    79.90,
    50,
    -1, 
    ARRAY[
        'Monitoramento de até 50 sites',
        'Análises de IA ilimitadas',
        'Notificações por e-mail',
        'Suporte prioritário via E-mail e WhatsApp'
    ]
);

UPDATE users SET plan_id = (SELECT id FROM plans WHERE name = 'Iniciante') WHERE plan_id IS NULL;
