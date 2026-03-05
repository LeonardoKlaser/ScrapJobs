# Design: React Email Templates

**Data:** 2026-03-05
**Status:** Aprovado

## Contexto

Os email templates atuais sao inline Go HTML com CSS basico e nao seguem o design system
do frontend. Precisamos padronizar usando React Email para gerar templates bonitos e
consistentes com a identidade visual do ScrapJobs.

## Decisoes

- Pacote isolado `email-templates/` na raiz do workspace
- React Email gera HTML com placeholders Go (`{{.Field}}`)
- Light mode apenas (compatibilidade com clientes de email)
- Layout base compartilhado (header + footer + container)
- Go embed para incluir HTMLs no binario (zero I/O runtime)

## Arquitetura

```
email-templates/
├── package.json
├── tsconfig.json
├── src/
│   ├── components/
│   │   └── base-layout.tsx
│   └── emails/
│       ├── welcome.tsx
│       ├── job-analysis.tsx
│       ├── new-jobs-alert.tsx
│       └── password-reset.tsx
├── scripts/
│   └── build.ts
└── dist/
    └── *.html

ScrapJobs/templates/
├── emails/*.html       ← copia do dist/
└── embed.go            ← //go:embed
```

### Visual

- Fundo: #ffffff, container card: #f9fafb
- Accent: #10b981 (emerald) para botoes e destaques
- Texto: #18181b (principal), #71717a (muted)
- Fonte: system font stack (-apple-system, Segoe UI, sans-serif)
- Header: "ScrapJobs" com accent emerald
- Footer: texto muted com link de contato
- Botoes: emerald bg, branco texto, border-radius 6px
- Container: 600px max-width

### Templates

**BaseLayout** — container, header, footer, recebe children e previewText

**Welcome** — saudacao + CTA dashboard
- Placeholders: `{{.UserName}}`, `{{.DashboardURL}}`

**JobAnalysis** — resultado analise AI
- Placeholders: `{{.UserName}}`, `{{.JobTitle}}`, `{{.CompanyName}}`, `{{.Score}}`,
  `{{.QualitativeScore}}`, `{{.Strengths}}`, `{{.Gaps}}`, `{{.Suggestions}}`,
  `{{.Considerations}}`, `{{.DashboardURL}}`

**NewJobsAlert** — tabela de vagas novas
- Placeholders: `{{.UserName}}`, `{{.SiteName}}`, `{{range .Jobs}}` (Title, Company, Location, URL)

**PasswordReset** — link de reset
- Placeholders: `{{.ResetURL}}`

### Build Pipeline

1. `npm run build` executa `scripts/build.ts`
2. Script importa cada email, chama `render()` do `@react-email/render`
3. Salva HTML em `dist/`
4. Dev copia `dist/*.html` para `ScrapJobs/templates/emails/`
5. Go usa `//go:embed`

### Mudancas no Go Backend

- `usecase/emailAdapter.go`: remove templates inline, carrega HTMLs embedded
- `html/template.Execute()` aplica dados nos placeholders
- Funcoes `generateXxxEmailBodyText()` permanecem (plain text fallback)
- Interface `EmailService` nao muda

### O que NAO muda

- Interface MailSender / EmailService
- EmailOrchestrator, controllers, processors
- Logica de envio (SES/Resend)
- Plain text fallbacks
