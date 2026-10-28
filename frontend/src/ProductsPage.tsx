import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, ApiError } from './api'
import { AdjustStockModal } from './AdjustStockModal'
import { SummaryCards } from './SummaryCards'
import { formatRp, relativeTime } from './format'
import type { Product } from './types'
import { useDebounce } from './useDebounce'

type Filter = 'all' | 'active' | 'inactive' | 'low'

export function ProductsPage() {
  const [products, setProducts] = useState<Product[]>([])
  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState<Filter>('all')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [adjustTarget, setAdjustTarget] = useState<Product | null>(null)
  const [actionMsg, setActionMsg] = useState<string | null>(null)

  const debouncedSearch = useDebounce(search, 350)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      // limit 100 is enough for seed + summary cards
      const res = await api.listProducts({ search: debouncedSearch, page: 1, limit: 100 })
      setProducts(res.data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load products')
      setProducts([])
    } finally {
      setLoading(false)
    }
  }, [debouncedSearch])

  useEffect(() => {
    void load()
  }, [load])

  const filtered = products.filter((p) => {
    if (filter === 'active') return p.is_active
    if (filter === 'inactive') return !p.is_active
    if (filter === 'low') return p.is_active && p.stock <= 10
    return true
  })

  async function handleDelete(p: Product) {
    if (!window.confirm(`Delete ${p.sku}?`)) return
    setActionMsg(null)
    try {
      await api.deleteProduct(p.id)
      setActionMsg(`Deleted ${p.sku}`)
      void load()
    } catch (err) {
      if (err instanceof ApiError && err.code === 'PRODUCT_HAS_HISTORY') {
        setActionMsg(`Cannot delete ${p.sku}: has movement history`)
      } else {
        setActionMsg(err instanceof Error ? err.message : 'Delete failed')
      }
    }
  }

  return (
    <div>
      <h1 className="page-title">Products</h1>
      <SummaryCards products={products} loading={loading && products.length === 0} />

      <div className="toolbar">
        <input
          placeholder="Search SKU or name…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          style={{ minWidth: 220 }}
        />
        <select value={filter} onChange={(e) => setFilter(e.target.value as Filter)}>
          <option value="all">All</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
          <option value="low">Low stock (≤10)</option>
        </select>
        <Link to="/orders/new">
          <button type="button" className="primary">
            New order
          </button>
        </Link>
      </div>

      {actionMsg && (
        <div className={`alert ${actionMsg.startsWith('Cannot') || actionMsg.includes('fail') ? 'error' : 'ok'}`}>
          {actionMsg}
        </div>
      )}

      <div className="panel">
        {loading && (
          <div className="state-box">Loading products…</div>
        )}
        {!loading && error && (
          <div className="state-box error">
            <div>{error}</div>
            <button type="button" style={{ marginTop: 12 }} onClick={() => void load()}>
              Retry
            </button>
          </div>
        )}
        {!loading && !error && filtered.length === 0 && (
          <div className="state-box">No products match.</div>
        )}
        {!loading && !error && filtered.length > 0 && (
          <table>
            <thead>
              <tr>
                <th>SKU</th>
                <th>Name</th>
                <th>Price</th>
                <th>Stock</th>
                <th>Status</th>
                <th>Last movement</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((p) => (
                <tr key={p.id}>
                  <td className="mono">{p.sku}</td>
                  <td>{p.name}</td>
                  <td>{formatRp(p.price)}</td>
                  <td style={{ color: p.stock <= 10 ? 'var(--warn)' : undefined, fontWeight: 600 }}>
                    {p.stock}
                  </td>
                  <td>
                    <span className={`badge ${p.is_active ? 'active' : 'inactive'}`}>
                      {p.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td className="muted">{relativeTime(p.last_movement_at)}</td>
                  <td>
                    <div className="actions">
                      <button type="button" className="linkish" onClick={() => setAdjustTarget(p)}>
                        Adjust
                      </button>
                      <button type="button" className="linkish danger" onClick={() => void handleDelete(p)}>
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {adjustTarget && (
        <AdjustStockModal
          product={adjustTarget}
          onClose={() => setAdjustTarget(null)}
          onDone={() => {
            setAdjustTarget(null)
            void load()
          }}
        />
      )}
    </div>
  )
}
