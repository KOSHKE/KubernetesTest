import React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Grid from '@mui/material/Grid';
import Box from '@mui/material/Box';

interface HomePageProps {
  isAuthenticated: boolean;
}

const HomePage: React.FC<HomePageProps> = ({ isAuthenticated }) => {
  return (
    <Container>
      <Box sx={{
        mt: 8,
        mb: 4,
        p: { xs: 3, md: 6 },
        borderRadius: 4,
        background: (t) => `linear-gradient(135deg, ${t.palette.primary.main}22, ${t.palette.secondary.main}22)`
      }}>
        <Typography variant="overline" color="primary">Microservices E‑Commerce</Typography>
        <Typography variant="h3" fontWeight={800}>Order System</Typography>
        <Typography sx={{ fontSize: 18, my: 2, color: 'text.secondary' }}>
          A microservices-based e-commerce platform built with Kubernetes, Go, and React
        </Typography>
        <Button component={RouterLink} to="/products" variant="contained" sx={{ mr: 1 }}>
          Browse Products
        </Button>
        {isAuthenticated ? (
          <Button component={RouterLink} to="/orders" variant="outlined">
            View Orders
          </Button>
        ) : (
          <>
            <Button component={RouterLink} to="/login" variant="outlined" sx={{ mr: 1 }}>
              Sign In
            </Button>
            <Button component={RouterLink} to="/register" variant="contained">
              Register
            </Button>
          </>
        )}
      </Box>

      <Typography variant="h5" sx={{ mt: 6, mb: 2 }}>Architecture Features</Typography>
      <Grid container spacing={3}>
        {[
          { title: '🏗️ Microservices', desc: 'Domain-driven design with separate services for users, orders, inventory, and payments' },
          { title: '🚀 Kubernetes', desc: 'Container orchestration with deployments, services, and ingress controllers' },
          { title: '⚡ gRPC', desc: 'High-performance inter-service communication with protocol buffers' },
          { title: '🛒 Ordering', desc: 'Cart, checkout, and order lifecycle demo' },
        ].map((item) => (
          <Grid item xs={12} md={6} lg={3} key={item.title}>
            <Card>
              <CardContent>
                <Typography variant="h6">{item.title}</Typography>
                <Typography color="text.secondary">{item.desc}</Typography>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>
    </Container>
  );
};

export default HomePage;