import { getData, postData, putData } from './request';
import type { PageResult, Requirement } from '../types';

export interface RequirementQuery {
  status?: string;
  minBudget?: number;
  maxBudget?: number;
  skill?: string;
  page?: number;
  page_size?: number;
}

export const requirementApi = {
  list: (query: RequirementQuery) => getData<PageResult<Requirement>>('/requirements', query as Record<string, unknown>),
  detail: (id: number) => getData<Requirement>(`/requirements/${id}`),
  create: (payload: Partial<Requirement>) => postData<Requirement>('/requirements', payload),
  update: (id: number, payload: Partial<Requirement>) => putData<Requirement>(`/requirements/${id}`, payload),
  updateStatus: (id: number, status: string) => postData<Requirement>(`/requirements/${id}/status`, { status }),
  acceptBid: (id: number, bidId: number) => postData<unknown>(`/requirements/${id}/accept-bid`, { bidId })
};
