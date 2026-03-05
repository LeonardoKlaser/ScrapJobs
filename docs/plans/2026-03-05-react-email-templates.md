# React Email Templates Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace inline Go HTML email templates with React Email components that match the ScrapJobs design system, outputting HTML with Go template placeholders.

**Architecture:** A standalone Node.js package (`email-templates/`) uses React Email to build 4 email templates with a shared base layout. A build script renders them to static HTML files containing Go `{{.Field}}` placeholders. The Go backend embeds these files via `//go:embed` and executes them with `html/template` at runtime.

**Tech Stack:** React Email (`@react-email/components`, `@react-email/render`), TypeScript, tsx (runner), Go `embed` + `html/template`

---

### Task 1: Scaffold email-templates package

**Files:**
- Create: `email-templates/package.json`
- Create: `email-templates/tsconfig.json`

**Step 1: Create package.json**

```json
{
  "name": "scrapjobs-email-templates",
  "version": "1.0.0",
  "private": true,
  "scripts": {
    "dev": "email dev",
    "build": "tsx scripts/build.ts"
  },
  "dependencies": {
    "@react-email/components": "^0.0.36",
    "@react-email/render": "^1.0.5",
    "react": "^19.0.0"
  },
  "devDependencies": {
    "@types/react": "^19.0.0",
    "tsx": "^4.19.0",
    "typescript": "^5.8.0",
    "react-email": "^3.0.0"
  }
}
```

**Step 2: Create tsconfig.json**

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "jsx": "react-jsx",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "outDir": "dist",
    "rootDir": "src"
  },
  "include": ["src/**/*", "scripts/**/*"]
}
```

**Step 3: Install dependencies**

Run: `cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs/email-templates && npm install`
Expected: `node_modules/` created, no errors

**Step 4: Verify React Email dev server works**

Run: `cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs/email-templates && npx email dev`
Expected: Dev server starts (can cancel with Ctrl+C after confirming). If no emails dir found, that's fine — we'll create them next.

**Step 5: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add email-templates/package.json email-templates/tsconfig.json email-templates/package-lock.json
git commit -m "feat: scaffold email-templates package with React Email"
```

> **Note:** Add `email-templates/node_modules/` to `.gitignore` if not already covered.

---

### Task 2: Create BaseLayout component

**Files:**
- Create: `email-templates/src/components/base-layout.tsx`

**Step 1: Create the shared layout component**

This component provides the consistent wrapper for all emails: DOCTYPE, head, body container, header with branding, footer.

Design tokens (from frontend `index.css`):
- Primary: `#10b981` (emerald)
- Background: `#ffffff`
- Card: `#f9fafb`
- Text: `#18181b`
- Muted text: `#71717a`
- Border: `#e4e4e7`
- Font: `-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif` (system stack — web fonts don't work in email)

```tsx
import {
  Html,
  Head,
  Body,
  Container,
  Section,
  Text,
  Hr,
  Preview,
} from '@react-email/components'
import * as React from 'react'

interface BaseLayoutProps {
  previewText: string
  children: React.ReactNode
}

export function BaseLayout({ previewText, children }: BaseLayoutProps) {
  return (
    <Html lang="pt-BR">
      <Head />
      <Preview>{previewText}</Preview>
      <Body style={body}>
        <Container style={container}>
          {/* Header */}
          <Section style={header}>
            <Text style={logo}>
              Scrap<span style={logoAccent}>Jobs</span>
            </Text>
          </Section>

          {/* Content */}
          <Section style={content}>
            {children}
          </Section>

          {/* Footer */}
          <Hr style={divider} />
          <Section style={footer}>
            <Text style={footerText}>
              ScrapJobs — Sua busca por vagas, automatizada.
            </Text>
            <Text style={footerText}>
              Este e-mail foi enviado automaticamente. Em caso de duvidas, responda a este e-mail.
            </Text>
          </Section>
        </Container>
      </Body>
    </Html>
  )
}

const body: React.CSSProperties = {
  backgroundColor: '#f4f4f5',
  fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
  margin: 0,
  padding: 0,
}

const container: React.CSSProperties = {
  maxWidth: '600px',
  margin: '0 auto',
  backgroundColor: '#ffffff',
  borderRadius: '8px',
  overflow: 'hidden',
  marginTop: '24px',
  marginBottom: '24px',
}

const header: React.CSSProperties = {
  backgroundColor: '#ffffff',
  borderBottom: '1px solid #e4e4e7',
  padding: '24px 32px',
}

const logo: React.CSSProperties = {
  fontSize: '24px',
  fontWeight: 700,
  color: '#18181b',
  margin: 0,
}

const logoAccent: React.CSSProperties = {
  color: '#10b981',
}

const content: React.CSSProperties = {
  padding: '32px',
}

const divider: React.CSSProperties = {
  borderColor: '#e4e4e7',
  margin: '0 32px',
}

const footer: React.CSSProperties = {
  padding: '24px 32px',
}

const footerText: React.CSSProperties = {
  fontSize: '12px',
  color: '#71717a',
  lineHeight: '20px',
  margin: '0 0 4px 0',
}
```

**Step 2: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add email-templates/src/components/base-layout.tsx
git commit -m "feat: add BaseLayout email component with ScrapJobs branding"
```

---

### Task 3: Create Welcome email template

**Files:**
- Create: `email-templates/src/emails/welcome.tsx`

**Step 1: Create the welcome email component**

Go placeholders used: `{{.UserName}}`, `{{.DashboardLink}}`

```tsx
import { Text, Button, Section } from '@react-email/components'
import { BaseLayout } from '../components/base-layout'
import * as React from 'react'

export default function WelcomeEmail() {
  return (
    <BaseLayout previewText="Bem-vindo(a) ao ScrapJobs!">
      <Text style={heading}>
        {'Bem-vindo(a), {{.UserName}}!'}
      </Text>

      <Text style={paragraph}>
        Sua conta foi criada com sucesso!
      </Text>

      <Text style={paragraph}>
        Agora voce pode comecar a automatizar sua busca por vagas e receber
        analises personalizadas diretamente no seu e-mail.
      </Text>

      <Text style={paragraph}>
        Acesse seu painel para configurar os sites que deseja monitorar e fazer
        upload do seu curriculo:
      </Text>

      <Section style={buttonContainer}>
        <Button style={button} href="{{.DashboardLink}}">
          Acessar meu Dashboard
        </Button>
      </Section>

      <Text style={paragraphMuted}>
        Se tiver alguma duvida, responda a este e-mail ou contate nosso suporte.
      </Text>
    </BaseLayout>
  )
}

const heading: React.CSSProperties = {
  fontSize: '22px',
  fontWeight: 700,
  color: '#18181b',
  margin: '0 0 16px 0',
}

const paragraph: React.CSSProperties = {
  fontSize: '15px',
  lineHeight: '24px',
  color: '#18181b',
  margin: '0 0 12px 0',
}

const paragraphMuted: React.CSSProperties = {
  fontSize: '14px',
  lineHeight: '22px',
  color: '#71717a',
  margin: '16px 0 0 0',
}

const buttonContainer: React.CSSProperties = {
  textAlign: 'center' as const,
  margin: '24px 0',
}

const button: React.CSSProperties = {
  backgroundColor: '#10b981',
  color: '#ffffff',
  padding: '12px 28px',
  borderRadius: '6px',
  fontSize: '15px',
  fontWeight: 600,
  textDecoration: 'none',
  display: 'inline-block',
}
```

**Step 2: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add email-templates/src/emails/welcome.tsx
git commit -m "feat: add Welcome email template"
```

---

### Task 4: Create Password Reset email template

**Files:**
- Create: `email-templates/src/emails/password-reset.tsx`

**Step 1: Create the password reset email component**

Go placeholders used: `{{.UserName}}`, `{{.ResetLink}}`

```tsx
import { Text, Button, Section } from '@react-email/components'
import { BaseLayout } from '../components/base-layout'
import * as React from 'react'

export default function PasswordResetEmail() {
  return (
    <BaseLayout previewText="Redefinicao de senha — ScrapJobs">
      <Text style={heading}>
        {'Ola {{.UserName}},'}
      </Text>

      <Text style={paragraph}>
        Recebemos uma solicitacao para redefinir sua senha.
      </Text>

      <Text style={paragraph}>
        Clique no botao abaixo para criar uma nova senha:
      </Text>

      <Section style={buttonContainer}>
        <Button style={button} href="{{.ResetLink}}">
          Redefinir Senha
        </Button>
      </Section>

      <Text style={paragraphMuted}>
        Este link e valido por 1 hora. Se voce nao solicitou esta redefinicao, ignore este e-mail.
      </Text>
    </BaseLayout>
  )
}

const heading: React.CSSProperties = {
  fontSize: '22px',
  fontWeight: 700,
  color: '#18181b',
  margin: '0 0 16px 0',
}

const paragraph: React.CSSProperties = {
  fontSize: '15px',
  lineHeight: '24px',
  color: '#18181b',
  margin: '0 0 12px 0',
}

const paragraphMuted: React.CSSProperties = {
  fontSize: '14px',
  lineHeight: '22px',
  color: '#71717a',
  margin: '16px 0 0 0',
}

const buttonContainer: React.CSSProperties = {
  textAlign: 'center' as const,
  margin: '24px 0',
}

const button: React.CSSProperties = {
  backgroundColor: '#10b981',
  color: '#ffffff',
  padding: '12px 28px',
  borderRadius: '6px',
  fontSize: '15px',
  fontWeight: 600,
  textDecoration: 'none',
  display: 'inline-block',
}
```

**Step 2: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add email-templates/src/emails/password-reset.tsx
git commit -m "feat: add Password Reset email template"
```

---

### Task 5: Create Job Analysis email template

**Files:**
- Create: `email-templates/src/emails/job-analysis.tsx`

**Step 1: Create the job analysis email component**

This is the most complex template. It uses Go template range directives for dynamic lists.

Go placeholders used:
- `{{.Job.Title}}`, `{{.Job.Company}}`
- `{{.Analysis.MatchAnalysis.OverallScoreNumeric}}`, `{{.Analysis.MatchAnalysis.OverallScoreQualitative}}`, `{{.Analysis.MatchAnalysis.Summary}}`
- `{{range .Analysis.StrengthsForThisJob}}` ... `{{.Point}}`, `{{.RelevanceToJob}}` ... `{{end}}`
- `{{range .Analysis.GapsAndImprovementAreas}}` ... `{{.AreaDescription}}`, `{{.JobRequirementImpacted}}` ... `{{end}}`
- `{{range .Analysis.ActionableResumeSuggestions}}` ... `{{.Suggestion}}`, `{{.CurriculumSectionToApply}}`, `{{.ExampleWording}}`, `{{.ReasoningForThisJob}}` ... `{{end}}`
- `{{.Analysis.FinalConsiderations}}`
- `{{.DashboardLink}}`

**IMPORTANT:** Go range/end directives must appear as literal text in the HTML output. React Email will render them as-is. The Go backend's `html/template` will interpret them at runtime.

```tsx
import { Text, Section, Hr, Button } from '@react-email/components'
import { BaseLayout } from '../components/base-layout'
import * as React from 'react'

export default function JobAnalysisEmail() {
  return (
    <BaseLayout previewText="Analise de Vaga — {{.Job.Title}}">
      <Text style={heading}>
        {'Analise de Vaga: {{.Job.Title}}'}
      </Text>

      <Text style={subheading}>
        {'{{.Job.Company}}'}
      </Text>

      <Text style={paragraph}>
        Segue abaixo a analise detalhada da compatibilidade do seu curriculo com a vaga encontrada.
      </Text>

      {/* Score Card */}
      <Section style={card}>
        <Text style={cardTitle}>Compatibilidade</Text>
        <Text style={scoreText}>
          {'{{.Analysis.MatchAnalysis.OverallScoreNumeric}}'}
          <span style={scoreLabel}>{'/100'}</span>
        </Text>
        <Text style={qualitativeText}>
          {'{{.Analysis.MatchAnalysis.OverallScoreQualitative}}'}
        </Text>
        <Text style={summaryText}>
          {'{{.Analysis.MatchAnalysis.Summary}}'}
        </Text>
      </Section>

      {/* Strengths */}
      <Section style={card}>
        <Text style={cardTitle}>Pontos Fortes para esta Vaga</Text>
        {/* Go range block — rendered as literal HTML */}
        <Text style={hiddenText}>{'{{range .Analysis.StrengthsForThisJob}}'}</Text>
        <Section style={listItem}>
          <Text style={listItemTitle}>{'{{.Point}}'}</Text>
          <Text style={listItemDetail}>{'Relevancia: {{.RelevanceToJob}}'}</Text>
        </Section>
        <Text style={hiddenText}>{'{{end}}'}</Text>
        <Text style={hiddenText}>{'{{if not .Analysis.StrengthsForThisJob}}'}</Text>
        <Text style={emptyText}>Nenhum ponto forte especifico identificado.</Text>
        <Text style={hiddenText}>{'{{end}}'}</Text>
      </Section>

      {/* Gaps */}
      <Section style={card}>
        <Text style={cardTitle}>Lacunas e Areas de Melhoria</Text>
        <Text style={hiddenText}>{'{{range .Analysis.GapsAndImprovementAreas}}'}</Text>
        <Section style={listItem}>
          <Text style={listItemTitle}>{'{{.AreaDescription}}'}</Text>
          <Text style={listItemDetail}>{'Impacto: {{.JobRequirementImpacted}}'}</Text>
        </Section>
        <Text style={hiddenText}>{'{{end}}'}</Text>
        <Text style={hiddenText}>{'{{if not .Analysis.GapsAndImprovementAreas}}'}</Text>
        <Text style={emptyText}>Nenhuma lacuna especifica identificada.</Text>
        <Text style={hiddenText}>{'{{end}}'}</Text>
      </Section>

      {/* Suggestions */}
      <Section style={card}>
        <Text style={cardTitle}>Sugestoes para o Curriculo</Text>
        <Text style={hiddenText}>{'{{range .Analysis.ActionableResumeSuggestions}}'}</Text>
        <Section style={listItem}>
          <Text style={listItemTitle}>{'{{.Suggestion}}'}</Text>
          <Text style={listItemDetail}>{'Secao: {{.CurriculumSectionToApply}}'}</Text>
          <Text style={listItemDetail}>{'Exemplo: "{{.ExampleWording}}"'}</Text>
          <Text style={listItemDetail}>{'Justificativa: {{.ReasoningForThisJob}}'}</Text>
        </Section>
        <Text style={hiddenText}>{'{{end}}'}</Text>
        <Text style={hiddenText}>{'{{if not .Analysis.ActionableResumeSuggestions}}'}</Text>
        <Text style={emptyText}>Nenhuma sugestao especifica identificada.</Text>
        <Text style={hiddenText}>{'{{end}}'}</Text>
      </Section>

      {/* Final Considerations */}
      <Hr style={divider} />
      <Text style={cardTitle}>Consideracoes Finais</Text>
      <Text style={paragraph}>
        {'{{.Analysis.FinalConsiderations}}'}
      </Text>

      <Section style={buttonContainer}>
        <Button style={button} href="{{.DashboardLink}}">
          Ver no Dashboard
        </Button>
      </Section>
    </BaseLayout>
  )
}

const heading: React.CSSProperties = {
  fontSize: '22px',
  fontWeight: 700,
  color: '#18181b',
  margin: '0 0 4px 0',
}

const subheading: React.CSSProperties = {
  fontSize: '15px',
  color: '#71717a',
  margin: '0 0 16px 0',
}

const paragraph: React.CSSProperties = {
  fontSize: '15px',
  lineHeight: '24px',
  color: '#18181b',
  margin: '0 0 16px 0',
}

const card: React.CSSProperties = {
  backgroundColor: '#f9fafb',
  borderRadius: '8px',
  border: '1px solid #e4e4e7',
  padding: '20px',
  marginBottom: '16px',
}

const cardTitle: React.CSSProperties = {
  fontSize: '16px',
  fontWeight: 700,
  color: '#18181b',
  margin: '0 0 12px 0',
}

const scoreText: React.CSSProperties = {
  fontSize: '36px',
  fontWeight: 700,
  color: '#10b981',
  margin: '0 0 4px 0',
}

const scoreLabel: React.CSSProperties = {
  fontSize: '16px',
  fontWeight: 400,
  color: '#71717a',
}

const qualitativeText: React.CSSProperties = {
  fontSize: '14px',
  fontWeight: 600,
  color: '#10b981',
  margin: '0 0 8px 0',
}

const summaryText: React.CSSProperties = {
  fontSize: '14px',
  lineHeight: '22px',
  color: '#71717a',
  margin: 0,
}

const listItem: React.CSSProperties = {
  padding: '8px 0',
  borderBottom: '1px solid #e4e4e7',
}

const listItemTitle: React.CSSProperties = {
  fontSize: '14px',
  fontWeight: 600,
  color: '#18181b',
  margin: '0 0 4px 0',
}

const listItemDetail: React.CSSProperties = {
  fontSize: '13px',
  color: '#71717a',
  margin: '0 0 2px 0',
  lineHeight: '20px',
}

const emptyText: React.CSSProperties = {
  fontSize: '14px',
  color: '#71717a',
  fontStyle: 'italic',
  margin: 0,
}

const hiddenText: React.CSSProperties = {
  display: 'none',
  fontSize: 0,
  lineHeight: 0,
  maxHeight: 0,
  overflow: 'hidden',
}

const divider: React.CSSProperties = {
  borderColor: '#e4e4e7',
  margin: '24px 0',
}

const buttonContainer: React.CSSProperties = {
  textAlign: 'center' as const,
  margin: '24px 0',
}

const button: React.CSSProperties = {
  backgroundColor: '#10b981',
  color: '#ffffff',
  padding: '12px 28px',
  borderRadius: '6px',
  fontSize: '15px',
  fontWeight: 600,
  textDecoration: 'none',
  display: 'inline-block',
}
```

**IMPORTANT IMPLEMENTATION NOTE:** The `hiddenText` style hides the Go template directives (`{{range}}`, `{{end}}`, `{{if}}`) from visual rendering while keeping them in the HTML source. The Go `html/template` engine processes them from the raw HTML regardless of CSS. If `display:none` causes issues with some email clients stripping content, an alternative approach is to wrap them in HTML comments — but Go's template engine processes before the HTML is rendered, so `display:none` works here.

**Step 2: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add email-templates/src/emails/job-analysis.tsx
git commit -m "feat: add Job Analysis email template"
```

---

### Task 6: Create New Jobs Alert email template

**Files:**
- Create: `email-templates/src/emails/new-jobs-alert.tsx`

**Step 1: Create the new jobs alert email component**

Go placeholders used:
- `{{.UserName}}`
- `{{len .Jobs}}` (for job count)
- `{{range .Jobs}}` ... `{{.Title}}`, `{{.Company}}`, `{{.Location}}`, `{{.JobLink}}` ... `{{end}}`
- `{{.DashboardLink}}`

```tsx
import { Text, Section, Link, Hr, Button } from '@react-email/components'
import { BaseLayout } from '../components/base-layout'
import * as React from 'react'

export default function NewJobsAlertEmail() {
  return (
    <BaseLayout previewText="Novas vagas encontradas para voce!">
      <Text style={heading}>
        {'Novas vagas encontradas, {{.UserName}}!'}
      </Text>

      <Text style={paragraph}>
        {'Encontramos {{len .Jobs}} nova(s) vaga(s) nos sites que voce esta monitorando:'}
      </Text>

      {/* Jobs list */}
      <Text style={hiddenText}>{'{{range .Jobs}}'}</Text>
      <Section style={jobCard}>
        <Text style={jobTitle}>{'{{.Title}}'}</Text>
        <Text style={jobDetail}>{'{{.Company}} — {{.Location}}'}</Text>
        <Link href="{{.JobLink}}" style={jobLink}>
          Ver vaga →
        </Link>
      </Section>
      <Text style={hiddenText}>{'{{end}}'}</Text>

      <Hr style={divider} />

      <Text style={paragraph}>
        Acesse seu painel no ScrapJobs para analisar essas vagas com IA e receber
        sugestoes personalizadas para o seu curriculo.
      </Text>

      <Section style={buttonContainer}>
        <Button style={button} href="{{.DashboardLink}}">
          Acessar Dashboard
        </Button>
      </Section>
    </BaseLayout>
  )
}

const heading: React.CSSProperties = {
  fontSize: '22px',
  fontWeight: 700,
  color: '#18181b',
  margin: '0 0 16px 0',
}

const paragraph: React.CSSProperties = {
  fontSize: '15px',
  lineHeight: '24px',
  color: '#18181b',
  margin: '0 0 16px 0',
}

const jobCard: React.CSSProperties = {
  backgroundColor: '#f9fafb',
  borderRadius: '8px',
  border: '1px solid #e4e4e7',
  padding: '16px 20px',
  marginBottom: '8px',
}

const jobTitle: React.CSSProperties = {
  fontSize: '15px',
  fontWeight: 600,
  color: '#18181b',
  margin: '0 0 4px 0',
}

const jobDetail: React.CSSProperties = {
  fontSize: '13px',
  color: '#71717a',
  margin: '0 0 8px 0',
}

const jobLink: React.CSSProperties = {
  fontSize: '13px',
  color: '#10b981',
  fontWeight: 600,
  textDecoration: 'none',
}

const hiddenText: React.CSSProperties = {
  display: 'none',
  fontSize: 0,
  lineHeight: 0,
  maxHeight: 0,
  overflow: 'hidden',
}

const divider: React.CSSProperties = {
  borderColor: '#e4e4e7',
  margin: '24px 0',
}

const buttonContainer: React.CSSProperties = {
  textAlign: 'center' as const,
  margin: '24px 0',
}

const button: React.CSSProperties = {
  backgroundColor: '#10b981',
  color: '#ffffff',
  padding: '12px 28px',
  borderRadius: '6px',
  fontSize: '15px',
  fontWeight: 600,
  textDecoration: 'none',
  display: 'inline-block',
}
```

**Step 2: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add email-templates/src/emails/new-jobs-alert.tsx
git commit -m "feat: add New Jobs Alert email template"
```

---

### Task 7: Create build script

**Files:**
- Create: `email-templates/scripts/build.ts`

**Step 1: Create the build script that renders all emails to HTML**

```typescript
import { render } from '@react-email/render'
import { writeFileSync, mkdirSync } from 'fs'
import { join } from 'path'

import WelcomeEmail from '../src/emails/welcome'
import PasswordResetEmail from '../src/emails/password-reset'
import JobAnalysisEmail from '../src/emails/job-analysis'
import NewJobsAlertEmail from '../src/emails/new-jobs-alert'

const DIST_DIR = join(__dirname, '..', 'dist')
const GO_TEMPLATES_DIR = join(__dirname, '..', '..', 'templates', 'emails')

const emails = [
  { name: 'welcome', component: WelcomeEmail },
  { name: 'password-reset', component: PasswordResetEmail },
  { name: 'job-analysis', component: JobAnalysisEmail },
  { name: 'new-jobs-alert', component: NewJobsAlertEmail },
] as const

async function build() {
  mkdirSync(DIST_DIR, { recursive: true })
  mkdirSync(GO_TEMPLATES_DIR, { recursive: true })

  for (const { name, component } of emails) {
    const html = await render(component())

    // Save to dist/ (source of truth)
    const distPath = join(DIST_DIR, `${name}.html`)
    writeFileSync(distPath, html, 'utf-8')

    // Copy to Go templates dir (for embed)
    const goPath = join(GO_TEMPLATES_DIR, `${name}.html`)
    writeFileSync(goPath, html, 'utf-8')

    console.log(`  Built: ${name}.html`)
  }

  console.log(`\nDone! ${emails.length} templates built.`)
}

build().catch((err) => {
  console.error('Build failed:', err)
  process.exit(1)
})
```

**Step 2: Run the build**

Run: `cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs/email-templates && npm run build`
Expected: Output shows 4 templates built, `dist/` and `../templates/emails/` contain HTML files

**Step 3: Verify output contains Go placeholders**

Run: `grep '{{.UserName}}' /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs/templates/emails/welcome.html`
Expected: At least one match — confirms Go template placeholders survived rendering

**Step 4: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add email-templates/scripts/build.ts email-templates/dist/ templates/emails/
git commit -m "feat: add email build script, generate HTML templates"
```

---

### Task 8: Create Go embed package

**Files:**
- Create: `templates/embed.go`

**Step 1: Create the embed package**

```go
package templates

import "embed"

//go:embed emails/*.html
var EmailTemplates embed.FS
```

**Step 2: Verify it compiles**

Run: `cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go build ./templates/...`
Expected: No errors

**Step 3: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add templates/embed.go
git commit -m "feat: add Go embed package for email templates"
```

---

### Task 9: Refactor emailAdapter.go to use embedded templates

**Files:**
- Modify: `usecase/emailAdapter.go`

This is the core change. Replace inline HTML templates with embedded template loading. Keep all plain text generators unchanged. Keep the `SESSenderAdapter` struct and public methods unchanged.

**Step 1: Rewrite the HTML generation functions**

Replace the entire file content. The key changes:
1. Import `web-scrapper/templates` package
2. Remove all inline HTML template strings
3. Load templates from embedded FS using `template.ParseFS`
4. Keep all `generateXxxEmailBodyText()` functions exactly as they are
5. Keep `SESSenderAdapter` struct, constructor, and public methods exactly as they are

New `emailAdapter.go`:

```go
package usecase

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"
	"web-scrapper/interfaces"
	"web-scrapper/model"
	"web-scrapper/templates"
)

var emailTemplates *template.Template

func init() {
	var err error
	emailTemplates, err = template.ParseFS(templates.EmailTemplates, "emails/*.html")
	if err != nil {
		panic(fmt.Sprintf("failed to parse email templates: %v", err))
	}
}

func generateWelcomeEmailBodyHTML(userName, dashboardLink string) (string, error) {
	data := struct {
		UserName      string
		DashboardLink string
	}{UserName: userName, DashboardLink: dashboardLink}

	var body bytes.Buffer
	if err := emailTemplates.ExecuteTemplate(&body, "welcome.html", data); err != nil {
		return "", err
	}
	return body.String(), nil
}

func generateWelcomeEmailBodyText(userName, dashboardLink string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Bem-vindo(a) ao ScrapJobs, %s!\n\n", userName))
	sb.WriteString("Sua conta foi criada com sucesso!\n\n")
	sb.WriteString("Agora você pode começar a automatizar sua busca por vagas e receber análises personalizadas diretamente no seu e-mail.\n\n")
	sb.WriteString("Acesse seu painel para configurar os sites que deseja monitorar e fazer upload do seu currículo:\n")
	sb.WriteString(dashboardLink + "\n\n")
	sb.WriteString("Se tiver alguma dúvida, responda a este e-mail ou contate nosso suporte.\n\n")
	sb.WriteString("Atenciosamente,\nEquipe ScrapJobs\n")
	return sb.String()
}

func generateEmailBodyHTML(analysis model.ResumeAnalysis, job model.Job) (string, error) {
	data := struct {
		Analysis      model.ResumeAnalysis
		Job           model.Job
		DashboardLink string
	}{
		Analysis:      analysis,
		Job:           job,
		DashboardLink: "", // Not used in current flow, but available for future
	}

	var body bytes.Buffer
	if err := emailTemplates.ExecuteTemplate(&body, "job-analysis.html", data); err != nil {
		return "", fmt.Errorf("ERROR to execute email template: %w", err)
	}
	return body.String(), nil
}

func generateEmailBodyText(analysis model.ResumeAnalysis, job model.Job) string {
	var sb strings.Builder

	sb.WriteString("Prezados(as),\n\n")
	sb.WriteString(fmt.Sprintf("Segue abaixo a análise detalhada do currículo para a posição de %s:\n\n", job.Title))

	sb.WriteString("**Análise de Compatibilidade:**\n")
	sb.WriteString(fmt.Sprintf("*   **Pontuação Geral:** %d\n", analysis.MatchAnalysis.OverallScoreNumeric))
	sb.WriteString(fmt.Sprintf("*   **Avaliação Qualitativa:** %s\n", analysis.MatchAnalysis.OverallScoreQualitative))
	sb.WriteString(fmt.Sprintf("*   **Resumo:** %s\n", analysis.MatchAnalysis.Summary))
	sb.WriteString("\n---\n\n")

	sb.WriteString("**Pontos Fortes para esta Vaga:**\n\n")
	if len(analysis.StrengthsForThisJob) > 0 {
		for i, strength := range analysis.StrengthsForThisJob {
			sb.WriteString(fmt.Sprintf("%d.  **Ponto:** %s\n", i+1, strength.Point))
			sb.WriteString(fmt.Sprintf("    *   **Relevância:** %s\n\n", strength.RelevanceToJob))
		}
	} else {
		sb.WriteString("Nenhum ponto forte específico identificado.\n\n")
	}
	sb.WriteString("---\n\n")

	sb.WriteString("**Lacunas e Áreas de Melhoria:**\n\n")
	if len(analysis.GapsAndImprovementAreas) > 0 {
		for i, gap := range analysis.GapsAndImprovementAreas {
			sb.WriteString(fmt.Sprintf("%d.  **Área:** %s\n", i+1, gap.AreaDescription))
			sb.WriteString(fmt.Sprintf("    *   **Impacto no Requisito da Vaga:** %s\n\n", gap.JobRequirementImpacted))
		}
	} else {
		sb.WriteString("Nenhuma lacuna ou área de melhoria específica identificada.\n\n")
	}
	sb.WriteString("---\n\n")

	sb.WriteString("**Sugestões Acionáveis para o Currículo:**\n\n")
	if len(analysis.ActionableResumeSuggestions) > 0 {
		for i, suggestion := range analysis.ActionableResumeSuggestions {
			sb.WriteString(fmt.Sprintf("%d.  **Sugestão:** %s\n", i+1, suggestion.Suggestion))
			sb.WriteString(fmt.Sprintf("    *   **Seção do Currículo:** %s\n", suggestion.CurriculumSectionToApply))
			sb.WriteString(fmt.Sprintf("    *   **Exemplo de Redação:** %s\n", suggestion.ExampleWording))
			sb.WriteString(fmt.Sprintf("    *   **Justificativa:** %s\n\n", suggestion.ReasoningForThisJob))
		}
	} else {
		sb.WriteString("Nenhuma sugestão acionável para o currículo identificada.\n\n")
	}
	sb.WriteString("---\n\n")

	sb.WriteString("**Considerações Finais:**\n")
	sb.WriteString(analysis.FinalConsiderations)
	sb.WriteString("\n\n---\n")

	sb.WriteString("Atenciosamente,\n\n")
	sb.WriteString("Equipe ScrapJobs\n")

	return sb.String()
}

type SESSenderAdapter struct {
	mailSender interfaces.MailSender
}

func NewSESSenderAdapter(mailSender interfaces.MailSender) *SESSenderAdapter {
	return &SESSenderAdapter{
		mailSender: mailSender,
	}
}

func (adapter *SESSenderAdapter) SendAnalysisEmail(ctx context.Context, userEmail string, job model.Job, analysis model.ResumeAnalysis) error {
	subject := fmt.Sprintf("Análise de Vaga Encontrada: %s", job.Title)

	bodyHtml, err := generateEmailBodyHTML(analysis, job)
	if err != nil {
		return fmt.Errorf("ERROR to generate html body: %w", err)
	}

	bodyText := generateEmailBodyText(analysis, job)

	return adapter.mailSender.SendEmail(ctx, userEmail, subject, bodyText, bodyHtml)
}

func generateNewJobsEmailBodyHTML(userName string, jobs []*model.Job) (string, error) {
	data := struct {
		UserName      string
		Jobs          []*model.Job
		DashboardLink string
	}{UserName: userName, Jobs: jobs, DashboardLink: ""}

	var body bytes.Buffer
	if err := emailTemplates.ExecuteTemplate(&body, "new-jobs-alert.html", data); err != nil {
		return "", err
	}
	return body.String(), nil
}

func generateNewJobsEmailBodyText(userName string, jobs []*model.Job) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Novas vagas encontradas para você, %s!\n\n", userName))
	sb.WriteString(fmt.Sprintf("Encontramos %d nova(s) vaga(s) nos sites que você está monitorando:\n\n", len(jobs)))
	for i, job := range jobs {
		sb.WriteString(fmt.Sprintf("%d. %s — %s (%s)\n   Link: %s\n\n", i+1, job.Title, job.Company, job.Location, job.JobLink))
	}
	sb.WriteString("Acesse seu painel no ScrapJobs para analisar essas vagas com IA.\n\n")
	sb.WriteString("Atenciosamente,\nEquipe ScrapJobs\n")
	return sb.String()
}

func (adapter *SESSenderAdapter) SendNewJobsEmail(ctx context.Context, userEmail string, userName string, jobs []*model.Job) error {
	subject := fmt.Sprintf("ScrapJobs: %d nova(s) vaga(s) encontrada(s)!", len(jobs))

	bodyHtml, err := generateNewJobsEmailBodyHTML(userName, jobs)
	if err != nil {
		return fmt.Errorf("erro ao gerar corpo HTML do email de novas vagas: %w", err)
	}

	bodyText := generateNewJobsEmailBodyText(userName, jobs)

	return adapter.mailSender.SendEmail(ctx, userEmail, subject, bodyText, bodyHtml)
}

func (adapter *SESSenderAdapter) SendWelcomeEmail(ctx context.Context, userEmail, userName, dashboardLink string) error {
	subject := "Bem-vindo(a) ao ScrapJobs!"

	bodyHtml, err := generateWelcomeEmailBodyHTML(userName, dashboardLink)
	if err != nil {
		return fmt.Errorf("erro ao gerar corpo HTML do email de boas-vindas: %w", err)
	}

	bodyText := generateWelcomeEmailBodyText(userName, dashboardLink)

	return adapter.mailSender.SendEmail(ctx, userEmail, subject, bodyText, bodyHtml)
}

func generatePasswordResetEmailHTML(userName, resetLink string) (string, error) {
	data := struct {
		UserName  string
		ResetLink string
	}{UserName: userName, ResetLink: resetLink}

	var body bytes.Buffer
	if err := emailTemplates.ExecuteTemplate(&body, "password-reset.html", data); err != nil {
		return "", err
	}
	return body.String(), nil
}

func (adapter *SESSenderAdapter) SendPasswordResetEmail(ctx context.Context, email, userName, resetLink string) error {
	subject := "ScrapJobs — Redefinição de Senha"

	bodyHTML, err := generatePasswordResetEmailHTML(userName, resetLink)
	if err != nil {
		return fmt.Errorf("erro ao gerar corpo HTML do email de redefinição de senha: %w", err)
	}

	bodyText := fmt.Sprintf("Olá %s, clique no link para redefinir sua senha: %s (válido por 1 hora)", userName, resetLink)

	return adapter.mailSender.SendEmail(ctx, email, subject, bodyText, bodyHTML)
}
```

**Step 2: Verify it compiles**

Run: `cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add usecase/emailAdapter.go
git commit -m "refactor: use embedded React Email templates instead of inline HTML"
```

---

### Task 10: Fix and run existing tests

**Files:**
- Modify: `usecase/emailAdapter_test.go` (if needed)

The existing tests call `generateEmailBodyHTML`, `generateWelcomeEmailBodyHTML`, etc. These functions now use embedded templates instead of inline strings. The tests should still pass if the templates contain the same data — the output HTML will be different (React Email styled) but the dynamic data should still be present.

**Step 1: Run existing tests**

Run: `cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go test ./usecase/... -run TestGenerate -v`
Expected: All tests pass. The `assert.Contains` checks look for data values (names, scores, etc.) which should still be present in the new templates.

**Step 2: If any test fails, investigate**

Possible issues:
- The `{{range}}` / `{{if}}` / `{{end}}` Go directives might be inside `<p style="display:none">` tags. The `html/template` engine processes directives regardless of CSS, so `{{range .Analysis.StrengthsForThisJob}}` inside a hidden `<p>` tag should still work — the engine sees it as a template action, not HTML content.
- If `html/template` escapes the `{{` inside the hidden `<p>` tag output from React Email, we'll need to adjust the approach. In that case, we'd switch to `text/template` for these email templates (since the content is pre-built HTML, not user input).

**IMPORTANT:** If `html/template` fails to parse because the Go directives are inside HTML attribute values or are being double-escaped, switch the import in `emailAdapter.go` from `html/template` to `text/template`. This is safe because the template HTML is pre-built by React Email (trusted), and the dynamic data (names, URLs) is not user-controlled HTML.

**Step 3: Commit if changes were needed**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add usecase/emailAdapter_test.go
git commit -m "test: update email adapter tests for embedded templates"
```

---

### Task 11: Add .gitignore entries

**Files:**
- Modify: `.gitignore` (root of ScrapJobs)

**Step 1: Add node_modules to gitignore**

Add these lines:

```
# Email templates
email-templates/node_modules/
```

**Step 2: Commit**

```bash
cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs
git add .gitignore
git commit -m "chore: add email-templates/node_modules to gitignore"
```

---

### Task 12: Final verification

**Step 1: Full Go build**

Run: `cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go build ./...`
Expected: exit 0

**Step 2: Full Go test suite**

Run: `cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go test ./usecase/... ./controller/... ./middleware/... -v`
Expected: All tests pass

**Step 3: Verify email templates render correctly with sample data**

Run: `cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs && go test ./usecase/... -run TestGenerate -v`
Expected: All template generation tests pass, confirming Go placeholders work with the React Email output

**Step 4: Visual check (optional)**

Run: `cd /Users/erickschaedler/Documents/Scrap\ Jobs/ScrapJobs/email-templates && npx email dev`
Open browser to preview templates in React Email dev server. Note: Go placeholders will show as literal text (e.g., `{{.UserName}}`), which is expected.
