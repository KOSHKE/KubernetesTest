import React from 'react';
import { Routes, Route } from 'react-router-dom';
import Header from './components/Header';
import Layout from './components/Layout';
import HomePage from './pages/HomePage';
import ProductsPage from './pages/ProductsPage';
import CartPage from './pages/CartPage';
import OrdersPage from './pages/OrdersPage';

function App() {
  return (
    <div className="App">
      <Header />
      <main>
        <Layout>
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/products" element={<ProductsPage />} />
            <Route path="/cart" element={<CartPage />} />
            <Route path="/orders" element={<OrdersPage />} />
          </Routes>
        </Layout>
      </main>
    </div>
  );
}

export default App;