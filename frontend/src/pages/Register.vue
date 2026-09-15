<template>
  <div class="auth-page">
    <el-card class="auth-card">
      <h2 class="auth-title">注册账号</h2>
      <el-form :model="form" label-width="80px">
        <el-form-item label="用户名"><el-input v-model="form.username" placeholder="2-64 字符" /></el-form-item>
        <el-form-item label="密码"><el-input v-model="form.password" type="password" show-password placeholder="至少 6 位" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="form.email" placeholder="选填" /></el-form-item>
        <el-form-item label="昵称"><el-input v-model="form.name" placeholder="选填" /></el-form-item>
        <el-form-item label="角色">
          <el-radio-group v-model="form.role">
            <el-radio value="requester">需求方</el-radio>
            <el-radio value="freelancer">自由职业者</el-radio>
            <el-radio value="both">双角色</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" style="width: 100%" :loading="loading" @click="register">注册</el-button>
        </el-form-item>
        <el-form-item>
          <el-button style="width: 100%" @click="$router.push('/login')">返回登录</el-button>
        </el-form-item>
      </el-form>
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
const form = ref({ username: '', password: '', email: '', name: '', role: 'freelancer' });

async function register() {
  if (!form.value.username || form.value.password.length < 6) {
    ElMessage.warning('请填写用户名且密码不少于 6 位');
    return;
  }
  loading.value = true;
  try {
    await store.register(form.value);
    ElMessage.success('注册成功');
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
.auth-card { width: 440px; }
.auth-title { text-align: center; margin-bottom: 20px; color: #243b53; }
</style>
