import { getData, patchData } from './request';
import type { User } from '../types';

export const userApi = {
  get: (id: number) => getData<User>(`/users/${id}`),
  update: (id: number, payload: Partial<User>) => patchData<User>(`/users/${id}`, payload)
};
