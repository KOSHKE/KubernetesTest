import type { ReactNode } from 'react';
export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  phone: string;
  created_at: string;
  updated_at: string;
}

export interface Money {
  amount: number; // minor units
  currency: string;
}

export interface Product {
  id: string;
  name: string;
  description: string;
  price: Money;
  category_id: string;
  category_name: string;
  stock_quantity: number;
  image_url: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Category {
  id: string;
  name: string;
  description: string;
  is_active: boolean;
}

export interface CartItem {
  product_id: string;
  product_name: string;
  price: Money; // minor units in price.amount
  quantity: number;
}

export interface OrderItem {
  id: string;
  product_id: string;
  product_name: string;
  quantity: number;
  price: Money; // minor units in price.amount
  total: Money; // minor units in total.amount
}

export interface Order {
  id: string;
  user_id: string;
  status: string;
  items: OrderItem[];
  total_amount: Money; // minor units in total_amount.amount
  shipping_address: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  user: User;
  token: string;
  message: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  phone: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthProviderProps {
  children: ReactNode;
}

export interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (credentials: LoginRequest) => Promise<{ success: boolean; error?: string }>;
  register: (userData: RegisterRequest) => Promise<{ success: boolean; error?: string }>;
  logout: () => void;
  updateProfile: (profileData: Partial<RegisterRequest>) => Promise<{ success: boolean; error?: string }>;
  isAuthenticated: boolean;
}

export interface PaymentDetails {
  card_number: string;
  card_holder: string;
  expiry_month: string;
  expiry_year: string;
  cvv: string;
}

export interface CreateOrderRequest {
  items: { product_id: string; quantity: number }[];
  shipping_address: string;
  payment_method: string;
  payment_details: PaymentDetails;
}