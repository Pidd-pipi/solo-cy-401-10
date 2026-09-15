import { computed } from 'vue';
import { useUserStore } from '../stores/user';

export function useAuth() {
  const store = useUserStore();
  const isAuthenticated = computed(() => store.isAuthenticated);
  const user = computed(() => store.user);
  return { store, isAuthenticated, user };
}
