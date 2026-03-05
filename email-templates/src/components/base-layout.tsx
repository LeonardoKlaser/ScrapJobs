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
    <Html lang="pt-BR">
      <Head />
      <Preview>{previewText}</Preview>
      <Body style={body}>
        <Container style={container}>
          <Section style={header}>
            <Text style={logo}>Scrap<span style={logoAccent}>Jobs</span></Text>
          </Section>
          <Section style={content}>{children}</Section>
          <Hr style={divider} />
          <Section style={footer}>
            <Text style={footerText}>ScrapJobs — Sua busca por vagas, automatizada.</Text>
            <Text style={footerText}>Este e-mail foi enviado automaticamente. Em caso de dúvidas, responda a este e-mail.</Text>
          </Section>
        </Container>
      </Body>
    </Html>
  )
}

const body: React.CSSProperties = {
  backgroundColor: '#f4f4f5',
  fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
  margin: 0, padding: 0,
}
const container: React.CSSProperties = {
  maxWidth: '600px', margin: '0 auto', backgroundColor: '#ffffff',
  borderRadius: '8px', overflow: 'hidden', marginTop: '24px', marginBottom: '24px',
}
const header: React.CSSProperties = {
  backgroundColor: '#ffffff', borderBottom: '1px solid #e4e4e7', padding: '24px 32px',
}
const logo: React.CSSProperties = { fontSize: '24px', fontWeight: 700, color: '#18181b', margin: 0 }
const logoAccent: React.CSSProperties = { color: '#10b981' }
const content: React.CSSProperties = { padding: '32px' }
const divider: React.CSSProperties = { borderColor: '#e4e4e7', margin: '0 32px' }
const footer: React.CSSProperties = { padding: '24px 32px' }
const footerText: React.CSSProperties = { fontSize: '12px', color: '#71717a', lineHeight: '20px', margin: '0 0 4px 0' }
