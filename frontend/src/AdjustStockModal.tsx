import { useState, type FormEvent } from 'react'
import { api, ApiError } from './api'
import type { Product } from './types'
import { validateStockQty } from './validation'

type Props = {
  product: Product
  onClose: () => void
  onDone: () => void
}

export function AdjustStockModal({ product, onClose, onDone }: Props) {
  const [type, setType] = useState<'in' | 'out'>('in')
  const [qty, setQty] = useState('')
  const [note, setNote] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const qtyError = qty === '' ? null : validateStockQty(qty)
  const canSubmit = validateStockQty(qty) === null && !saving

  async function submit(e: FormEvent) {
    e.preventDefault()
    const v = validateStockQty(qty)
    if (v) {
      setError(v)
      return
    }
    setSaving(true)
    setError(null)
    try {
      await api.adjustStock(product.id, {
        type,
        qty: Number(qty),
        note: note.trim() || undefined,
      })
      onDone()
    } catch (err) {
      if (err instanceof ApiError && err.code === 'INSUFFICIENT_STOCK') {
        setError(`Insufficient stock. Available: ${err.available ?? 0}`)
      } else if (err instanceof Error) {
        setError(err.message)
      } else {
        setError('Failed to adjust stock')
      }
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h2>Adjust stock — {product.sku}</h2>
        <p className="muted" style={{ marginTop: 0 }}>
          Current stock: <strong>{product.stock}</strong>
        </p>
        {error && <div className="alert error">{error}</div>}
        <form onSubmit={submit}>
          <div className="form-row">
            <label>Direction</label>
            <select value={type} onChange={(e) => setType(e.target.value as 'in' | 'out')}>
              <option value="in">Stock in</option>
              <option value="out">Stock out</option>
            </select>
          </div>
          <div className="form-row">
            <label>Qty</label>
            <input
              type="number"
              min={1}
              step={1}
              value={qty}
              onChange={(e) => setQty(e.target.value)}
              autoFocus
            />
            {qtyError && <div className="muted" style={{ color: 'var(--danger)', marginTop: 4 }}>{qtyError}</div>}
          </div>
          <div className="form-row">
            <label>Note (optional)</label>
            <textarea rows={2} value={note} onChange={(e) => setNote(e.target.value)} />
          </div>
          <div className="form-actions">
            <button type="button" onClick={onClose} disabled={saving}>
              Cancel
            </button>
            <button type="submit" className="primary" disabled={!canSubmit}>
              {saving ? 'Saving…' : 'Save'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
