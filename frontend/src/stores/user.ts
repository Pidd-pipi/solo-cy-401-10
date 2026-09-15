import { defineStore } from 'pinia';
import type { User } from '../types';
import { authApi } from '../api/auth';

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('cyf_token') || '',
    user: (localStorage.getItem('cyf_user') ? JSON.parse(localStorage.getItem('cyf_user')!) : null) as User | null
  }),
  getters: {
    isAuthenticated: (state) => !!state.token,
    isRequester: (state) => !!state.user && (state.user.role === 'requester' || state.user.role === 'both'),
    isFreelancer: (state) => !!state.user && (state.user.role === 'freelancer' || state.user.role === 'both')
  },
  actions: {
    async login(username: string, password: string) {
      const data = await authApi.login(username, password);
      this.token = data.token;
      this.user = data.user;
      localStorage.setItem('cyf_token', data.token);
      localStorage.setItem('cyf_user', JSON.stringify(data.user));
    },
    async register(payload: { username: string; password: string; email?: string; name?: string; role: string }) {
      const data = await authApi.register(payload);
      this.token = data.token;
      this.user = data.user;
      localStorage.setItem('cyf_token', data.token);
      localStorage.setItem('cyf_user', JSON.stringify(data.user));
    },
    async refresh() {
      if (!this.token) return;
      try {
        this.user = await authApi.me();
        localStorage.setItem('cyf_user', JSON.stringify(this.user));
      } catch {
        // token invalid
      }
    },
    logout() {
      this.token = '';
      this.user = null;
      localStorage.removeItem('cyf_token');
      localStorage.removeItem('cyf_user');
    }
  }
});
