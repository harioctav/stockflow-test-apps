export type OrderLineInput = {
  sku: string
  qty: string
}

export function validateOrderForm(customer: string, items: OrderLineInput[]): string[] {
  const errors: string[] = []
  if (!customer.trim()) {
    errors.push('Customer name is required')
  }
  if (items.length === 0) {
    errors.push('Add at least one item')
  }

  const seen = new Set<string>()
  items.forEach((it, idx) => {
    const sku = it.sku.trim()
    const label = `Item ${idx + 1}`
    if (!sku) {
      errors.push(`${label}: SKU is required`)
    } else if (seen.has(sku)) {
      errors.push(`${label}: duplicate SKU ${sku}`)
    } else {
      seen.add(sku)
    }

    const qty = Number(it.qty)
    if (!Number.isInteger(qty) || qty <= 0) {
      errors.push(`${label}: qty must be an integer > 0`)
    }
  })

  return errors
}

export function validateStockQty(qty: string): string | null {
  const n = Number(qty)
  if (!Number.isInteger(n) || n <= 0) {
    return 'Qty must be an integer greater than 0'
  }
  return null
}
