import React from 'react';
import ReactDOM from 'react-dom/client';
import { QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter } from 'react-router-dom';
import App from './App';
import recipickIconUrl from './assets/recipick-icon.svg';
import { AuthProvider } from './auth/AuthContext';
import { queryClient } from './lib/queryClient';
import './styles.css';

const updateFavicon = () => {
  let iconLink = document.querySelector('link[rel="icon"]') as HTMLLinkElement | null;
  if (!iconLink) {
    iconLink = document.createElement('link');
    iconLink.rel = 'icon';
    iconLink.type = 'image/svg+xml';
    document.head.appendChild(iconLink);
  }
  iconLink.href = recipickIconUrl;
};

updateFavicon();

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <BrowserRouter>
          <App />
        </BrowserRouter>
      </AuthProvider>
    </QueryClientProvider>
  </React.StrictMode>,
);
