import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api, ApiError } from './api'
import type { OrderLineInput } from './validation'
import { validateOrderForm } from './validation'

export function CreateOrderPage() {
  const navigate = useNavigate()
  const [customer, setCustomer] = useState('')
  const [items, setItems] = useState<OrderLineInput[]>([{ sku: '', qty: '1' }])
  const [clientErrors, setClientErrors] = useState<string[]>([])
  const [serverError, setServerError] = useState<string | null>(null)
  const [failedSku, setFailedSku] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  function updateItem(idx: number, patch: Partial<OrderLineInput>) {
    setItems((prev) => prev.map((it, i) => (i === idx ? { ...it, ...patch } : it)))
  }

  function addLine() {
    setItems((prev) => [...prev, { sku: '', qty: '1' }])
  }

  function removeLine(idx: number) {
    setItems((prev) => (prev.length === 1 ? prev : prev.filter((_, i) => i !== idx)))
  }

  async function submit(e: FormEvent) {
    e.preventDefault()
    setServerError(null)
    setFailedSku(null)

    const errs = validateOrderForm(customer, items)
    setClientErrors(errs)
    if (errs.length) return

    setSaving(true)
    try {
      const order = await api.createOrder({
        customer: customer.trim(),
        items: items.map((it) => ({ sku: it.sku.trim(), qty: Number(it.qty) })),
      })
      navigate(`/orders/${order.id}`, { state: { justCreated: true } })
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === 'INSUFFICIENT_STOCK') {
          setFailedSku(err.sku || null)
          setServerError(
            `Insufficient stock for ${err.sku || 'item'}. Available: ${err.available ?? 0}`,
          )
        } else {
          setServerError(err.message)
        }
      } else {
        setServerError('Failed to create order')
      }
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <h1 className="page-title">Create order</h1>
      <p className="muted" style={{ marginTop: -4 }}>
        Prices are taken from the server — you only pick SKU + qty.
      </p>

      {(clientErrors.length > 0 || serverError) && (
        <div className="alert error">
          {serverError && <div>{serverError}</div>}
          {clientErrors.map((e) => (
            <div key={e}>{e}</div>
          ))}
        </div>
      )}

      <form onSubmit={submit} className="panel" style={{ padding: '1rem' }}>
        <div className="form-row">
          <label>Customer</label>
          <input
            value={customer}
            onChange={(e) => setCustomer(e.target.value)}
            placeholder="e.g. Toko Maju Jaya"
          />
        </div>

        <div style={{ marginBottom: 8, fontSize: '0.82rem', color: 'var(--muted)' }}>Items</div>
        {items.map((it, idx) => (
          <div className="line-row" key={idx}>
            <div className="form-row" style={{ marginBottom: 0 }}>
              <label>SKU</label>
              <input
                style={
                  failedSku && it.sku.trim() === failedSku
                    ? { borderColor: 'var(--danger)', background: '#fef2f2' }
                    : undefined
                }
                value={it.sku}
                onChange={(e) => updateItem(idx, { sku: e.target.value })}
                placeholder="SKU-BRG-001"
              />
            </div>
            <div className="form-row" style={{ marginBottom: 0 }}>
              <label>Qty</label>
              <input
                type="number"
                min={1}
                value={it.qty}
                onChange={(e) => updateItem(idx, { qty: e.target.value })}
              />
            </div>
            <button type="button" onClick={() => removeLine(idx)} disabled={items.length === 1}>
              Remove
            </button>
          </div>
        ))}

        <button type="button" onClick={addLine} style={{ marginTop: 4 }}>
          + Add item
        </button>

        <div className="form-actions">
          <Link to="/">
            <button type="button" disabled={saving}>
              Back
            </button>
          </Link>
          <button type="submit" className="primary" disabled={saving}>
            {saving ? 'Creating…' : 'Create order'}
          </button>
        </div>
      </form>

      <p className="muted" style={{ fontSize: '0.85rem' }}>
        Seed SKUs look like <code>SKU-BRG-001</code> … <code>SKU-BRG-014</code>.
      </p>
    </div>
  )
}
