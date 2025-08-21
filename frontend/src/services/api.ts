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
  createOrder: (orderData: CreateOrderRequest) => {
    // user_id is included in the request body
    return api.post('/orders', orderData, { withCredentials: false });
  },
  getOrders: (params: Record<string, unknown> = {}) => {
    // Add user_id if not provided
    const finalParams = { ...params };
    if (!finalParams.user_id) {
      finalParams.user_id = 'dev-user-1'; // Temporary user_id for development
    }
    return api.get('/orders', { params: finalParams, withCredentials: false });
  },
  getOrder: (id: string) => {
    const params = { user_id: 'dev-user-1' }; // Temporary user_id for development
    return api.get(`/orders/${id}`, { params, withCredentials: false });
  },

};

export default api;