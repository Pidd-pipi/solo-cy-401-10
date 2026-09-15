import { createRouter, createWebHistory } from 'vue-router';
import { useUserStore } from '../stores/user';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'Login', component: () => import('../pages/Login.vue') },
    { path: '/register', name: 'Register', component: () => import('../pages/Register.vue') },
    {
      path: '/',
      component: () => import('../pages/Layout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', redirect: '/requirements' },
        { path: 'requirements', name: 'Requirements', component: () => import('../pages/Requirements.vue') },
        { path: 'requirements/:id', name: 'RequirementDetail', component: () => import('../pages/RequirementDetail.vue') },
        { path: 'dashboard', name: 'Dashboard', component: () => import('../pages/Dashboard.vue') },
        { path: 'contracts/:id', name: 'ContractDetail', component: () => import('../pages/ContractDetail.vue') },
        { path: 'profile/:id', name: 'Profile', component: () => import('../pages/Profile.vue') }
      ]
    }
  ]
});

router.beforeEach((to) => {
  const store = useUserStore();
  if (to.meta.requiresAuth && !store.isAuthenticated) {
    return { path: '/login', query: { redirect: to.fullPath } };
  }
  if ((to.path === '/login' || to.path === '/register') && store.isAuthenticated) {
    return { path: '/dashboard' };
  }
  return true;
});

export default router;
