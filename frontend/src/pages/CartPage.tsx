import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { ordersAPI } from '../services/api';
import type { CartItem, PaymentDetails, CreateOrderRequest } from '../types';
import { isAxiosError } from 'axios';
import { formatMoneyMinor, sumMinorWithSameCurrency } from '../utils/money';
import Container from '@mui/material/Container';
import Grid from '@mui/material/Grid';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import Alert from '@mui/material/Alert';
import Stack from '@mui/material/Stack';
import Snackbar from '@mui/material/Snackbar';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';

type OrderForm = {
  shipping_address: string;
  payment_method: string;
  payment_details: PaymentDetails;
};

function formatMinor(amountMinor: number, currency?: string): string {
  return formatMoneyMinor(amountMinor, currency);
}

const CartPage: React.FC = () => {
  const [cartItems, setCartItems] = useState<CartItem[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>('');
  const [orderForm, setOrderForm] = useState<OrderForm>({
    shipping_address: '',
    payment_method: 'credit_card', 
    payment_details: {
      card_number: '',
      card_holder: '',
      expiry_month: '',
      expiry_year: '',
      cvv: '',
    },
  });
  const [snackbarOpen, setSnackbarOpen] = useState(false);

  const navigate = useNavigate();

  useEffect(() => {
    loadCart();
  }, []);

  const loadCart = () => {
    const raw = JSON.parse(localStorage.getItem('cart') || '[]') as any[];
    // Normalize legacy entries to unified Money shape
    const cart: CartItem[] = raw.map((item) => ({
      product_id: item.product_id,
      product_name: item.product_name,
      price: {
        amount: typeof item.price?.amount === 'number' ? item.price.amount : (typeof item.price_minor === 'number' ? item.price_minor : 0),
        currency: typeof item.price?.currency === 'string' ? item.price.currency : (typeof item.currency === 'string' ? item.currency : 'USD'),
      },
      quantity: Number(item.quantity || 1),
    }));
    setCartItems(cart);
    localStorage.setItem('cart', JSON.stringify(cart));
  };

  const updateQuantity = (productId: string, newQuantity: number) => {
    if (newQuantity <= 0) {
      removeItem(productId);
      return;
    }

    const updatedCart = cartItems.map(item =>
      item.product_id === productId
        ? { ...item, quantity: newQuantity }
        : item
    );
    
    setCartItems(updatedCart);
    localStorage.setItem('cart', JSON.stringify(updatedCart));
  };

  const removeItem = (productId: string) => {
    const updatedCart = cartItems.filter(item => item.product_id !== productId);
    setCartItems(updatedCart);
    localStorage.setItem('cart', JSON.stringify(updatedCart));
  };

  const getTotalAmount = (): string => {
    const aggregated = sumMinorWithSameCurrency(
      cartItems.map((item) => ({ amountMinor: Number(item.price?.amount || 0) * item.quantity, currency: item.price?.currency }))
    );
    return formatMoneyMinor(aggregated.amountMinor, aggregated.currency).replace(/[^0-9.,-]/g, '');
  };

  const schema = z.object({
    shipping_address: z.string().min(5, 'Enter full address'),
    card_holder: z.string().min(2, 'Enter name'),
    card_number: z.string().min(12, 'Card number too short'),
    expiry_month: z.string().min(2, 'MM'),
    expiry_year: z.string().min(2, 'YY'),
    cvv: z.string().min(3, 'CVV'),
  });
  type FormValues = z.infer<typeof schema>;
  const { register: rhfRegister, handleSubmit, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
    mode: 'onChange',
    defaultValues: {
      shipping_address: '',
      card_holder: '',
      card_number: '',
      expiry_month: '',
      expiry_year: '',
      cvv: '',
    },
  });

  const onSubmit = async (data: FormValues) => {
    setError('');
    setLoading(true);

    try {
      const orderData: CreateOrderRequest = {
        user_id: 'dev-user-1', // Temporary user_id for development
        items: cartItems.map(item => ({
          product_id: item.product_id,
          quantity: item.quantity,
        })),
        shipping_address: data.shipping_address,
        payment_method: orderForm.payment_method,
        payment_details: {
          card_holder: data.card_holder,
          card_number: data.card_number,
          expiry_month: data.expiry_month,
          expiry_year: data.expiry_year,
          cvv: data.cvv,
        },
      };

            try {
        const response = await ordersAPI.createOrder(orderData);
      } catch (error) {
        throw error;
      }
      
      // Clear cart
      localStorage.removeItem('cart');
      setCartItems([]);
      // Show success notification before redirect
      setSnackbarOpen(true);
      setTimeout(() => navigate('/orders', { state: { orderPlaced: true } }), 1500);
    } catch (err: unknown) {
      const message = isAxiosError(err)
        ? (err.response?.data as any)?.error ?? err.message
        : err instanceof Error
          ? err.message
          : 'Failed to place order';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  if (cartItems.length === 0) {
    return (
      <Container sx={{ mt: 4 }}>
        <Card>
          <CardContent sx={{ textAlign: 'center', py: 6 }}>
            <Typography variant="h5" gutterBottom>Your cart is empty</Typography>
            <Typography color="text.secondary" sx={{ mb: 2 }}>Add some products to your cart to get started!</Typography>
            <Button variant="contained" onClick={() => navigate('/products')}>Browse Products</Button>
          </CardContent>
        </Card>
      </Container>
    );
  }

  return (
    <Container sx={{ mt: 4 }}>
      <Typography variant="h4" sx={{ mb: 2 }}>Shopping Cart</Typography>
      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
      <Grid container spacing={3}>
        <Grid item xs={12} md={8}>
          <Card>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 2 }}>Cart Items</Typography>
              <Stack spacing={2}>
                {cartItems.map(item => (
                  <Stack key={item.product_id} direction="row" alignItems="center" justifyContent="space-between" sx={{ borderBottom: '1px solid', borderColor: 'divider', pb: 2 }}>
                    <div>
                      <Typography fontWeight={600}>{item.product_name}</Typography>
                      <Typography color="text.secondary">{formatMinor(item.price?.amount, item.price?.currency)} each</Typography>
                    </div>
                    <Stack direction="row" spacing={1} alignItems="center">
                      <Button variant="outlined" onClick={() => updateQuantity(item.product_id, item.quantity - 1)}>-</Button>
                      <Typography sx={{ width: 32, textAlign: 'center' }}>{item.quantity}</Typography>
                      <Button variant="outlined" onClick={() => updateQuantity(item.product_id, item.quantity + 1)}>+</Button>
                      <Button color="error" onClick={() => removeItem(item.product_id)}>Remove</Button>
                    </Stack>
                  </Stack>
                ))}
              </Stack>
              <Typography variant="h6" sx={{ mt: 2, textAlign: 'right' }}>Total: ${getTotalAmount()}</Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={4}>
          <Card>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 2 }}>Checkout</Typography>
              <form onSubmit={handleSubmit(onSubmit)}>
                <TextField id="shipping_address" label="Shipping Address" {...rhfRegister('shipping_address')} error={!!errors.shipping_address} helperText={errors.shipping_address?.message} fullWidth multiline rows={3} sx={{ mb: 2 }} />
                <TextField id="card_holder" label="Card Holder Name" {...rhfRegister('card_holder')} error={!!errors.card_holder} helperText={errors.card_holder?.message} fullWidth sx={{ mb: 2 }} />
                <TextField id="card_number" label="Card Number" {...rhfRegister('card_number')} error={!!errors.card_number} helperText={errors.card_number?.message} placeholder="1234 5678 9012 3456" fullWidth sx={{ mb: 2 }} />
                <Stack direction="row" spacing={2} sx={{ mb: 2 }}>
                  <TextField id="expiry_month" label="Month" {...rhfRegister('expiry_month')} error={!!errors.expiry_month} helperText={errors.expiry_month?.message} placeholder="MM" fullWidth />
                  <TextField id="expiry_year" label="Year" {...rhfRegister('expiry_year')} error={!!errors.expiry_year} helperText={errors.expiry_year?.message} placeholder="YY" fullWidth />
                  <TextField id="cvv" label="CVV" {...rhfRegister('cvv')} error={!!errors.cvv} helperText={errors.cvv?.message} placeholder="123" fullWidth />
                </Stack>
                <Button type="submit" variant="contained" fullWidth disabled={loading}>
                  {loading ? 'Placing Order...' : `Place Order - $${getTotalAmount()}`}
                </Button>
              </form>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
      <Snackbar open={snackbarOpen} autoHideDuration={2500} onClose={() => setSnackbarOpen(false)} anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}>
        <Alert onClose={() => setSnackbarOpen(false)} severity="success" variant="filled" sx={{ width: '100%' }}>
          Order placed successfully
        </Alert>
      </Snackbar>
    </Container>
  );
};

export default CartPage;