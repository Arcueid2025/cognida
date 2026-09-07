import { http } from '@/utils/request'
import type { PaymentDispositionDraft, PaymentIncident, PaymentIncidentDetail, PaymentIncidentListResponse, PaymentIncidentStatus, SimulatedCallbackLog, SimulatedPaymentOrder, SimulatedPaymentRecord } from '@/types'

export const paymentIncidentApi = {
  create(data: { order_id?: string; channel?: string; incident_type: string; priority?: string; assessment_request_id?: string; recommendation_version?: string; evidence_snapshot?: unknown }) {
    return http.post<PaymentIncident>('/payment-incidents', data)
  },
  list(params?: { page?: number; page_size?: number }) {
    return http.get<PaymentIncidentListResponse>('/payment-incidents', { params })
  },
  transition(id: string, status: PaymentIncidentStatus) {
    return http.post<PaymentIncident>(`/payment-incidents/${id}/status`, { status })
  },
  get(id: string) {
    return http.get<PaymentIncidentDetail>(`/payment-incidents/${id}`)
  },
  update(id: string, data: { assignee_id?: number; conclusion?: string }) {
    return http.put<PaymentIncident>(`/payment-incidents/${id}`, data)
  },
  addTimelineNote(id: string, data: { content: string; attachment_ref?: string }) {
    return http.post<{ message: string }>(`/payment-incidents/${id}/timeline`, data)
  },
  getSimulatedOrder(orderID: string) {
    return http.get<SimulatedPaymentOrder>(`/payment-read/orders/${encodeURIComponent(orderID)}`)
  },
  getSimulatedPayments(orderID: string) {
    return http.get<{ order_id: string; items: SimulatedPaymentRecord[] }>(`/payment-read/orders/${encodeURIComponent(orderID)}/payments`)
  },
  getSimulatedCallbackLogs(orderID: string) {
    return http.get<{ order_id: string; items: SimulatedCallbackLog[] }>(`/payment-read/orders/${encodeURIComponent(orderID)}/callback-logs`)
  }
  ,createDispositionDraft(id: string, data: { action_type: string; idempotency_key: string }) { return http.post<PaymentDispositionDraft>(`/payment-incidents/${id}/disposition-drafts`, data) }
  ,decideDispositionDraft(id: string, approve: boolean) { return http.post<PaymentDispositionDraft>(`/payment-incidents/disposition-drafts/${id}/decision`, { approve }) }
  ,executeDispositionDraft(id: string) { return http.post<PaymentDispositionDraft>(`/payment-incidents/disposition-drafts/${id}/execute`) }
}
