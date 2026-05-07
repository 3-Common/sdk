import { describe, expect, it } from 'vitest'

import { ThreeCommon } from '@/client'

describe('ThreeCommon', () => {
  it('exposes an `events` resource with the documented methods', () => {
    const client = new ThreeCommon({ apiKey: '3co_x' })
    expect(client.events).toBeDefined()
    expect(typeof client.events.list).toBe('function')
    expect(typeof client.events.retrieve).toBe('function')
    expect(typeof client.events.update).toBe('function')
    expect(typeof client.events.listAutoPaginate).toBe('function')
  })

  it('throws when no apiKey is supplied and env var is unset', () => {
    const original = process.env['THREECOMMON_API_KEY']
    delete process.env['THREECOMMON_API_KEY']
    try {
      expect(() => new ThreeCommon()).toThrow(/API key is required/u)
    } finally {
      if (original !== undefined) process.env['THREECOMMON_API_KEY'] = original
    }
  })

  it('disableTelemetry() is callable', () => {
    const client = new ThreeCommon({ apiKey: '3co_x' })
    expect(() => {
      client.disableTelemetry()
    }).not.toThrow()
  })
})
