import React, { useState, useEffect } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import Header from './components/Header';
import Layout from './components/Layout';
import HomePage from './pages/HomePage';
import ProductsPage from './pages/ProductsPage';
import CartPage from './pages/CartPage';
import OrdersPage from './pages/OrdersPage';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import UnauthorizedView from './components/UnauthorizedView';
import { authService } from './services/auth';

interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
}

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    // Check if user is authenticated on app load
    const checkAuth = () => {
      const authenticated = authService.isAuthenticated();
      setIsAuthenticated(authenticated);
      
      // If authenticated, try to get user info from localStorage
      if (authenticated) {
        const userInfo = localStorage.getItem('user_info');
        if (userInfo) {
          try {
            setUser(JSON.parse(userInfo));
          } catch (error) {
            console.error('Failed to parse user info:', error);
          }
        }
      }
      
      setIsLoading(false);
    };

    checkAuth();
  }, []);

  const handleLogout = async () => {
    await authService.logout();
    setIsAuthenticated(false);
    setUser(null);
    localStorage.removeItem('user_info');
  };

  const handleLoginSuccess = (userData: User) => {
    setIsAuthenticated(true);
    setUser(userData);
    localStorage.setItem('user_info', JSON.stringify(userData));
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  return (
    <div className="App">
      <Header 
        onLogout={handleLogout} 
        isAuthenticated={isAuthenticated}
        user={user}
      />
      
      <main>
        <Layout>
          <Routes>
            <Route 
              path="/" 
              element={
                <HomePage 
                  isAuthenticated={isAuthenticated}
                />
              } 
            />
            <Route 
              path="/products" 
              element={
                <ProductsPage />
              } 
            />
            <Route 
              path="/cart" 
              element={
                <CartPage 
                  isAuthenticated={isAuthenticated}
                  user={user}
                />
              } 
            />
            <Route 
              path="/orders" 
              element={
                isAuthenticated ? (
                  <OrdersPage 
                    isAuthenticated={isAuthenticated}
                    user={user}
                  />
                ) : (
                  <UnauthorizedView />
                )
              } 
            />
            <Route 
              path="/login" 
              element={
                <LoginPage 
                  onLoginSuccess={handleLoginSuccess}
                />
              } 
            />
            <Route 
              path="/register" 
              element={
                <RegisterPage />
              } 
            />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </Layout>
      </main>
    </div>
  );
}

export default App;