import type { SearchResult } from '@/types'

export interface HighlightSegment {
  text: string
  matched: boolean
}

export const paymentStarterQuestions = [
  'ORD-1020 支付成功但订单仍为 PAYING，应该如何排查？',
  '支付回调超时后，如何判断是否需要补单？',
  '回调重放前需要核验哪些订单与幂等信息？'
]

export function buildInvestigationChecklist(query: string, evidence: SearchResult[]): string[] {
  const checks = [
    '核对订单号、支付流水号、当前订单状态与支付渠道最终状态。',
    '核对回调是否收到、签名是否通过，以及幂等键是否已有成功消费记录。'
  ]

  if (/超时|timeout/i.test(query)) {
    checks.push('记录超时发生位置与时间窗，确认后再决定是否重试或回放。')
  }
  if (/PAYING|处理中|成功/i.test(query)) {
    checks.push('比对支付成功时间与订单状态变更记录，排除状态更新延迟。')
  }
  if (evidence.length === 0) {
    checks.push('当前知识库未检索到直接证据，请转人工处理，不执行自动补单或回放。')
  }

  return checks
}

// RRF 融合分不是概率，使用同一次结果中的最大分归一化后才展示为相对匹配度。
export function relativeMatchPercent(score: number, maxScore: number): number {
  if (maxScore <= 0 || score <= 0) return 0
  return Math.round(Math.min(score / maxScore, 1) * 100)
}

export function splitByQueryTerms(content: string, query: string): HighlightSegment[] {
  const terms = [...new Set(query.match(/[A-Za-z0-9_-]{2,}|[\u4e00-\u9fff]{2,}/g) ?? [])]
    .sort((a, b) => b.length - a.length)
  if (!content || terms.length === 0) return [{ text: content, matched: false }]

  const expression = new RegExp(`(${terms.map(escapeRegExp).join('|')})`, 'gi')
  return content.split(expression).filter(Boolean).map(text => ({
    text,
    matched: terms.some(term => term.toLowerCase() === text.toLowerCase())
  }))
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}
