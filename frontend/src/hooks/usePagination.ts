import { ref, computed } from 'vue';

export function usePagination(total: number, pageSize = 10) {
  const page = ref(1);
  const pageSizeRef = ref(pageSize);
  const totalRef = ref(total);
  const pagedItems = computed(() => [] as unknown[]);
  return { page, pageSize: pageSizeRef, total: totalRef, pagedItems };
}
