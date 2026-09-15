import { getData, postData } from './request';
import type { Contract } from '../types';

export const contractApi = {
  list: () => getData<Contract[]>('/contracts'),
  detail: (id: number) => getData<Contract>(`/contracts/${id}`),
  sign: (id: number) => postData<Contract>(`/contracts/${id}/sign`),
  complete: (id: number) => postData<Contract>(`/contracts/${id}/complete`)
};
