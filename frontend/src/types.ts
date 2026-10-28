export type Product = {
  id: number
  sku: string
  name: string
  price: number
  stock: number
  is_active: boolean
  created_at: string
  last_movement_at: string | null
}

export type ProductListResponse = {
  data: Product[]
  page: number
  limit: number
  total: number
}

export type StockMovement = {
  id: number
  product_id: number
  type: 'in' | 'out'
  qty: number
  note: string | null
  created_at: string
}

export type OrderItem = {
  id: number
  order_id: number
  sku: string
  qty: number
  unit_price: number
}

export type SalesOrder = {
  id: number
  customer: string
  status: 'CREATED' | 'CANCELLED'
  total: number
  created_at: string
  cancelled_at: string | null
  items?: OrderItem[]
}

export type ApiErrorBody = {
  error: {
    code?: string
    message?: string
    available?: number
    sku?: string
  }
}

export type CreateOrderPayload = {
  customer: string
  items: { sku: string; qty: number }[]
}

export type AdjustStockPayload = {
  type: 'in' | 'out'
  qty: number
  note?: string
}
