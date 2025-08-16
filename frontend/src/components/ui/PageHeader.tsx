import React from 'react';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';

type PageHeaderProps = {
  title: string;
  subtitle?: string;
  actions?: React.ReactNode;
};

const PageHeader: React.FC<PageHeaderProps> = ({ title, subtitle, actions }) => {
  return (
    <Box sx={{ mb: 2, display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 2 }}>
      <Box>
        <Typography variant="h4" fontWeight={800}>{title}</Typography>
        {subtitle && (
          <Typography color="text.secondary">{subtitle}</Typography>
        )}
      </Box>
      {actions}
    </Box>
  );
};

export default PageHeader;


