import { createTheme } from '@mui/material/styles';
import type { PaletteMode } from '@mui/material';

export function createAppTheme(mode: PaletteMode) {
  return createTheme({
    palette: {
      mode,
      primary: { main: mode === 'light' ? '#2563eb' : '#60a5fa' },
      secondary: { main: mode === 'light' ? '#0ea5e9' : '#38bdf8' },
      background: {
        default: mode === 'light' ? '#f6f7fb' : '#0b1220',
        paper: mode === 'light' ? '#ffffff' : '#0e1526',
      },
    },
    shape: { borderRadius: 12 },
    typography: {
      fontFamily: 'Inter, Roboto, Helvetica, Arial, sans-serif',
      h1: { fontWeight: 800 },
      h2: { fontWeight: 700 },
      h3: { fontWeight: 700 },
    },
    components: {
      MuiCard: {
        styleOverrides: {
          root: {
            transition: 'transform 120ms ease, box-shadow 120ms ease',
            '&:hover': { transform: 'translateY(-2px)' },
          },
        },
      },
      MuiButton: {
        defaultProps: { size: 'medium' },
        styleOverrides: {
          root: { textTransform: 'none', fontWeight: 600, borderRadius: 10 },
        },
      },
    },
  });
}



