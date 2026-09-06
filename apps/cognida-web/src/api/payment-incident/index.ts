import { http } from '@/utils/request'
import type { PaymentIncident, PaymentIncidentListResponse, PaymentIncidentStatus } from '@/types'

export const paymentIncidentApi = {
  create(data: { order_id?: string; channel?: string; incident_type: string; priority?: string; assessment_request_id?: string }) {
    return http.post<PaymentIncident>('/payment-incidents', data)
  },
  list(params?: { page?: number; page_size?: number }) {
    return http.get<PaymentIncidentListResponse>('/payment-incidents', { params })
  },
  transition(id: string, status: PaymentIncidentStatus) {
    return http.post<PaymentIncident>(`/payment-incidents/${id}/status`, { status })
  }
}
