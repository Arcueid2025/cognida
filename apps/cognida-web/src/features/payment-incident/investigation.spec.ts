import { describe, expect, it } from 'vitest'
import { buildEvidenceBoundRecommendation, buildInvestigationChecklist, relativeMatchPercent, splitByQueryTerms } from './investigation'

describe('buildInvestigationChecklist', () => {
  it('marks timeout cases for time-window verification', () => {
    const checks = buildInvestigationChecklist('支付回调 timeout', [{ chunk_id: 'c1' } as never])
    expect(checks.join('\n')).toContain('超时发生位置与时间窗')
  })

  it('requires manual handling when no evidence is found', () => {
    const checks = buildInvestigationChecklist('未知支付异常', [])
    expect(checks.at(-1)).toContain('转人工处理')
  })

  it('normalizes RRF scores instead of presenting them as probabilities', () => {
    expect(relativeMatchPercent(0.02, 0.02)).toBe(100)
    expect(relativeMatchPercent(0.01, 0.02)).toBe(50)
  })

  it('marks order numbers and Chinese terms without injecting HTML', () => {
    expect(splitByQueryTerms('ORD-1020 的支付状态仍为 PAYING', 'ORD-1020 支付状态').filter(s => s.matched).map(s => s.text))
      .toEqual(['ORD-1020', '支付状态'])
  })

  it('only creates an evidence-bound recommendation when a citation is present', () => {
    const recommendation = buildEvidenceBoundRecommendation('支付成功但订单未更新', [
      { chunk_id: 'c1', knowledge_title: '状态机', content: '【PI-KB-003-3】先检查回调。', score: 0.8 } as never
    ])

    expect(recommendation.judgment).toContain('可追溯规则')
    expect(recommendation.evidenceReferences).toEqual(['PI-KB-003-3'])
    expect(recommendation.riskNotice).toContain('不得')
  })

  it('requires human escalation for uncited or absent evidence', () => {
    const recommendation = buildEvidenceBoundRecommendation('未知异常', [{ chunk_id: 'c1', content: '未编号片段' } as never])

    expect(recommendation.judgment).toBe('信息不足，建议人工升级')
    expect(recommendation.recommendedAction).toContain('补充订单号')
  })
})
