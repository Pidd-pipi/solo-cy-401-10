import axios from 'axios';
import { ElMessage } from 'element-plus';
import router from '../router';

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export const request = axios.create({
  baseURL: '/api/v1',
  timeout: 15000
});

request.interceptors.request.use((config) => {
  const token = localStorage.getItem('cyf_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

request.interceptors.response.use(
  (response) => {
    const body = response.data as ApiResponse<unknown>;
    if (body && typeof body === 'object' && 'code' in body) {
      if (body.code === 0) {
        return response;
      }
      ElMessage.error(body.message || '请求失败');
      return Promise.reject(new Error(body.message || '请求失败'));
    }
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('cyf_token');
      localStorage.removeItem('cyf_user');
      if (router.currentRoute.value.path !== '/login') {
        router.push('/login');
      }
    }
    const text = error.response?.data?.message || error.message || '请求失败';
    ElMessage.error(text);
    return Promise.reject(error);
  }
);

export async function getData<T>(url: string, params?: Record<string, unknown>): Promise<T> {
  const res = await request.get<ApiResponse<T>>(url, { params });
  return res.data.data;
}

export async function postData<T>(url: string, payload?: unknown): Promise<T> {
  const res = await request.post<ApiResponse<T>>(url, payload);
  return res.data.data;
}

export async function putData<T>(url: string, payload?: unknown): Promise<T> {
  const res = await request.put<ApiResponse<T>>(url, payload);
  return res.data.data;
}

export async function patchData<T>(url: string, payload?: unknown): Promise<T> {
  const res = await request.patch<ApiResponse<T>>(url, payload);
  return res.data.data;
}
