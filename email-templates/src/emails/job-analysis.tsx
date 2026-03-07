import { Text, Section, Hr, Button } from '@react-email/components'
import { BaseLayout } from '../components/base-layout'
import * as React from 'react'

export default function JobAnalysisEmail() {
  return (
    <BaseLayout previewText="Análise de Vaga — {{.Job.Title}}">
      <Text style={heading}>{'Análise de Vaga: {{.Job.Title}}'}</Text>
      <Text style={subheading}>{'{{.Job.Company}}'}</Text>
      <Text style={paragraph}>
        Segue abaixo a análise detalhada da compatibilidade do seu currículo com a vaga encontrada.
      </Text>

      {/* Score Card */}
      <Section style={scoreCard}>
        <Text style={cardLabel}>Compatibilidade</Text>
        <Text style={scoreValue}>
          {'{{.Analysis.MatchAnalysis.OverallScoreNumeric}}'}
          <span style={scoreMax}>/100</span>
        </Text>
        <Text style={qualitativeText}>
          {'{{.Analysis.MatchAnalysis.OverallScoreQualitative}}'}
        </Text>
        <Text style={summaryText}>{'{{.Analysis.MatchAnalysis.Summary}}'}</Text>
      </Section>

      {/* Strengths */}
      <Section style={card}>
        <Text style={cardTitle}>Pontos Fortes para esta Vaga</Text>
        <Text style={hidden}>{'{{range .Analysis.StrengthsForThisJob}}'}</Text>
        <Section style={listItem}>
          <Text style={listItemTitle}>{'{{.Point}}'}</Text>
          <Text style={listItemDetail}>{'Relevância: {{.RelevanceToJob}}'}</Text>
        </Section>
        <Text style={hidden}>{'{{end}}'}</Text>
        <Text style={hidden}>{'{{if not .Analysis.StrengthsForThisJob}}'}</Text>
        <Text style={emptyText}>Nenhum ponto forte específico identificado.</Text>
        <Text style={hidden}>{'{{end}}'}</Text>
      </Section>

      {/* Gaps */}
      <Section style={card}>
        <Text style={cardTitle}>Lacunas e Áreas de Melhoria</Text>
        <Text style={hidden}>{'{{range .Analysis.GapsAndImprovementAreas}}'}</Text>
        <Section style={listItem}>
          <Text style={listItemTitle}>{'{{.AreaDescription}}'}</Text>
          <Text style={listItemDetail}>{'Impacto: {{.JobRequirementImpacted}}'}</Text>
        </Section>
        <Text style={hidden}>{'{{end}}'}</Text>
        <Text style={hidden}>{'{{if not .Analysis.GapsAndImprovementAreas}}'}</Text>
        <Text style={emptyText}>Nenhuma lacuna específica identificada.</Text>
        <Text style={hidden}>{'{{end}}'}</Text>
      </Section>

      {/* Suggestions */}
      <Section style={card}>
        <Text style={cardTitle}>Sugestões para o Currículo</Text>
        <Text style={hidden}>{'{{range .Analysis.ActionableResumeSuggestions}}'}</Text>
        <Section style={listItem}>
          <Text style={listItemTitle}>{'{{.Suggestion}}'}</Text>
          <Text style={listItemDetail}>{'Seção: {{.CurriculumSectionToApply}}'}</Text>
          <Text style={listItemDetail}>{'Exemplo: "{{.ExampleWording}}"'}</Text>
          <Text style={listItemDetail}>{'Justificativa: {{.ReasoningForThisJob}}'}</Text>
        </Section>
        <Text style={hidden}>{'{{end}}'}</Text>
        <Text style={hidden}>{'{{if not .Analysis.ActionableResumeSuggestions}}'}</Text>
        <Text style={emptyText}>Nenhuma sugestão específica identificada.</Text>
        <Text style={hidden}>{'{{end}}'}</Text>
      </Section>

      <Hr style={sectionDivider} />
      <Text style={cardTitle}>Considerações Finais</Text>
      <Text style={paragraph}>{'{{.Analysis.FinalConsiderations}}'}</Text>

      <Section style={buttonContainer}>
        <Button style={button} href="{{.DashboardLink}}">Ver no Dashboard</Button>
      </Section>
    </BaseLayout>
  )
}

const heading: React.CSSProperties = {
  fontSize: '24px',
  fontWeight: 700,
  color: '#fafafa',
  margin: '0 0 4px 0',
  letterSpacing: '-0.02em',
}
const subheading: React.CSSProperties = {
  fontSize: '15px',
  color: '#a1a1aa',
  margin: '0 0 20px 0',
}
const paragraph: React.CSSProperties = {
  fontSize: '15px',
  lineHeight: '26px',
  color: '#d4d4d8',
  margin: '0 0 16px 0',
}
const scoreCard: React.CSSProperties = {
  backgroundColor: '#27272a',
  borderRadius: '10px',
  border: '1px solid #3f3f46',
  borderLeft: '4px solid #10b981',
  padding: '24px',
  marginBottom: '16px',
}
const cardLabel: React.CSSProperties = {
  fontSize: '12px',
  fontWeight: 600,
  color: '#a1a1aa',
  margin: '0 0 8px 0',
  textTransform: 'uppercase' as const,
  letterSpacing: '0.05em',
}
const scoreValue: React.CSSProperties = {
  fontSize: '42px',
  fontWeight: 700,
  color: '#10b981',
  margin: '0 0 4px 0',
  lineHeight: '1',
}
const scoreMax: React.CSSProperties = {
  fontSize: '18px',
  fontWeight: 400,
  color: '#71717a',
}
const qualitativeText: React.CSSProperties = {
  fontSize: '14px',
  fontWeight: 600,
  color: '#10b981',
  margin: '0 0 12px 0',
}
const summaryText: React.CSSProperties = {
  fontSize: '14px',
  lineHeight: '22px',
  color: '#a1a1aa',
  margin: 0,
}
const card: React.CSSProperties = {
  backgroundColor: '#27272a',
  borderRadius: '10px',
  border: '1px solid #3f3f46',
  padding: '24px',
  marginBottom: '16px',
}
const cardTitle: React.CSSProperties = {
  fontSize: '16px',
  fontWeight: 700,
  color: '#fafafa',
  margin: '0 0 16px 0',
}
const listItem: React.CSSProperties = {
  padding: '12px 0',
  borderBottom: '1px solid #3f3f46',
}
const listItemTitle: React.CSSProperties = {
  fontSize: '14px',
  fontWeight: 600,
  color: '#e4e4e7',
  margin: '0 0 4px 0',
}
const listItemDetail: React.CSSProperties = {
  fontSize: '13px',
  color: '#a1a1aa',
  margin: '0 0 2px 0',
  lineHeight: '20px',
}
const emptyText: React.CSSProperties = {
  fontSize: '14px',
  color: '#71717a',
  fontStyle: 'italic',
  margin: 0,
}
const hidden: React.CSSProperties = {
  display: 'none',
  fontSize: 0,
  lineHeight: 0,
  maxHeight: 0,
  overflow: 'hidden',
}
const sectionDivider: React.CSSProperties = {
  borderColor: '#27272a',
  margin: '24px 0',
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
