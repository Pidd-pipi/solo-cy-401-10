<template>
  <div class="auth-page">
    <el-card class="auth-card">
      <h2 class="auth-title">自由职业者撮合平台</h2>
      <el-form :model="form" label-width="0">
        <el-form-item>
          <el-input v-model="form.username" placeholder="用户名" size="large" />
        </el-form-item>
        <el-form-item>
          <el-input v-model="form.password" type="password" placeholder="密码" size="large" show-password @keyup.enter="login" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" size="large" style="width: 100%" :loading="loading" @click="login">登录</el-button>
        </el-form-item>
        <el-form-item>
          <el-button size="large" style="width: 100%" @click="$router.push('/register')">注册新账号</el-button>
        </el-form-item>
      </el-form>
      <div class="auth-tips muted">
        <p>测试账号：requester01 / demo123456（需求方）</p>
        <p>freelancer01 / demo123456（自由职业者）</p>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { useUserStore } from '../stores/user';

const router = useRouter();
const store = useUserStore();
const loading = ref(false);
const form = ref({ username: '', password: '' });

async function login() {
  if (!form.value.username || !form.value.password) {
    ElMessage.warning('请输入用户名和密码');
    return;
  }
  loading.value = true;
  try {
    await store.login(form.value.username, form.value.password);
    ElMessage.success('登录成功');
    router.push('/dashboard');
  } catch {
    // message shown by interceptor
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.auth-page { min-height: 100vh; display: flex; align-items: center; justify-content: center; background: linear-gradient(135deg, #243b53 0%, #56789c 100%); }
.auth-card { width: 400px; padding: 12px 8px; }
.auth-title { text-align: center; margin-bottom: 24px; color: #243b53; }
.auth-tips { margin-top: 8px; font-size: 12px; }
.auth-tips p { margin: 4px 0; }
</style>
