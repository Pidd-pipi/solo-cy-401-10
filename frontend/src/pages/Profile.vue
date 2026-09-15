<template>
  <div class="page">
    <el-card v-if="profile" class="profile-card">
      <div class="profile-head">
        <UserAvatar :name="profile.name" :avatar="profile.avatar" :size="64" />
        <div>
          <h2>{{ profile.name }}</h2>
          <span class="muted">@{{ profile.username }} · {{ RoleLabel[profile.role] || profile.role }}</span>
        </div>
      </div>
      <el-descriptions :column="2" border class="desc">
        <el-descriptions-item label="邮箱">{{ profile.email || '-' }}</el-descriptions-item>
        <el-descriptions-item label="联系方式">{{ profile.contact || '-' }}</el-descriptions-item>
        <el-descriptions-item label="评分">{{ profile.rating?.toFixed(1) || '0.0' }}</el-descriptions-item>
        <el-descriptions-item label="简介">{{ profile.bio || '-' }}</el-descriptions-item>
      </el-descriptions>
      <div class="skills">
        <b>技能标签：</b>
        <SkillTag :skills="profile.skills" />
      </div>

      <template v-if="isSelf">
        <el-divider />
        <h3>编辑资料</h3>
        <el-form :model="form" label-width="90px" class="edit-form">
          <el-form-item label="昵称"><el-input v-model="form.name" /></el-form-item>
          <el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item>
          <el-form-item label="联系方式"><el-input v-model="form.contact" /></el-form-item>
          <el-form-item label="简介"><el-input v-model="form.bio" type="textarea" :rows="2" /></el-form-item>
          <el-form-item label="技能标签">
            <el-select v-model="form.skills" multiple filterable allow-create default-first-option style="width: 100%">
              <el-option v-for="s in commonSkills" :key="s" :label="s" :value="s" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="saving" @click="save">保存</el-button>
          </el-form-item>
        </el-form>
      </template>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';
import type { User } from '../types';
import { RoleLabel } from '../types/enums';
import { userApi } from '../api/user';
import { useUserStore } from '../stores/user';
import UserAvatar from '../components/common/UserAvatar.vue';
import SkillTag from '../components/common/SkillTag.vue';

const route = useRoute();
const userStore = useUserStore();
const profile = ref<User | null>(null);
const saving = ref(false);
const commonSkills = ['Go', 'Vue', 'React', 'UI设计', 'ECharts', 'MySQL', 'WebSocket', 'Figma', '品牌'];

const isSelf = computed(() => profile.value?.id === userStore.user?.id);
const form = ref({ name: '', email: '', contact: '', bio: '', skills: [] as string[] });

async function load() {
  const id = Number(route.params.id);
  profile.value = await userApi.get(id);
  form.value = {
    name: profile.value.name || '',
    email: profile.value.email || '',
    contact: profile.value.contact || '',
    bio: profile.value.bio || '',
    skills: profile.value.skills || []
  };
}

async function save() {
  if (!profile.value) return;
  saving.value = true;
  try {
    profile.value = await userApi.update(profile.value.id, form.value);
    ElMessage.success('保存成功');
  } finally {
    saving.value = false;
  }
}

onMounted(() => void load());
</script>

<style scoped>
.profile-card { max-width: 900px; margin: 0 auto; }
.profile-head { display: flex; align-items: center; gap: 16px; margin-bottom: 20px; }
.desc { margin-bottom: 16px; }
.skills { margin-top: 12px; }
.edit-form { margin-top: 12px; }
</style>
