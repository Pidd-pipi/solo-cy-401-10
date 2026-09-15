import { getData, postData } from './request';
import type { Bid } from '../types';

export const bidApi = {
  listByRequirement: (requirementId: number) => getData<Bid[]>('/bids', { requirementId }),
  create: (payload: Partial<Bid>) => postData<Bid>('/bids', payload),
  withdraw: (id: number) => postData<Bid>(`/bids/${id}/withdraw`)
};
