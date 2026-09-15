import { useUserStore } from '../stores/user';

export function authGuard(): boolean {
  const store = useUserStore();
  return store.isAuthenticated;
}
