import { defineStore } from 'pinia';
import type { PageResult, Requirement } from '../types';
import { requirementApi, type RequirementQuery } from '../api/requirement';

export const useRequirementStore = defineStore('requirement', {
  state: () => ({
    list: [] as Requirement[],
    total: 0,
    loading: false,
    current: null as Requirement | null
  }),
  actions: {
    async fetchList(query: RequirementQuery = {}) {
      this.loading = true;
      try {
        const res: PageResult<Requirement> = await requirementApi.list(query);
        this.list = res.items;
        this.total = res.total;
      } finally {
        this.loading = false;
      }
    },
    async fetchDetail(id: number) {
      this.current = await requirementApi.detail(id);
      return this.current;
    }
  }
});
