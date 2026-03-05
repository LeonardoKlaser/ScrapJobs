import { Text, Button, Section } from '@react-email/components'
import { BaseLayout } from '../components/base-layout'
import * as React from 'react'

export default function WelcomeEmail() {
  return (
    <BaseLayout previewText="Bem-vindo(a) ao ScrapJobs!">
      <Text style={heading}>{'Bem-vindo(a), {{.UserName}}!'}</Text>
      <Text style={paragraph}>Sua conta foi criada com sucesso!</Text>
      <Text style={paragraph}>Agora você pode começar a automatizar sua busca por vagas e receber análises personalizadas diretamente no seu e-mail.</Text>
      <Text style={paragraph}>Acesse seu painel para configurar os sites que deseja monitorar e fazer upload do seu currículo:</Text>
      <Section style={buttonContainer}>
        <Button style={button} href="{{.DashboardLink}}">Acessar meu Dashboard</Button>
      </Section>
      <Text style={paragraphMuted}>Se tiver alguma dúvida, responda a este e-mail ou contate nosso suporte.</Text>
    </BaseLayout>
  )
}

const heading: React.CSSProperties = { fontSize: '22px', fontWeight: 700, color: '#18181b', margin: '0 0 16px 0' }
const paragraph: React.CSSProperties = { fontSize: '15px', lineHeight: '24px', color: '#18181b', margin: '0 0 12px 0' }
const paragraphMuted: React.CSSProperties = { fontSize: '14px', lineHeight: '22px', color: '#71717a', margin: '16px 0 0 0' }
const buttonContainer: React.CSSProperties = { textAlign: 'center' as const, margin: '24px 0' }
const button: React.CSSProperties = {
  backgroundColor: '#10b981', color: '#ffffff', padding: '12px 28px',
  borderRadius: '6px', fontSize: '15px', fontWeight: 600, textDecoration: 'none', display: 'inline-block',
}
