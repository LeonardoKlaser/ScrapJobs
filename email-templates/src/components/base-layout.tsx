import {
  Html, Head, Body, Container, Section, Text, Hr, Preview,
} from '@react-email/components'
import * as React from 'react'

interface BaseLayoutProps {
  previewText: string
  children: React.ReactNode
}

export function BaseLayout({ previewText, children }: BaseLayoutProps) {
  return (
    <Html lang="pt-BR" dir="ltr">
      <Head />
      <Preview>{previewText}</Preview>
      <Body style={body}>
        <Container style={container}>
          <Section style={header}>
            <Text style={logo}>
              Scrap<span style={logoAccent}>Jobs</span>
            </Text>
          </Section>
          <Section style={content}>{children}</Section>
          <Hr style={divider} />
          <Section style={footer}>
            <Text style={footerBrand}>
              Scrap<span style={logoAccent}>Jobs</span>
              <span style={footerTagline}> — Sua busca por vagas, automatizada.</span>
            </Text>
            <Text style={footerText}>
              Este e-mail foi enviado automaticamente. Em caso de dúvidas, responda a este e-mail.
            </Text>
          </Section>
        </Container>
      </Body>
    </Html>
  )
}

const body: React.CSSProperties = {
  backgroundColor: '#09090b',
  fontFamily: "Inter, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
  margin: 0,
  padding: 0,
}
const container: React.CSSProperties = {
  maxWidth: '600px',
  margin: '0 auto',
  backgroundColor: '#18181b',
  borderRadius: '12px',
  overflow: 'hidden',
  marginTop: '32px',
  marginBottom: '32px',
  borderTop: '4px solid #10b981',
}
const header: React.CSSProperties = {
  padding: '28px 32px 0',
}
const logo: React.CSSProperties = {
  fontSize: '28px',
  fontWeight: 700,
  color: '#fafafa',
  margin: 0,
  letterSpacing: '-0.02em',
}
const logoAccent: React.CSSProperties = { color: '#10b981' }
const content: React.CSSProperties = { padding: '24px 32px 32px' }
const divider: React.CSSProperties = { borderColor: '#27272a', margin: '0 32px' }
const footer: React.CSSProperties = { padding: '24px 32px' }
const footerBrand: React.CSSProperties = {
  fontSize: '13px',
  fontWeight: 600,
  color: '#52525b',
  lineHeight: '20px',
  margin: '0 0 4px 0',
}
const footerTagline: React.CSSProperties = { fontWeight: 400 }
const footerText: React.CSSProperties = {
  fontSize: '12px',
  color: '#52525b',
  lineHeight: '20px',
  margin: 0,
}
