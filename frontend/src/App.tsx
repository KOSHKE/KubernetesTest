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
import UnauthorizedPage from './components/UnauthorizedPage';
import { authService } from './services/auth';

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Check if user is authenticated on app load
    const checkAuth = () => {
      const authenticated = authService.isAuthenticated();
      setIsAuthenticated(authenticated);
      setIsLoading(false);
    };

    checkAuth();
  }, []);

  const handleLogout = async () => {
    await authService.logout();
    setIsAuthenticated(false);
  };

  const handleLoginSuccess = () => {
    setIsAuthenticated(true);
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
                />
              } 
            />
            <Route 
              path="/orders" 
              element={
                isAuthenticated ? (
                  <OrdersPage 
                    isAuthenticated={isAuthenticated}
                  />
                ) : (
                  <UnauthorizedPage />
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