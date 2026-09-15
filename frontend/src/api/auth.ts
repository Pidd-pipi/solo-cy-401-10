import { getData, postData } from './request';
import type { User } from '../types';

export const authApi = {
  login: (username: string, password: string) =>
    postData<{ token: string; user: User }>('/auth/login', { username, password }),
  register: (payload: { username: string; password: string; email?: string; name?: string; role: string }) =>
    postData<{ token: string; user: User }>('/auth/register', payload),
  me: () => getData<User>('/auth/me')
};
