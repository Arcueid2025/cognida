import { describe, expect, it } from 'vitest'
import { availableIncidentTransitions, extractOrderID, inferIncidentType } from './cases'

describe('payment incident cases', () => {
  it('only exposes allowed state transitions', () => {
    expect(availableIncidentTransitions('pending_verification')).toEqual(['investigating', 'escalated'])
    expect(availableIncidentTransitions('resolved')).toEqual([])
  })
  it('derives safe draft fields from the query', () => {
    expect(extractOrderID('ORD-1020 回调超时')).toBe('ORD-1020')
    expect(inferIncidentType('支付回调超时')).toBe('callback_failure')
  })
})
