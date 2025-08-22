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

// Add request interceptor to include auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Products API (public endpoints: no credentials needed)
export const productsAPI = {
  getProducts: (params: { category_id?: string; search?: string } = {}) =>
    api.get('/inventory/products', { params, withCredentials: false }),
  getProduct: (id: string) =>
    api.get(`/inventory/products/${id}`, { withCredentials: false }),
  getCategories: () =>
    api.get('/inventory/categories?active_only=true', { withCredentials: false }),
};

// Orders API (requires authentication)
export const ordersAPI = {
  createOrder: (orderData: CreateOrderRequest) => {
    return api.post('/orders', orderData);
  },
  getOrders: (params: Record<string, unknown> = {}) => {
    return api.get('/orders', { params });
  },
  getOrder: (id: string) => {
    return api.get(`/orders/${id}`);
  },
};

// Auth API
export const authAPI = {
  login: (credentials: { email: string; password: string }) =>
    api.post('/auth/login', credentials),
  register: (userData: { email: string; password: string; first_name: string; last_name: string; phone: string }) =>
    api.post('/auth/register', userData),
  refreshToken: (refreshToken: string) =>
    api.post('/auth/refresh', { refresh_token: refreshToken }),
  logout: (refreshToken: string) =>
    api.post('/auth/logout', { refresh_token: refreshToken }),
};

export default api;