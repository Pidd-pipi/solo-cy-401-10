import { defineStore } from 'pinia';
import type { Bid } from '../types';
import { bidApi } from '../api/bid';

export const useBidStore = defineStore('bid', {
  state: () => ({
    bids: [] as Bid[]
  }),
  actions: {
    async fetchByRequirement(requirementId: number) {
      this.bids = await bidApi.listByRequirement(requirementId);
      return this.bids;
    }
  }
});
