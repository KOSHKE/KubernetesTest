import React, { useState, useEffect } from 'react';
import { productsAPI } from '../services/api';
import { useQuery } from '@tanstack/react-query';
import type { Product, Category, CartItem } from '../types';
import Container from '@mui/material/Container';
import Grid from '@mui/material/Grid';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import TextField from '@mui/material/TextField';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import InputLabel from '@mui/material/InputLabel';
import FormControl from '@mui/material/FormControl';
import Button from '@mui/material/Button';
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';
import Stack from '@mui/material/Stack';
import Skeleton from '@mui/material/Skeleton';
import PageHeader from '../components/ui/PageHeader';
import EmptyState from '../components/ui/EmptyState';
import ProductCard from '../components/ProductCard';

interface ProductsPageProps {
  // No props needed anymore
}

const ProductsPage: React.FC<ProductsPageProps> = () => {
  const [categories, setCategories] = useState<Category[]>([]);
  const [selectedCategory, setSelectedCategory] = useState('');
  const [searchTerm, setSearchTerm] = useState('');
  const [snackbarOpen, setSnackbarOpen] = useState(false);

  useEffect(() => {
    loadCategories();
  }, []);

  const { data: productsData, isFetching, error: queryError } = useQuery({
    queryKey: ['products', selectedCategory],
    queryFn: async () => {
      const params: { category_id?: string } = {};
      if (selectedCategory) params.category_id = selectedCategory;
      try {
        const response = await productsAPI.getProducts(params as any);
        const fetched = (response.data.data.products || []) as Product[];
        return fetched;
      } catch (error) {
        throw error;
      }
    },
    enabled: true,
    retry: 1,
    retryDelay: 1000,
  });

  // Filter products by search term
  const filteredProducts = React.useMemo(() => {
    if (!productsData) {
      return [];
    }
    const normalize = (value: unknown) => (value ?? '').toString().toLowerCase();
    const query = normalize(searchTerm);
    if (!query) {
      return productsData;
    }
    const filtered = productsData.filter((p: Product) => {
      const name = normalize(p.name);
      const description = normalize((p as any).description);
      return name.includes(query) || description.includes(query);
    });
    return filtered;
  }, [productsData, searchTerm]);



  const loadCategories = async () => {
    try {
      const response = await productsAPI.getCategories();
      setCategories((response.data.data.categories || []) as Category[]);
    } catch (err) {
      // Don't throw error - let products load anyway
    }
  };

  const formatMoney = (m?: { amount: number; currency: string }) => {
    if (!m) return '';
    const major = (m.amount / 100).toFixed(2);
    return `${major} ${m.currency}`;
  };

  const addToCart = (product: Product) => {
    const existingCart = JSON.parse(localStorage.getItem('cart') || '[]') as CartItem[];
    const existingItem = existingCart.find((item: CartItem) => item.product_id === product.id);
    if (existingItem) {
      existingItem.quantity += 1;
    } else {
      existingCart.push({
        product_id: product.id,
        product_name: product.name,
        price: { amount: (product.price as any)?.amount ?? 0, currency: (product.price as any)?.currency ?? 'USD' },
        quantity: 1,
      });
    }
    localStorage.setItem('cart', JSON.stringify(existingCart));
    setSnackbarOpen(true);
  };

  const handleSnackbarClose = (_e?: unknown, reason?: string) => {
    if (reason === 'clickaway') return;
    setSnackbarOpen(false);
  };

  return (
    <>
    <Container sx={{ mt: 2 }}>
      <PageHeader title="Products" />
      {queryError && <Alert severity="error" sx={{ mb: 2 }}>{queryError.message}</Alert>}

      <Card sx={{ mb: 4, borderRadius: 3, boxShadow: 2 }}>
        <CardContent sx={{ p: 3 }}>
          <Stack 
            direction={{ xs: 'column', sm: 'row' }} 
            spacing={3} 
            alignItems="center"
            sx={{ 
              '& .MuiFormControl-root': { 
                minWidth: { xs: '100%', sm: 200 },
                '& .MuiOutlinedInput-root': {
                  borderRadius: 2,
                }
              }
            }}
          >
            <FormControl>
              <InputLabel id="category-label">Category</InputLabel>
              <Select
                labelId="category-label"
                id="category"
                label="Category"
                value={selectedCategory}
                onChange={(e) => setSelectedCategory(e.target.value)}
                sx={{ borderRadius: 2 }}
              >
                <MenuItem value="">All Categories</MenuItem>
                {categories.map(category => (
                  <MenuItem key={category.id} value={category.id}>{category.name}</MenuItem>
                ))}
              </Select>
            </FormControl>
            <TextField
              id="search"
              placeholder="Search products..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              fullWidth
              label="Search Products"
              sx={{ 
                '& .MuiOutlinedInput-root': {
                  borderRadius: 2,
                }
              }}
            />
          </Stack>
        </CardContent>
      </Card>

      <Grid container spacing={3}>
        {isFetching ? (
          Array.from({ length: 8 }).map((_, i) => (
            <Grid key={i} item xs={12} sm={6} md={4} lg={3} xl={2}>
              <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
                <Skeleton variant="rectangular" height={140} />
                <CardContent sx={{ flexGrow: 1, p: 1 }}>
                  <Skeleton width="90%" height={20} sx={{ mb: 0 }} />
                  <Skeleton width="60%" height={16} sx={{ mb: 0.5 }} />
                  <Skeleton width="40%" height={20} sx={{ mb: 0.5 }} />
                  <Skeleton width="50%" height={14} sx={{ mb: 0.5 }} />
                  <Skeleton width="100%" height={26} />
                </CardContent>
              </Card>
            </Grid>
          ))
        ) : (!filteredProducts || filteredProducts.length === 0) ? (
          <Grid item xs={12}>
            <EmptyState title="No products found" description="Try changing filters or search query" />
          </Grid>
        ) : (
          filteredProducts.map((product: Product) => (
            <Grid item xs={12} sm={6} md={4} lg={3} xl={2} key={product.id}>
              <ProductCard 
                product={product} 
                onAddToCart={addToCart} 
              />
            </Grid>
          ))
        )}
      </Grid>
    </Container>
    
    <Snackbar 
      open={snackbarOpen} 
      autoHideDuration={3000} 
      onClose={handleSnackbarClose}
      anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
    >
      <Alert onClose={handleSnackbarClose} severity="success" variant="filled" sx={{ width: '100%' }}>
        Product added to cart
      </Alert>
    </Snackbar>
    </>
  );
};

export default ProductsPage;