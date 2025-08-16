import React from 'react';
import Chip from '@mui/material/Chip';

type Props = { status: string };

const StatusChip: React.FC<Props> = ({ status }) => {
  const map: Record<string, { label: string; color: 'default' | 'success' | 'warning' | 'info' | 'error' }> = {
    PENDING: { label: 'Pending', color: 'warning' },
    CONFIRMED: { label: 'Confirmed', color: 'info' },
    PROCESSING: { label: 'Processing', color: 'info' },
    SHIPPED: { label: 'Shipped', color: 'success' },
    DELIVERED: { label: 'Delivered', color: 'success' },
    CANCELLED: { label: 'Cancelled', color: 'error' },
  };
  const cfg = map[status] || { label: status, color: 'default' as const };
  return <Chip label={cfg.label} color={cfg.color} size="small" />;
};

export default StatusChip;


