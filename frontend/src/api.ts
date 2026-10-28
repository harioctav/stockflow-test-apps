import type {
  AdjustStockPayload,
  ApiErrorBody,
  CreateOrderPayload,
  Product,
  ProductListResponse,
  SalesOrder,
  StockMovement,
} from './types'

export class ApiError extends Error {
  status: number
  code?: string
  available?: number
  sku?: string

  constructor(status: number, body: ApiErrorBody | null, fallback: string) {
    const msg = body?.error?.message || fallback
    super(msg)
    this.status = status
    this.code = body?.error?.code
    this.available = body?.error?.available
    this.sku = body?.error?.sku
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers || {}),
    },
  })

  if (res.status === 204) {
    return undefined as T
  }

  const text = await res.text()
  let body: unknown = null
  if (text) {
    try {
      body = JSON.parse(text)
    } catch {
      body = null
    }
  }

  if (!res.ok) {
    throw new ApiError(res.status, body as ApiErrorBody | null, res.statusText || 'Request failed')
  }

  return body as T
}

export const api = {
  listProducts(params: { search?: string; page?: number; limit?: number } = {}) {
    const q = new URLSearchParams()
    if (params.search) q.set('search', params.search)
    if (params.page) q.set('page', String(params.page))
    if (params.limit) q.set('limit', String(params.limit))
    const qs = q.toString()
    return request<ProductListResponse>(`/api/products${qs ? `?${qs}` : ''}`)
  },

  createProduct(payload: { sku: string; name: string; price: number; stock?: number }) {
    return request<Product>('/api/products', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  updateProduct(id: number, payload: { sku: string; name: string; price: number; is_active: boolean }) {
    return request<Product>(`/api/products/${id}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    })
  },

  deleteProduct(id: number) {
    return request<void>(`/api/products/${id}`, { method: 'DELETE' })
  },

  adjustStock(id: number, payload: AdjustStockPayload) {
    return request<Product>(`/api/products/${id}/stock`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  listMovements(id: number) {
    return request<{ data: StockMovement[] }>(`/api/products/${id}/movements`)
  },

  createOrder(payload: CreateOrderPayload) {
    return request<SalesOrder>('/api/orders', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  getOrder(id: number) {
    return request<SalesOrder>(`/api/orders/${id}`)
  },

  cancelOrder(id: number) {
    return request<SalesOrder>(`/api/orders/${id}/cancel`, { method: 'POST' })
  },
}
