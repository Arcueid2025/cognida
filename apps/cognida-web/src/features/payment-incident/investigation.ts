import type { SearchResult } from '@/types'

export interface HighlightSegment {
  text: string
  matched: boolean
}

export interface EvidenceBoundRecommendation {
  judgment: string
  knownFacts: string[]
  pendingChecks: string[]
  recommendedAction: string
  riskNotice: string
  evidenceReferences: string[]
}

const evidenceReferencePattern = /【(PI-KB-\d+-\d+)】/g

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

// 处置建议只描述已取得的知识库证据和仍待核验的项目，绝不把用户输入当作订单事实。
// 这使 UI 在未接入真实订单、支付流水和回调日志前保持“先证据、后建议”的边界。
export function buildEvidenceBoundRecommendation(query: string, evidence: SearchResult[]): EvidenceBoundRecommendation {
  const evidenceReferences = extractEvidenceReferences(evidence)
  const pendingChecks = buildInvestigationChecklist(query, evidence)

  if (evidence.length === 0 || evidenceReferences.length === 0) {
    return {
      judgment: '信息不足，建议人工升级',
      knownFacts: evidence.length === 0
        ? ['本次检索没有返回可用于核验的知识库片段。']
        : ['本次检索返回了片段，但片段不含可追溯的证据编号。'],
      pendingChecks,
      recommendedAction: '补充订单号、支付流水号、订单状态、支付渠道状态和回调日志后，由人工继续核验。',
      riskNotice: '不得自动补单、回放回调、关闭订单或修改订单状态。',
      evidenceReferences
    }
  }

  const sources = [...new Set(evidence.map(item => item.knowledge_title).filter(Boolean))]
  return {
    judgment: '已检索到可追溯规则，仍需人工核验订单与支付事实',
    knownFacts: [
      `已检索到 ${evidence.length} 条知识库片段，来源：${sources.join('、')}。`,
      `可引用的规则证据：${evidenceReferences.join('、')}。`,
      '用户输入是待核验线索，不等同于订单、支付或回调的真实记录。'
    ],
    pendingChecks,
    recommendedAction: '按待核验项收集订单、支付与回调记录；确认事实与引用规则一致后，再由人工决定是否创建处置草稿。',
    riskNotice: '当前页面只提供核验建议；不得基于检索结果直接补单、回放回调、关闭订单或修改订单状态。',
    evidenceReferences
  }
}

export function extractEvidenceReferences(evidence: SearchResult[]): string[] {
  const references = new Set<string>()
  for (const item of evidence) {
    for (const match of item.content.matchAll(evidenceReferencePattern)) {
      references.add(match[1])
    }
  }
  return [...references].sort()
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
