import React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';
import Link from '@mui/material/Link';

interface HeaderProps {
  onLogout: () => void;
  isAuthenticated: boolean;
  user?: { id: string; email: string; first_name: string; last_name: string } | null;
}

const Header: React.FC<HeaderProps> = ({ onLogout, isAuthenticated, user }) => {
  return (
    <AppBar position="static" color="inherit" elevation={1} sx={{ mb: 3 }}>
      <Toolbar>
        <Typography variant="h6" sx={{ flexGrow: 1, fontWeight: 700 }}>
          <Link component={RouterLink} to="/" underline="none" color="primary.main">
            Order System
          </Link>
        </Typography>
        <Stack direction="row" spacing={1} alignItems="center">
          <Button component={RouterLink} to="/products" color="primary">Products</Button>
          <Button component={RouterLink} to="/cart">Cart</Button>
          {isAuthenticated && (
            <Button component={RouterLink} to="/orders">Orders</Button>
          )}
          {isAuthenticated ? (
            <Stack direction="row" spacing={1} alignItems="center">
              {user && (
                <Typography variant="body2" color="text.secondary" sx={{ mr: 1 }}>
                  Welcome, {user.first_name}!
                </Typography>
              )}
              <Button onClick={onLogout} color="secondary">Logout</Button>
            </Stack>
          ) : (
            <>
              <Button component={RouterLink} to="/login" color="primary" variant="outlined">Sign In</Button>
              <Button component={RouterLink} to="/register" color="primary" variant="contained">Register</Button>
            </>
          )}
        </Stack>
      </Toolbar>
    </AppBar>
  );
};

export default Header;