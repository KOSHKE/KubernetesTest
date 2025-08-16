import React from 'react';
import Container from '@mui/material/Container';
import Box from '@mui/material/Box';

type Props = { children: React.ReactNode };

const Layout: React.FC<Props> = ({ children }) => {
  return (
    <>
      <Container maxWidth="lg" sx={{ pb: 6 }}>
        {children}
      </Container>
      <Box component="footer" sx={{ py: 4, mt: 8, borderTop: '1px solid', borderColor: 'divider', textAlign: 'center', color: 'text.secondary' }}>
        © {new Date().getFullYear()} Order System
      </Box>
    </>
  );
};

export default Layout;


