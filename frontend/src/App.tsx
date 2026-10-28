import { NavLink, Route, Routes } from 'react-router-dom'
import { CreateOrderPage } from './CreateOrderPage'
import { OrderDetailPage } from './OrderDetailPage'
import { ProductsPage } from './ProductsPage'

export default function App() {
  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="brand">
          Stock<span>Flow</span>
        </div>
        <nav className="nav">
          <NavLink to="/" end className={({ isActive }) => (isActive ? 'active' : undefined)}>
            Products
          </NavLink>
          <NavLink to="/orders/new" className={({ isActive }) => (isActive ? 'active' : undefined)}>
            New order
          </NavLink>
        </nav>
      </header>

      <Routes>
        <Route path="/" element={<ProductsPage />} />
        <Route path="/orders/new" element={<CreateOrderPage />} />
        <Route path="/orders/:id" element={<OrderDetailPage />} />
      </Routes>
    </div>
  )
}
