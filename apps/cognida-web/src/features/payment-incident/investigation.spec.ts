import { describe, expect, it } from 'vitest'
import { buildInvestigationChecklist, relativeMatchPercent, splitByQueryTerms } from './investigation'

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
})
