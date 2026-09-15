<template>
  <el-card class="req-card" shadow="hover" @click="$router.push(`/requirements/${requirement.id}`)">
    <div class="req-header">
      <span class="req-title">{{ requirement.title }}</span>
      <StatusBadge :status="requirement.status" kind="requirement" />
    </div>
    <p class="req-desc">{{ requirement.description }}</p>
    <div class="req-meta">
      <span class="muted">预算：<b>{{ formatCurrency(requirement.minBudget) }} ~ {{ formatCurrency(requirement.maxBudget) }}</b></span>
      <span class="muted">截止：{{ formatDate(requirement.deadline) }}</span>
    </div>
    <SkillTag :skills="requirement.skills" />
    <div class="req-footer">
      <span class="muted">{{ requirement.publisher?.name || '未知发布者' }}</span>
      <span class="muted">{{ (requirement.bids || []).length }} 份报价</span>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import type { Requirement } from '../../types';
import StatusBadge from './StatusBadge.vue';
import SkillTag from './SkillTag.vue';
import { formatCurrency, formatDate } from '../../utils/formatCurrency';

defineProps<{ requirement: Requirement }>();
</script>

<style scoped>
.req-card { margin-bottom: 16px; cursor: pointer; }
.req-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.req-title { font-size: 16px; font-weight: 600; }
.req-desc { color: #606266; font-size: 13px; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; margin-bottom: 10px; }
.req-meta { display: flex; gap: 16px; margin-bottom: 8px; }
.req-footer { display: flex; justify-content: space-between; margin-top: 10px; }
</style>
