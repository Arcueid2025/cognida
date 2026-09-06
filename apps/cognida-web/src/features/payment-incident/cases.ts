import type { PaymentIncidentStatus } from '@/types'

const nextStatuses: Record<PaymentIncidentStatus, PaymentIncidentStatus[]> = {
  pending_verification: ['investigating', 'escalated'],
  investigating: ['awaiting_confirmation', 'resolved', 'escalated'],
  awaiting_confirmation: ['investigating', 'resolved', 'escalated'],
  resolved: [],
  escalated: []
}

const labels: Record<PaymentIncidentStatus, string> = {
  pending_verification: '待核验', investigating: '处理中', awaiting_confirmation: '待确认', resolved: '已解决', escalated: '已升级'
}

export function availableIncidentTransitions(status: PaymentIncidentStatus): PaymentIncidentStatus[] { return nextStatuses[status] }
export function incidentStatusLabel(status: PaymentIncidentStatus): string { return labels[status] }
export function inferIncidentType(query: string): string { return /回调|callback/i.test(query) ? 'callback_failure' : 'payment_timeout' }
export function extractOrderID(query: string): string { return query.match(/\bORD-[A-Za-z0-9-]+\b/i)?.[0]?.toUpperCase() ?? '' }
