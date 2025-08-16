import React from 'react';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';

type EmptyStateProps = {
  title: string;
  description?: string;
  actionLabel?: string;
  onAction?: () => void;
};

const EmptyState: React.FC<EmptyStateProps> = ({ title, description, actionLabel, onAction }) => {
  return (
    <Card>
      <CardContent style={{ textAlign: 'center', padding: '48px' }}>
        <Typography variant="h6" gutterBottom>{title}</Typography>
        {description && <Typography color="text.secondary" sx={{ mb: 2 }}>{description}</Typography>}
        {actionLabel && onAction && (
          <Button variant="contained" onClick={onAction}>{actionLabel}</Button>
        )}
      </CardContent>
    </Card>
  );
};

export default EmptyState;


