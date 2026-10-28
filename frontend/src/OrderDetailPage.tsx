import { useCallback, useEffect, useState } from 'react'
import { Link, useLocation, useParams } from 'react-router-dom'
import { api, ApiError } from './api'
import { formatRp } from './format'
import type { SalesOrder } from './types'

export function OrderDetailPage() {
  const { id } = useParams()
  const location = useLocation()
  const justCreated = Boolean((location.state as { justCreated?: boolean } | null)?.justCreated)

  const [order, setOrder] = useState<SalesOrder | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [cancelling, setCancelling] = useState(false)
  const [msg, setMsg] = useState<string | null>(justCreated ? 'Order created.' : null)

  const load = useCallback(async () => {
    if (!id) return
    setLoading(true)
    setError(null)
    try {
      const o = await api.getOrder(Number(id))
      setOrder(o)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load order')
      setOrder(null)
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    void load()
  }, [load])

  async function cancel() {
    if (!order) return
    if (!window.confirm(`Cancel order #${order.id}? Stock will be restored.`)) return
    setCancelling(true)
    setMsg(null)
    try {
      const updated = await api.cancelOrder(order.id)
      setOrder(updated)
      setMsg('Order cancelled. Stock restored — check the products list.')
    } catch (err) {
      if (err instanceof ApiError) {
        setMsg(err.message)
      } else {
        setMsg('Cancel failed')
      }
    } finally {
      setCancelling(false)
    }
  }

  if (loading) {
    return <div className="state-box">Loading order…</div>
  }

  if (error || !order) {
    return (
      <div className="state-box error">
        <div>{error || 'Order not found'}</div>
        <button type="button" style={{ marginTop: 12 }} onClick={() => void load()}>
          Retry
        </button>
      </div>
    )
  }

  return (
    <div>
      <div className="toolbar" style={{ justifyContent: 'space-between' }}>
        <h1 className="page-title" style={{ margin: 0 }}>
          Order #{order.id}
        </h1>
        <Link to="/">← Products</Link>
      </div>

      {msg && (
        <div className={`alert ${msg.toLowerCase().includes('fail') || msg.toLowerCase().includes('already') ? 'error' : 'ok'}`}>
          {msg}
        </div>
      )}

      <div className="panel" style={{ padding: '1rem', marginBottom: '1rem' }}>
        <div style={{ display: 'grid', gap: 6 }}>
          <div>
            <span className="muted">Customer: </span>
            <strong>{order.customer}</strong>
          </div>
          <div>
            <span className="muted">Status: </span>
            <span className={`badge ${order.status === 'CREATED' ? 'created' : 'cancelled'}`}>
              {order.status}
            </span>
          </div>
          <div>
            <span className="muted">Total: </span>
            <strong>{formatRp(order.total)}</strong>
          </div>
          <div className="muted" style={{ fontSize: '0.85rem' }}>
            Created {new Date(order.created_at).toLocaleString('id-ID')}
            {order.cancelled_at && ` · Cancelled ${new Date(order.cancelled_at).toLocaleString('id-ID')}`}
          </div>
        </div>
      </div>

      <div className="panel">
        <table>
          <thead>
            <tr>
              <th>SKU</th>
              <th>Qty</th>
              <th>Unit price</th>
              <th>Line total</th>
            </tr>
          </thead>
          <tbody>
            {(order.items || []).map((it) => (
              <tr key={it.id}>
                <td className="mono">{it.sku}</td>
                <td>{it.qty}</td>
                <td>{formatRp(it.unit_price)}</td>
                <td>{formatRp(it.unit_price * it.qty)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {order.status === 'CREATED' && (
        <div style={{ marginTop: '1rem' }}>
          <button type="button" className="danger" disabled={cancelling} onClick={() => void cancel()}>
            {cancelling ? 'Cancelling…' : 'Cancel order'}
          </button>
        </div>
      )}
    </div>
  )
}
