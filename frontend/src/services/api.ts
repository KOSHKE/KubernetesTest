import axios from 'axios';
import type { CreateOrderRequest } from '../types';

const API_BASE_URL = (import.meta as any).env?.VITE_API_URL || process.env.REACT_APP_API_URL || 'http://localhost:8080';

// Create axios instance
const api = axios.create({
  baseURL: `${API_BASE_URL}/api/v1`,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: false,
});

// Products API (public endpoints: no credentials needed)
export const productsAPI = {
  getProducts: (params: { category_id?: string; search?: string } = {}) =>
    api.get('/inventory/products', { params, withCredentials: false }),
  getProduct: (id: string) =>
    api.get(`/inventory/products/${id}`, { withCredentials: false }),
  getCategories: () =>
    api.get('/inventory/categories?active_only=true', { withCredentials: false }),
};

// Orders API (temporarily public in dev-noauth)
export const ordersAPI = {
  createOrder: (orderData: CreateOrderRequest) => api.post('/orders', orderData, { withCredentials: false }),
  getOrders: (params: Record<string, unknown> = {}) => api.get('/orders', { params, withCredentials: false }),
  getOrder: (id: string) => api.get(`/orders/${id}`, { withCredentials: false }),
  cancelOrder: (id: string) => api.put(`/orders/${id}/cancel`, undefined, { withCredentials: false }),
};

export default api;