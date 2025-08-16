import React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';
import Link from '@mui/material/Link';

const Header = () => {
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
          <Button component={RouterLink} to="/orders">Orders</Button>
        </Stack>
      </Toolbar>
    </AppBar>
  );
};

export default Header;