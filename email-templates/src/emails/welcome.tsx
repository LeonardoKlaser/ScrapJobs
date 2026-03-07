import { Text, Button, Section } from '@react-email/components'
import { BaseLayout } from '../components/base-layout'
import * as React from 'react'

export default function WelcomeEmail() {
  return (
    <BaseLayout previewText="Bem-vindo(a) ao ScrapJobs!">
      <Text style={heading}>{'Bem-vindo(a), {{.UserName}}!'}</Text>
      <Text style={paragraph}>Sua conta foi criada com sucesso!</Text>
      <Text style={paragraph}>
        Agora você pode começar a automatizar sua busca por vagas e receber
        análises personalizadas diretamente no seu e-mail.
      </Text>
      <Text style={paragraph}>
        Acesse seu painel para configurar os sites que deseja monitorar e
        fazer upload do seu currículo:
      </Text>
      <Section style={buttonContainer}>
        <Button style={button} href="{{.DashboardLink}}">
          Acessar meu Dashboard
        </Button>
      </Section>
      <Text style={mutedText}>
        Se tiver alguma dúvida, responda a este e-mail ou contate nosso suporte.
      </Text>
    </BaseLayout>
  )
}

const heading: React.CSSProperties = {
  fontSize: '24px',
  fontWeight: 700,
  color: '#fafafa',
  margin: '0 0 20px 0',
  letterSpacing: '-0.02em',
}
const paragraph: React.CSSProperties = {
  fontSize: '15px',
  lineHeight: '26px',
  color: '#d4d4d8',
  margin: '0 0 12px 0',
}
const mutedText: React.CSSProperties = {
  fontSize: '14px',
  lineHeight: '22px',
  color: '#71717a',
  margin: '16px 0 0 0',
}
const buttonContainer: React.CSSProperties = {
  textAlign: 'center' as const,
  margin: '28px 0',
}
const button: React.CSSProperties = {
  backgroundColor: '#10b981',
  color: '#ffffff',
  padding: '14px 32px',
  borderRadius: '8px',
  fontSize: '15px',
  fontWeight: 600,
  textDecoration: 'none',
  display: 'inline-block',
}
