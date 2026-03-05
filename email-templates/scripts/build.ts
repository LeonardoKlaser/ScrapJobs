import { render } from '@react-email/render'
import { writeFileSync, mkdirSync } from 'fs'
import { join, dirname } from 'path'
import { fileURLToPath } from 'url'

import WelcomeEmail from '../src/emails/welcome'
import PasswordResetEmail from '../src/emails/password-reset'
import JobAnalysisEmail from '../src/emails/job-analysis'
import NewJobsAlertEmail from '../src/emails/new-jobs-alert'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

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

    const distPath = join(DIST_DIR, `${name}.html`)
    writeFileSync(distPath, html, 'utf-8')

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
