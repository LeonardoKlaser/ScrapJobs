import { Text, Section, Link, Hr, Button } from '@react-email/components'
import { BaseLayout } from '../components/base-layout'
import * as React from 'react'

export default function NewJobsAlertEmail() {
  return (
    <BaseLayout previewText="Novas vagas encontradas para você!">
      <Text style={heading}>{'Novas vagas encontradas, {{.UserName}}!'}</Text>
      <Text style={paragraph}>{'Encontramos {{len .Jobs}} nova(s) vaga(s) nos sites que você está monitorando:'}</Text>

      <Text style={hidden}>{'{{range .Jobs}}'}</Text>
      <Section style={jobCard}>
        <Text style={jobTitle}>{'{{.Title}}'}</Text>
        <Text style={jobDetail}>{'{{.Company}} — {{.Location}}'}</Text>
        <Link href="{{.JobLink}}" style={jobLink}>Ver vaga →</Link>
      </Section>
      <Text style={hidden}>{'{{end}}'}</Text>

      <Hr style={dividerStyle} />
      <Text style={paragraph}>Acesse seu painel no ScrapJobs para analisar essas vagas com IA e receber sugestões personalizadas para o seu currículo.</Text>
      <Section style={buttonContainer}>
        <Button style={button} href="{{.DashboardLink}}">Acessar Dashboard</Button>
      </Section>
    </BaseLayout>
  )
}

const heading: React.CSSProperties = { fontSize: '22px', fontWeight: 700, color: '#18181b', margin: '0 0 16px 0' }
const paragraph: React.CSSProperties = { fontSize: '15px', lineHeight: '24px', color: '#18181b', margin: '0 0 16px 0' }
const jobCard: React.CSSProperties = {
  backgroundColor: '#f9fafb', borderRadius: '8px', border: '1px solid #e4e4e7', padding: '16px 20px', marginBottom: '8px',
}
const jobTitle: React.CSSProperties = { fontSize: '15px', fontWeight: 600, color: '#18181b', margin: '0 0 4px 0' }
const jobDetail: React.CSSProperties = { fontSize: '13px', color: '#71717a', margin: '0 0 8px 0' }
const jobLink: React.CSSProperties = { fontSize: '13px', color: '#10b981', fontWeight: 600, textDecoration: 'none' }
const hidden: React.CSSProperties = { display: 'none', fontSize: 0, lineHeight: 0, maxHeight: 0, overflow: 'hidden' }
const dividerStyle: React.CSSProperties = { borderColor: '#e4e4e7', margin: '24px 0' }
const buttonContainer: React.CSSProperties = { textAlign: 'center' as const, margin: '24px 0' }
const button: React.CSSProperties = {
  backgroundColor: '#10b981', color: '#ffffff', padding: '12px 28px',
  borderRadius: '6px', fontSize: '15px', fontWeight: 600, textDecoration: 'none', display: 'inline-block',
}
