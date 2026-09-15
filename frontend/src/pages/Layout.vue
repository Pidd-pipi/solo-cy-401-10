<template>
  <el-container class="layout">
    <el-header class="header">
      <div class="brand" @click="$router.push('/requirements')">
        <span class="brand-mark">CY</span>
        <span class="brand-name">自由职业者撮合平台</span>
      </div>
      <el-menu mode="horizontal" :default-active="active" router class="menu" :ellipsis="false">
        <el-menu-item index="/requirements">需求大厅</el-menu-item>
        <el-menu-item index="/dashboard">我的工作台</el-menu-item>
      </el-menu>
      <div class="user-box">
        <template v-if="store.isAuthenticated">
          <el-dropdown @command="onCommand">
            <span class="user-name">
              <UserAvatar :name="store.user?.name" :avatar="store.user?.avatar" :size="28" />
              <span class="muted">{{ store.user?.name }}</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人资料</el-dropdown-item>
                <el-dropdown-item command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
        <el-button v-else size="small" @click="$router.push('/login')">登录</el-button>
      </div>
    </el-header>
    <el-main class="main">
      <router-view />
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useUserStore } from '../stores/user';
import UserAvatar from '../components/common/UserAvatar.vue';

const route = useRoute();
const router = useRouter();
const store = useUserStore();

const active = computed(() => {
  const path = route.path;
  if (path.startsWith('/requirements')) return '/requirements';
  if (path.startsWith('/dashboard')) return '/dashboard';
  if (path.startsWith('/contracts')) return '/dashboard';
  if (path.startsWith('/profile')) return '/dashboard';
  return '';
});

function onCommand(cmd: string) {
  if (cmd === 'logout') {
    store.logout();
    router.push('/login');
  } else if (cmd === 'profile') {
    router.push(`/profile/${store.user?.id}`);
  }
}
</script>

<style scoped>
.layout { min-height: 100vh; }
.header { display: flex; align-items: center; background: #fff; border-bottom: 1px solid #e4e7ed; padding: 0 24px; gap: 24px; }
.brand { display: flex; align-items: center; gap: 10px; cursor: pointer; }
.brand-mark { display: inline-grid; place-items: center; width: 34px; height: 34px; background: #243b53; color: #fff; font-weight: 800; border-radius: 8px; }
.brand-name { font-weight: 600; color: #243b53; }
.menu { flex: 1; border-bottom: none; }
.user-box { display: flex; align-items: center; }
.user-name { display: flex; align-items: center; gap: 8px; cursor: pointer; }
.main { background: #f5f7fa; padding: 0; }
</style>
