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
  phone?: string;
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
  session_id: string;
  expires_at?: string;
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
  session_id: string;
}

export interface RefreshTokenResponse {
  access_token: string;
  expires_at?: string;
}

export interface LogoutRequest {
  refresh_token: string;
}

class AuthService {
  private accessTokenKey = 'access_token';
  private refreshTokenKey = 'refresh_token';
  private sessionIdKey = 'session_id';

  // Store tokens in localStorage
  setTokens(accessToken: string, refreshToken: string, sessionId?: string): void {
    localStorage.setItem(this.accessTokenKey, accessToken);
    localStorage.setItem(this.refreshTokenKey, refreshToken);
    if (sessionId) localStorage.setItem(this.sessionIdKey, sessionId);
  }

  // Get access token
  getAccessToken(): string | null {
    return localStorage.getItem(this.accessTokenKey);
  }

  // Get refresh token
  getRefreshToken(): string | null {
    return localStorage.getItem(this.refreshTokenKey);
  }

  // Get session id
  getSessionId(): string | null {
    return localStorage.getItem(this.sessionIdKey);
  }

  // Clear tokens
  clearTokens(): void {
    localStorage.removeItem(this.accessTokenKey);
    localStorage.removeItem(this.refreshTokenKey);
    localStorage.removeItem(this.sessionIdKey);
  }

  // Check if user is authenticated
  isAuthenticated(): boolean {
    return !!this.getAccessToken();
  }

  // Login user
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    try {
      const response = await api.post('/auth/login', credentials);
      const { data } = response.data;
      
      // Store tokens
      this.setTokens(data.access_token, data.refresh_token, data.session_id);
      
      // Set default authorization header
      api.defaults.headers.common['Authorization'] = `Bearer ${data.access_token}`;

      // Normalize user shape and persist for app reloads
      const user = {
        id: data.user_id,
        email: data.email,
        first_name: data.first_name,
        last_name: data.last_name,
      };
      try {
        localStorage.setItem('user_info', JSON.stringify(user));
      } catch {}

      // Return in expected shape for UI
      return {
        user,
        access_token: data.access_token,
        refresh_token: data.refresh_token,
        session_id: data.session_id,
        expires_at: data.expires_at,
      } as unknown as LoginResponse;
    } catch (error: any) {
      if (error.response?.data?.error?.message) {
        throw new Error(error.response.data.error.message);
      } else if (typeof error.response?.data?.error === 'string') {
        throw new Error(error.response.data.error);
      } else if (error.response?.data?.message) {
        throw new Error(error.response.data.message);
      } else if (error.message) {
        throw new Error(error.message);
      } else {
        throw new Error('Login failed');
      }
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
      if (error.response?.data?.error?.message) {
        throw new Error(error.response.data.error.message);
      } else if (typeof error.response?.data?.error === 'string') {
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
    const sessionId = this.getSessionId();
    if (!sessionId) {
      return null;
    }

    try {
      const response = await api.post('/auth/refresh', { session_id: sessionId });
      
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
    const sessionId = this.getSessionId();
    if (sessionId) {
      try {
        await api.post('/auth/logout', { session_id: sessionId });
      } catch (error) {
        // Continue with logout even if API call fails
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
