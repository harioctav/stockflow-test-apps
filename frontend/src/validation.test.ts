import { describe, expect, it } from 'vitest'
import { validateOrderForm, validateStockQty } from './validation'

describe('validateStockQty', () => {
  it('rejects empty / non-integer / <=0', () => {
    expect(validateStockQty('')).not.toBeNull()
    expect(validateStockQty('1.5')).not.toBeNull()
    expect(validateStockQty('0')).not.toBeNull()
    expect(validateStockQty('-2')).not.toBeNull()
  })

  it('accepts positive integers', () => {
    expect(validateStockQty('1')).toBeNull()
    expect(validateStockQty('42')).toBeNull()
  })
})

describe('validateOrderForm', () => {
  it('requires customer and items', () => {
    const errs = validateOrderForm('', [])
    expect(errs.some((e) => e.toLowerCase().includes('customer'))).toBe(true)
    expect(errs.some((e) => e.toLowerCase().includes('item'))).toBe(true)
  })

  it('flags duplicate skus and bad qty', () => {
    const errs = validateOrderForm('Andi', [
      { sku: 'A', qty: '2' },
      { sku: 'A', qty: '0' },
    ])
    expect(errs.some((e) => e.toLowerCase().includes('duplicate'))).toBe(true)
    expect(errs.some((e) => e.toLowerCase().includes('qty'))).toBe(true)
  })

  it('passes a normal order', () => {
    expect(
      validateOrderForm('Budi', [
        { sku: 'SKU-1', qty: '1' },
        { sku: 'SKU-2', qty: '3' },
      ]),
    ).toEqual([])
  })
})
