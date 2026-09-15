import { defineStore } from 'pinia';
import type { Contract } from '../types';
import { contractApi } from '../api/contract';

export const useContractStore = defineStore('contract', {
  state: () => ({
    contracts: [] as Contract[],
    current: null as Contract | null
  }),
  actions: {
    async fetchList() {
      this.contracts = await contractApi.list();
      return this.contracts;
    },
    async fetchDetail(id: number) {
      this.current = await contractApi.detail(id);
      return this.current;
    }
  }
});
