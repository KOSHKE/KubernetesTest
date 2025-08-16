import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import App from './App';
import './index.css';
import { ThemeProvider, CssBaseline } from '@mui/material';
import { createAppTheme } from './theme';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
const prefersDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
const theme = createAppTheme(prefersDark ? 'dark' : 'light');
const queryClient = new QueryClient();

const container = document.getElementById('root');
if (!container) {
  throw new Error('Root container with id "root" not found');
}
const root = ReactDOM.createRoot(container);
root.render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <BrowserRouter>
          <App />
        </BrowserRouter>
      </ThemeProvider>
    </QueryClientProvider>
  </React.StrictMode>
);