import { Text, Button, Section } from '@react-email/components'
import { BaseLayout } from '../components/base-layout'
import * as React from 'react'

export default function PasswordResetEmail() {
  return (
    <BaseLayout previewText="Redefinição de senha — ScrapJobs">
      <Text style={heading}>{'Olá {{.UserName}},'}</Text>
      <Text style={paragraph}>Recebemos uma solicitação para redefinir sua senha.</Text>
      <Text style={paragraph}>Clique no botão abaixo para criar uma nova senha:</Text>
      <Section style={buttonContainer}>
        <Button style={button} href="{{.ResetLink}}">Redefinir Senha</Button>
      </Section>
      <Text style={mutedText}>
        Este link é válido por 1 hora. Se você não solicitou esta redefinição, ignore este e-mail.
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
