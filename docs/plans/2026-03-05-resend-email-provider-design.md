# Design: Multi-Provider Email (Resend + SES)

**Data:** 2026-03-05
**Status:** Aprovado

## Contexto

AWS SES ainda nao esta autorizado. Precisamos do Resend como provider primario.
Apos autorizacao do SES, ele sera mantido como contingencia. Admin pode configurar
ordem e ativar/desativar cada provider pelo painel.

## Decisoes

- Config de providers salva no banco (tabela `email_provider_config`)
- Fallback automatico com logging (primario falha -> tenta secundario -> loga fallback)
- API keys em env vars apenas (`RESEND_API_KEY`)
- UI no admin dashboard como nova secao (nao pagina separada)

## Arquitetura

```
EmailService interface (nao muda)
    |
SESSenderAdapter (nao muda - gera templates HTML/texto)
    |
EmailOrchestrator (NOVO - decide qual sender usar, fallback, logging)
    +-- SESMailSender (existente)
    +-- ResendMailSender (NOVO - mesma interface)
```

### Interface MailSender (nova)

Contrato que ambos os senders implementam:
`SendEmail(ctx, to, subject, textBody, htmlBody) error`

### ResendMailSender (novo, infra/resend/)

- SDK: `github.com/resend/resend-go/v2`
- Env vars: `RESEND_API_KEY`, `RESEND_SENDER_EMAIL`

### EmailOrchestrator (novo, usecase/)

- Substitui SESMailSender direto no SESSenderAdapter
- Consulta config do banco (cache 5 min)
- Tenta primario -> fallback -> log

### Tabela email_provider_config

```sql
CREATE TABLE email_provider_config (
    id SERIAL PRIMARY KEY,
    provider_name VARCHAR(20) NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    priority INTEGER NOT NULL DEFAULT 1,
    updated_at TIMESTAMP DEFAULT NOW(),
    updated_by INTEGER REFERENCES users(id)
);

INSERT INTO email_provider_config (provider_name, is_active, priority)
VALUES ('resend', true, 1), ('ses', true, 2);
```

### Endpoints Admin

- `GET /api/admin/email-config` - retorna config atual
- `PUT /api/admin/email-config` - atualiza ordem e ativacao

### Frontend

Nova secao no admin dashboard com:
- Cards para cada provider com toggle ativar/desativar
- Botoes para definir primario/secundario
- Indicador visual do provider primario

## O que NAO muda

- EmailService interface (4 metodos)
- SESSenderAdapter (templates)
- Controllers, processor, usecases
- Testes existentes
