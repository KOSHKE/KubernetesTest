import React, { useState } from 'react';
import {
  Card,
  CardContent,
  CardMedia,
  Typography,
  Button,
  Box,
  Chip,
  Rating,
  Grow,
  Stack
} from '@mui/material';
import {
  ShoppingCart as CartIcon
} from '@mui/icons-material';
import type { Product } from '../types';

interface ProductCardProps {
  product: Product;
  onAddToCart: (product: Product) => void;
}

const ProductCard: React.FC<ProductCardProps> = ({ 
  product, 
  onAddToCart
}) => {
  const [rating] = useState(4.2 + Math.random() * 0.8); // Random rating for demo

  const formatMoney = (m?: { amount: number; currency: string }) => {
    if (!m) return '';
    const major = (m.amount / 100).toFixed(2);
    return `${major} ${m.currency}`;
  };

  const handleAddToCart = () => {
    onAddToCart(product);
  };

  const isOutOfStock = product.stock_quantity === 0;
  const price = formatMoney((product as any).price as any);

  return (
    <Card
      sx={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        position: 'relative',
        transition: 'all 0.2s ease',
        '&:hover': {
          transform: 'translateY(-2px)',
          boxShadow: '0 4px 12px rgba(0,0,0,0.1)',
        },
        cursor: 'pointer',
        borderRadius: 1,
      }}
    >
      {/* Stock Badge */}
      {isOutOfStock && (
        <Chip
          label="Out of Stock"
          color="error"
          size="small"
          sx={{
            position: 'absolute',
            top: 8,
            left: 8,
            zIndex: 2,
            backgroundColor: 'rgba(244,67,54,0.9)',
            color: 'white',
            fontWeight: 'bold',
            fontSize: '0.7rem',
            height: 20,
          }}
        />
      )}

      {/* Product Image */}
      <CardMedia
        component="img"
        height="140"
        image={product.image_url || '/placeholder-product.jpg'}
        alt={product.name}
        sx={{
          objectFit: 'cover',
          borderBottom: '1px solid',
          borderColor: 'divider',
        }}
      />

      {/* Card Content */}
      <CardContent sx={{ flexGrow: 1, p: 1, display: 'flex', flexDirection: 'column' }}>
        {/* Product Name - Bold and Compact */}
        <Typography
          variant="body2"
          component="h3"
          sx={{
            fontWeight: 700,
            mb: 0,
            lineHeight: 1.2,
            height: '2.4em',
            overflow: 'hidden',
            display: '-webkit-box',
            WebkitLineClamp: 2,
            WebkitBoxOrient: 'vertical',
            fontSize: '0.85rem',
            color: 'text.primary',
          }}
        >
          {product.name}
        </Typography>

        {/* Rating - Very Compact, no margin from name */}
        <Stack direction="row" alignItems="center" spacing={0.5} sx={{ mb: 0.5 }}>
          <Rating
            value={rating}
            precision={0.1}
            size="small"
            readOnly
            sx={{ 
              '& .MuiRating-iconFilled': { color: '#FFD700' },
              fontSize: '0.8rem'
            }}
          />
          <Typography variant="caption" color="text.secondary" sx={{ fontSize: '0.7rem' }}>
            ({rating.toFixed(1)})
          </Typography>
        </Stack>

        {/* Price - Prominent */}
        <Typography
          variant="h6"
          component="div"
          sx={{
            fontWeight: 700,
            color: 'primary.main',
            fontSize: '1.1rem',
            mb: 0.5,
          }}
        >
          {price}
        </Typography>

        {/* Stock Info - Compact */}
        <Typography
          variant="caption"
          color={isOutOfStock ? 'error.main' : 'success.main'}
          sx={{ 
            fontWeight: 500, 
            fontSize: '0.7rem',
            mb: 0.5,
          }}
        >
          {isOutOfStock ? 'Out of Stock' : `${product.stock_quantity} in stock`}
        </Typography>

        {/* Add to Cart Button - Compact */}
        <Button
          variant="contained"
          fullWidth
          size="small"
          onClick={handleAddToCart}
          disabled={isOutOfStock}
          startIcon={<CartIcon />}
          sx={{
            py: 0.4,
            px: 1.5,
            fontWeight: 600,
            fontSize: '0.75rem',
            textTransform: 'none',
            borderRadius: 1,
            transition: 'all 0.2s ease',
            minHeight: 26,
            mt: 'auto',
            '&:hover': {
              transform: 'translateY(-1px)',
              boxShadow: '0 2px 8px rgba(0,0,0,0.15)',
            },
            '&:disabled': {
              backgroundColor: 'grey.300',
              color: 'grey.600',
            },
          }}
        >
          {isOutOfStock 
            ? 'Out of Stock' 
            : 'Add to Cart'
          }
        </Button>
      </CardContent>
    </Card>
  );
};

export default ProductCard;
