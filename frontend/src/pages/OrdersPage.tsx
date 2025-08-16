import React, { useState, useEffect } from 'react';
import { isAxiosError } from 'axios';
import { useLocation, useNavigate } from 'react-router-dom';
import { ordersAPI } from '../services/api';
import type { Order } from '../types';
import { formatMoneyMinor } from '../utils/money';
import Container from '@mui/material/Container';
import Grid from '@mui/material/Grid';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Alert from '@mui/material/Alert';
import StatusChip from '../components/ui/StatusChip';
import Stack from '@mui/material/Stack';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Snackbar from '@mui/material/Snackbar';
import Skeleton from '@mui/material/Skeleton';
import PageHeader from '../components/ui/PageHeader';
import EmptyState from '../components/ui/EmptyState';

const OrdersPage: React.FC = () => {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [confirmId, setConfirmId] = useState<string | null>(null);
  const [snackbarOpen, setSnackbarOpen] = useState(false);
  const [snackbarMessage, setSnackbarMessage] = useState('');
  const location = useLocation();

  const navigate = useNavigate();

  useEffect(() => {
    loadOrders();
  }, []);

  useEffect(() => {
    const st: any = location.state as any;
    if (st?.orderPlaced) {
      setSnackbarMessage('Order placed successfully');
      setSnackbarOpen(true);
      // очистить state, чтобы уведомление не повторялось
      window.history.replaceState({}, document.title, window.location.pathname);
    }
  }, [location.state]);

  const loadOrders = async () => {
    try {
      setLoading(true);
      // Clear previous error on new attempt
      setError('');
      const response = await ordersAPI.getOrders();
      // Normalize empty/204 responses to empty list
      if (response.status === 204) {
        setOrders([]);
        return;
      }
      setOrders((response.data?.orders ?? []) as Order[]);
    } catch (err) {
      // Treat 404 as a valid empty state instead of an error
      if (isAxiosError(err) && err.response?.status === 404) {
        setOrders([]);
        return;
      }
      setError('Failed to load orders');
    } finally {
      setLoading(false);
    }
  };

  const cancelOrder = async (orderId: string) => {
    try {
      await ordersAPI.cancelOrder(orderId);
      await loadOrders();
      setSnackbarMessage('Order cancelled successfully');
      setSnackbarOpen(true);
    } catch (err) {
      setError('Failed to cancel order');
    }
  };

  const handleSnackbarClose = () => setSnackbarOpen(false);

  const statusChip = (status: string) => <StatusChip status={status} />;

  if (loading) {
    return (
      <Container sx={{ mt: 4 }}>
        <PageHeader title="My Orders" />
        <Grid container spacing={3}>
          {Array.from({ length: 3 }).map((_, i) => (
            <Grid key={i} item xs={12}>
              <Card>
                <CardContent>
                  <Skeleton width="30%" />
                  <Skeleton width="20%" />
                  <Skeleton height={80} />
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      </Container>
    );
  }

  return (
    <Container sx={{ mt: 4 }}>
      <PageHeader title="My Orders" />
      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
      {orders.length === 0 ? (
        <EmptyState title="No orders found" description="You haven't placed any orders yet." actionLabel="Start Shopping" onAction={() => navigate('/products')} />
      ) : (
        <Grid container spacing={3}>
          {orders.map(order => (
            <Grid item xs={12} key={order.id}>
              <Card>
                <CardContent>
                  <Stack direction="row" alignItems="flex-start" justifyContent="space-between" sx={{ mb: 2 }}>
                    <div>
                      <Typography variant="h6">Order #{order.id}</Typography>
                      <Typography color="text.secondary">Placed on {new Date(order.created_at).toLocaleDateString()}</Typography>
                      <Stack direction="row" spacing={1} sx={{ mt: 1 }}>{statusChip(order.status)}</Stack>
                    </div>
                    <div style={{ textAlign: 'right' as const }}>
                      <Typography variant="h6">{formatMoneyMinor((order as any).total_amount?.amount, (order as any).total_amount?.currency)}</Typography>
                      {order.status === 'PENDING' && (
                        <Button color="error" sx={{ mt: 1 }} onClick={() => setConfirmId(order.id)}>Cancel Order</Button>
                      )}
                    </div>
                  </Stack>
                  <Typography variant="subtitle1" sx={{ mb: 1 }}>Items:</Typography>
                  <Stack>
                    {order.items.map(item => (
                      <Stack key={item.id} direction="row" justifyContent="space-between" sx={{ py: 1, borderBottom: '1px solid', borderColor: 'divider' }}>
                        <div>
                          <Typography fontWeight={600}>{item.product_name}</Typography>
                          <Typography color="text.secondary">{formatMoneyMinor((item as any).price?.amount, (item as any).price?.currency)} × {item.quantity}</Typography>
                        </div>
                        <Typography fontWeight={700}>{formatMoneyMinor((item as any).total?.amount, (item as any).total?.currency)}</Typography>
                      </Stack>
                    ))}
                  </Stack>
                  <Typography sx={{ mt: 2 }}><strong>Shipping Address:</strong></Typography>
                  <Typography color="text.secondary">{order.shipping_address}</Typography>
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}

      <Dialog open={!!confirmId} onClose={() => setConfirmId(null)}>
        <DialogTitle>Cancel order?</DialogTitle>
        <DialogContent>
          <Typography>Are you sure you want to cancel this order?</Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setConfirmId(null)}>No</Button>
          <Button color="error" onClick={() => { if (confirmId) cancelOrder(confirmId); setConfirmId(null); }}>Yes, cancel</Button>
        </DialogActions>
      </Dialog>

      <Snackbar open={snackbarOpen} autoHideDuration={3000} onClose={handleSnackbarClose} anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}>
        <Alert onClose={handleSnackbarClose} severity="success" variant="filled" sx={{ width: '100%' }}>
          {snackbarMessage || 'Success'}
        </Alert>
      </Snackbar>
    </Container>
  );
};

export default OrdersPage;