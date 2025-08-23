import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { authService } from '../services/auth';
import Container from '@mui/material/Container';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import Alert from '@mui/material/Alert';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import Grid from '@mui/material/Grid';

const RegisterPage: React.FC = () => {
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    first_name: '',
    last_name: '',
    phone: '',
  });
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const navigate = useNavigate();

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    
    // Special handling for phone field - only allow digits and +
    if (name === 'phone') {
      const phoneValue = value.replace(/[^\d+]/g, '');
      setFormData({
        ...formData,
        [name]: phoneValue,
      });
    } else {
      setFormData({
        ...formData,
        [name]: value,
      });
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');

    try {
      const requestData = {
        email: formData.email,
        password: formData.password,
        first_name: formData.first_name,
        last_name: formData.last_name,
        Phone: formData.phone,
      };
      
      await authService.register(requestData);
      // After successful registration, redirect to login
      navigate('/login', { state: { message: 'Account created successfully! Please sign in.' } });
    } catch (err: any) {
      setError(err.message || 'Registration failed. Please check your information and try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Container maxWidth="md" sx={{ mt: 8 }}>
      <Card sx={{ p: 4, boxShadow: 3 }}>
        <CardContent>
          <Box sx={{ textAlign: 'center', mb: 4 }}>
            <Typography variant="h4" component="h1" gutterBottom fontWeight="bold">
              Create Your Account
            </Typography>
            <Typography variant="body1" color="text.secondary">
              Join us and start shopping today
            </Typography>
          </Box>

          <form onSubmit={handleSubmit}>
            <Stack spacing={3} sx={{ alignItems: 'flex-start' }}>
                             <Stack direction="row" spacing={2} sx={{ width: '100%' }}>
                 <TextField
                   id="firstName"
                   name="first_name"
                   label="First Name"
                   type="text"
                   required
                   sx={{ flex: 1 }}
                   value={formData.first_name}
                   onChange={handleChange}
                   placeholder="Enter your first name"
                   variant="outlined"
                   size="medium"
                 />
                 <TextField
                   id="lastName"
                   name="last_name"
                   label="Last Name"
                   type="text"
                   required
                   sx={{ flex: 1 }}
                   value={formData.last_name}
                   onChange={handleChange}
                   placeholder="Enter your last name"
                   variant="outlined"
                   size="medium"
                 />
               </Stack>

              <TextField
                id="email"
                name="email"
                label="Email Address"
                type="email"
                autoComplete="email"
                required
                fullWidth
                value={formData.email}
                onChange={handleChange}
                placeholder="Enter your email"
                variant="outlined"
                size="medium"
              />

                             <TextField
                 id="phone"
                 name="phone"
                 label="Phone Number (Optional)"
                 type="tel"
                 fullWidth
                 value={formData.phone}
                 onChange={handleChange}
                 placeholder="Enter your phone number"
                 variant="outlined"
                 size="medium"
               />
              
              <TextField
                id="password"
                name="password"
                label="Password"
                type="password"
                autoComplete="new-password"
                required
                fullWidth
                value={formData.password}
                onChange={handleChange}
                placeholder="Enter your password"
                variant="outlined"
                size="medium"
                helperText="Password must be at least 6 characters long"
              />

              {error && (
                <Alert severity="error" variant="filled">
                  {error}
                </Alert>
              )}

              <Button
                type="submit"
                variant="contained"
                size="large"
                fullWidth
                disabled={isLoading}
                sx={{ 
                  py: 1.5,
                  fontSize: '1rem',
                  fontWeight: 600,
                  mt: 2
                }}
              >
                {isLoading ? 'Creating Account...' : 'Create Account'}
              </Button>
            </Stack>
          </form>

          <Box sx={{ mt: 4, textAlign: 'center' }}>
            <Typography variant="body2" color="text.secondary">
              Already have an account?{' '}
              <Link
                component="button"
                variant="body2"
                onClick={() => navigate('/login')}
                sx={{ 
                  textDecoration: 'none',
                  fontWeight: 600
                }}
              >
                Sign in here
              </Link>
            </Typography>
          </Box>

          
        </CardContent>
      </Card>
    </Container>
  );
};

export default RegisterPage;
