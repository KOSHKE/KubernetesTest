import React from 'react';
import { useNavigate } from 'react-router-dom';
import Container from '@mui/material/Container';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import LockIcon from '@mui/icons-material/Lock';
import Stack from '@mui/material/Stack';

interface UnauthorizedViewProps {
  title?: string;
  message?: string;
  showLoginButton?: boolean;
  showRegisterButton?: boolean;
}

const UnauthorizedView: React.FC<UnauthorizedViewProps> = ({
  title = "Authentication Required",
  message = "Please sign in to access this page",
  showLoginButton = true,
  showRegisterButton = true
}) => {
  const navigate = useNavigate();

  return (
    <Container maxWidth="md" sx={{ mt: 8, mb: 8 }}>
      <Card sx={{ p: 6, boxShadow: 3, textAlign: 'center' }}>
        <CardContent>
          {/* Icon */}
          <Box sx={{ mb: 4 }}>
            <LockIcon 
              sx={{ 
                fontSize: 80, 
                color: 'primary.main',
                opacity: 0.7 
              }} 
            />
          </Box>

          {/* Title */}
          <Typography 
            variant="h3" 
            component="h1" 
            gutterBottom 
            fontWeight="bold"
            color="primary"
            sx={{ mb: 2 }}
          >
            {title}
          </Typography>

          {/* Message */}
          <Typography 
            variant="h6" 
            color="text.secondary" 
            sx={{ mb: 6, maxWidth: 500, mx: 'auto' }}
          >
            {message}
          </Typography>

          {/* Action Buttons */}
          <Stack 
            direction={{ xs: 'column', sm: 'row' }} 
            spacing={2} 
            justifyContent="center"
            sx={{ mb: 4 }}
          >
            {showLoginButton && (
              <Button
                variant="contained"
                size="large"
                onClick={() => navigate('/login')}
                sx={{ 
                  py: 1.5,
                  px: 4,
                  fontSize: '1.1rem',
                  fontWeight: 600,
                  minWidth: 140
                }}
              >
                Sign In
              </Button>
            )}
            
            {showRegisterButton && (
              <Button
                variant="outlined"
                size="large"
                onClick={() => navigate('/register')}
                sx={{ 
                  py: 1.5,
                  px: 4,
                  fontSize: '1.1rem',
                  fontWeight: 600,
                  minWidth: 140
                }}
              >
                Create Account
              </Button>
            )}
          </Stack>


        </CardContent>
      </Card>
    </Container>
  );
};

export default UnauthorizedView;
