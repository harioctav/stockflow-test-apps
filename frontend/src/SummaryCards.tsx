import type { Product } from './types'
import { formatRp } from './format'

type Props = {
  products: Product[]
  loading?: boolean
}

export function SummaryCards({ products, loading }: Props) {
  const total = products.length
  const active = products.filter((p) => p.is_active).length
  const stockValue = products.reduce((sum, p) => sum + p.price * p.stock, 0)
  const low = products.filter((p) => p.is_active && p.stock <= 10).length

  return (
    <div className="cards">
      <div className="card">
        <div className="label">Total products</div>
        <div className="value">{loading ? '…' : total}</div>
      </div>
      <div className="card">
        <div className="label">Active</div>
        <div className="value">{loading ? '…' : active}</div>
      </div>
      <div className="card">
        <div className="label">Stock value</div>
        <div className="value" style={{ fontSize: '1.05rem' }}>
          {loading ? '…' : formatRp(stockValue)}
        </div>
      </div>
      <div className="card">
        <div className="label">Low stock (≤10)</div>
        <div className="value" style={{ color: low > 0 ? 'var(--warn)' : undefined }}>
          {loading ? '…' : low}
        </div>
      </div>
    </div>
  )
}
