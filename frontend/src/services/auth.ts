import api from './api';

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  Phone?: string;
}

export interface LoginResponse {
  user: {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
  };
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface RegisterResponse {
  user: {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
  };
  message: string;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface RefreshTokenResponse {
  access_token: string;
  expires_in: number;
}

export interface LogoutRequest {
  refresh_token: string;
}

class AuthService {
  private accessTokenKey = 'access_token';
  private refreshTokenKey = 'refresh_token';

  // Store tokens in localStorage
  setTokens(accessToken: string, refreshToken: string): void {
    localStorage.setItem(this.accessTokenKey, accessToken);
    localStorage.setItem(this.refreshTokenKey, refreshToken);
  }

  // Get access token
  getAccessToken(): string | null {
    return localStorage.getItem(this.accessTokenKey);
  }

  // Get refresh token
  getRefreshToken(): string | null {
    return localStorage.getItem(this.refreshTokenKey);
  }

  // Clear tokens
  clearTokens(): void {
    localStorage.removeItem(this.accessTokenKey);
    localStorage.removeItem(this.refreshTokenKey);
  }

  // Check if user is authenticated
  isAuthenticated(): boolean {
    return !!this.getAccessToken();
  }

  // Login user
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    try {
      const response = await api.post('/auth/login', credentials);
      console.log('Login API response:', response.data);
      
      const { data } = response.data;
      console.log('Login data:', data);
      console.log('Access token:', data.access_token);
      console.log('Refresh token:', data.refresh_token);
      
      // Store tokens
      this.setTokens(data.access_token, data.refresh_token);
      
      // Set default authorization header
      api.defaults.headers.common['Authorization'] = `Bearer ${data.access_token}`;
      
      return data;
    } catch (error) {
      console.error('Login error:', error);
      throw new Error('Login failed');
    }
  }

  // Register user
  async register(userData: RegisterRequest): Promise<RegisterResponse> {
    try {
      const response = await api.post('/auth/register', userData);
      
      const { data } = response.data;
      return data;
    } catch (error: any) {
      // Extract actual error message from server response
      if (error.response?.data?.error) {
        throw new Error(error.response.data.error);
      } else if (error.response?.data?.message) {
        throw new Error(error.response.data.message);
      } else if (error.message) {
        throw new Error(error.message);
      } else {
        throw new Error('Registration failed. Please check your information and try again.');
      }
    }
  }

  // Refresh access token
  async refreshToken(): Promise<string | null> {
    const refreshToken = this.getRefreshToken();
    if (!refreshToken) {
      return null;
    }

    try {
      const response = await api.post('/auth/refresh', {
        refresh_token: refreshToken
      });
      
      const { data } = response.data;
      const newAccessToken = data.access_token;
      
      // Update stored access token
      localStorage.setItem(this.accessTokenKey, newAccessToken);
      
      // Update default authorization header
      api.defaults.headers.common['Authorization'] = `Bearer ${newAccessToken}`;
      
      return newAccessToken;
    } catch (error) {
      // If refresh fails, clear tokens and redirect to login
      this.clearTokens();
      delete api.defaults.headers.common['Authorization'];
      return null;
    }
  }

  // Logout user
  async logout(): Promise<void> {
    const refreshToken = this.getRefreshToken();
    
    if (refreshToken) {
      try {
        await api.post('/auth/logout', {
          refresh_token: refreshToken
        });
      } catch (error) {
        // Continue with logout even if API call fails
        console.warn('Logout API call failed:', error);
      }
    }

    // Clear tokens and headers
    this.clearTokens();
    delete api.defaults.headers.common['Authorization'];
  }

  // Setup axios interceptor for automatic token refresh
  setupTokenRefresh(): void {
    api.interceptors.response.use(
      (response) => response,
      async (error) => {
        const originalRequest = error.config;

        if (error.response?.status === 401 && !originalRequest._retry) {
          originalRequest._retry = true;

          const newToken = await this.refreshToken();
          if (newToken) {
            originalRequest.headers['Authorization'] = `Bearer ${newToken}`;
            return api(originalRequest);
          }
        }

        return Promise.reject(error);
      }
    );
  }
}

export const authService = new AuthService();

// Setup token refresh interceptor
authService.setupTokenRefresh();
