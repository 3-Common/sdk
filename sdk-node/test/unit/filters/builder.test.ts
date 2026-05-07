import { describe, expect, it } from 'vitest'

import { and, combine, field, filter, or } from '@/filters'

describe('field', () => {
  it('rejects empty names', () => {
    expect(() => field('')).toThrow(TypeError)
  })

  it('rejects non-string names', () => {
    expect(() => field(42 as unknown as string)).toThrow(TypeError)
  })

  describe('operator methods', () => {
    const f = field('status')

    it.each([
      ['isEmpty', f.isEmpty(), { field: 'status', operator: 'is_empty' }],
      ['isNotEmpty', f.isNotEmpty(), { field: 'status', operator: 'is_not_empty' }],
    ])('%s — no value', (_name, condition, expected) => {
      expect(condition).toEqual(expected)
    })

    it('isEqualTo / isNotEqualTo', () => {
      expect(f.isEqualTo('open')).toEqual({
        field: 'status',
        operator: 'is_equal_to',
        value: 'open',
      })
      expect(f.isNotEqualTo(5)).toEqual({ field: 'status', operator: 'is_not_equal_to', value: 5 })
    })

    it('isEqualToAnyOf / isNotEqualToAnyOf', () => {
      expect(f.isEqualToAnyOf(['open'])).toEqual({
        field: 'status',
        operator: 'is_equal_to_any_of',
        value: ['open'],
      })
      expect(f.isNotEqualToAnyOf(['open'])).toEqual({
        field: 'status',
        operator: 'is_not_equal_to_any_of',
        value: ['open'],
      })
    })

    it('isAnyOf / isNoneOf', () => {
      expect(f.isAnyOf(['open', 'closed'])).toEqual({
        field: 'status',
        operator: 'is_any_of',
        value: ['open', 'closed'],
      })
      expect(f.isNoneOf(['draft'])).toEqual({
        field: 'status',
        operator: 'is_none_of',
        value: ['draft'],
      })
    })

    it('contains / containsExactly', () => {
      expect(f.contains('foo')).toEqual({ field: 'status', operator: 'contains', value: 'foo' })
      expect(f.containsExactly('Foo')).toEqual({
        field: 'status',
        operator: 'contains_exactly',
        value: 'Foo',
      })
    })

    it('isBefore / isAfter', () => {
      const date = '2026-05-01T00:00:00.000Z'
      expect(f.isBefore(date)).toEqual({ field: 'status', operator: 'is_before', value: date })
      expect(f.isAfter(date)).toEqual({ field: 'status', operator: 'is_after', value: date })
    })

    it('isGreaterThan / isLessThan and their _or_equal_to variants', () => {
      expect(f.isGreaterThan(10)).toEqual({
        field: 'status',
        operator: 'is_greater_than',
        value: 10,
      })
      expect(f.isGreaterThanOrEqualTo(10)).toEqual({
        field: 'status',
        operator: 'is_greater_than_or_equal_to',
        value: 10,
      })
      expect(f.isLessThan(10)).toEqual({
        field: 'status',
        operator: 'is_less_than',
        value: 10,
      })
      expect(f.isLessThanOrEqualTo(10)).toEqual({
        field: 'status',
        operator: 'is_less_than_or_equal_to',
        value: 10,
      })
    })

    it('isBetween — date range', () => {
      const range = { start: '2026-01-01T00:00:00.000Z', end: '2026-12-31T23:59:59.999Z' }
      expect(f.isBetween(range)).toEqual({
        field: 'status',
        operator: 'is_between',
        value: range,
      })
    })

    it('isBetween — numeric range', () => {
      expect(f.isBetween({ start: 10, end: 20 })).toEqual({
        field: 'status',
        operator: 'is_between',
        value: { start: 10, end: 20 },
      })
    })
  })
})

describe('and / or', () => {
  it('rejects empty calls', () => {
    expect(() => and()).toThrow(TypeError)
    expect(() => or()).toThrow(TypeError)
  })

  it('builds an AND group', () => {
    const f = and(field('status').isEqualTo('open'))
    expect(f.logic).toBe('and')
    expect(f.conditions).toHaveLength(1)
  })

  it('builds an OR group', () => {
    const f = or(field('a').isGreaterThan(1), field('b').isLessThan(10))
    expect(f.logic).toBe('or')
    expect(f.conditions).toHaveLength(2)
  })

  it('groups can be nested', () => {
    const inner = or(field('a').isEqualTo(1), field('b').isEqualTo(2))
    const outer = and(field('status').isEqualTo('open'), inner)
    expect(outer.conditions).toHaveLength(2)
    expect(outer.conditions[1]).toMatchObject({ logic: 'or' })
  })

  it('serialize wraps the group in an array (the API wire format)', () => {
    const f = and(field('status').isEqualTo('open'))
    const serialized = JSON.parse(f.serialize()) as unknown
    expect(Array.isArray(serialized)).toBe(true)
    expect(serialized).toEqual([
      { logic: 'and', conditions: [{ field: 'status', operator: 'is_equal_to', value: 'open' }] },
    ])
  })

  it('toFilters returns the array form without re-stringifying', () => {
    const f = or(field('x').isEmpty())
    const arr = f.toFilters()
    expect(arr).toHaveLength(1)
    expect(arr[0]?.logic).toBe('or')
  })
})

describe('combine', () => {
  it('rejects empty calls', () => {
    expect(() => combine()).toThrow(TypeError)
  })

  it('combines multiple top-level groups', () => {
    const g1 = and(field('status').isEqualTo('open'))
    const g2 = or(field('ticketSum').isGreaterThan(10))
    const composed = combine(g1, g2)
    expect(composed.toFilters()).toHaveLength(2)
    const serialized = JSON.parse(composed.serialize()) as unknown
    expect(Array.isArray(serialized)).toBe(true)
    expect((serialized as unknown[]).length).toBe(2)
  })
})

describe('namespace export', () => {
  it('exposes field / and / or / combine', () => {
    expect(typeof filter.field).toBe('function')
    expect(typeof filter.and).toBe('function')
    expect(typeof filter.or).toBe('function')
    expect(typeof filter.combine).toBe('function')
    // Sanity: the namespace methods produce the same shape as the named exports.
    expect(filter.and(filter.field('x').isEmpty()).logic).toBe('and')
  })
})

describe('integration with EventListParams', () => {
  it('produces a string that can be passed verbatim to events.list', () => {
    const f = and(
      field('status').isAnyOf(['open', 'closed']),
      or(field('ticketSum').isGreaterThan(10), field('revenueCents').isGreaterThanOrEqualTo(10000)),
    )
    const filtersParam: string = f.serialize()
    expect(typeof filtersParam).toBe('string')
    const parsed = JSON.parse(filtersParam) as unknown
    expect(parsed).toEqual([
      {
        logic: 'and',
        conditions: [
          { field: 'status', operator: 'is_any_of', value: ['open', 'closed'] },
          {
            logic: 'or',
            conditions: [
              { field: 'ticketSum', operator: 'is_greater_than', value: 10 },
              {
                field: 'revenueCents',
                operator: 'is_greater_than_or_equal_to',
                value: 10000,
              },
            ],
          },
        ],
      },
    ])
  })
})
