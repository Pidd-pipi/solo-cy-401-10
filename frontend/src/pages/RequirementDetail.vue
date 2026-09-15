<template>
  <div class="page">
    <el-page-header content="需求详情" @back="$router.push('/requirements')" />
    <el-card v-if="req" class="detail-card">
      <div class="req-header">
        <h2>{{ req.title }}</h2>
        <StatusBadge :status="req.status" kind="requirement" />
      </div>
      <p class="req-desc">{{ req.description }}</p>
      <div class="req-meta">
        <span class="muted">预算：<b>{{ formatCurrency(req.minBudget) }} ~ {{ formatCurrency(req.maxBudget) }}</b></span>
        <span class="muted">截止：{{ formatDate(req.deadline) }}</span>
        <span class="muted">发布者：{{ req.publisher?.name }}</span>
      </div>
      <SkillTag :skills="req.skills" />
      <div class="req-actions" v-if="isOwner">
        <el-button v-if="req.status === 'open' || req.status === 'bidding'" @click="changeStatus('in_progress')">开始执行</el-button>
        <el-button v-if="req.status === 'in_progress'" @click="changeStatus('pending_review')">提交验收</el-button>
        <el-button v-if="req.status === 'pending_review'" type="primary" @click="changeStatus('completed')">确认完成</el-button>
      </div>
    </el-card>

    <el-card class="bids-card">
      <template #header><b>报价列表（{{ bids.length }}）</b></template>
      <BidCard
        v-for="bid in bids"
        :key="bid.id"
        :bid="bid"
        :can-accept="isOwner && req?.status === 'open' && bid.status === 'pending'"
        :can-withdraw="bid.bidderId === userStore.user?.id && bid.status === 'pending'"
        @accept="acceptBid"
        @withdraw="withdrawBid"
      />
      <el-empty v-if="bids.length === 0" description="暂无报价" />
    </el-card>

    <el-card v-if="userStore.isFreelancer" class="bid-form-card">
      <template #header><b>提交报价</b></template>
      <el-form :model="bidForm" label-width="90px">
        <el-form-item label="报价金额"><el-input-number v-model="bidForm.amount" :min="0" :step="1000" /></el-form-item>
        <el-form-item label="工期(天)"><el-input-number v-model="bidForm.durationDays" :min="1" /></el-form-item>
        <el-form-item label="提案"><el-input v-model="bidForm.proposal" type="textarea" :rows="3" placeholder="说明你的方案与经验（至少 10 字）" /></el-form-item>
      </el-form>
      <el-button type="primary" :loading="submitting" @click="submitBid">提交报价</el-button>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';
import type { Bid, Requirement } from '../types';
import { useRequirementStore } from '../stores/requirement';
import { useUserStore } from '../stores/user';
import { bidApi } from '../api/bid';
import { requirementApi } from '../api/requirement';
import StatusBadge from '../components/common/StatusBadge.vue';
import SkillTag from '../components/common/SkillTag.vue';
import BidCard from '../components/common/BidCard.vue';
import { formatCurrency, formatDate } from '../utils/formatCurrency';

const route = useRoute();
const reqStore = useRequirementStore();
const userStore = useUserStore();
const req = ref<Requirement | null>(null);
const bids = ref<Bid[]>([]);
const submitting = ref(false);
const bidForm = ref({ amount: 10000, durationDays: 30, proposal: '' });

const isOwner = computed(() => req.value?.publisherId === userStore.user?.id);

async function load() {
  const id = Number(route.params.id);
  req.value = await reqStore.fetchDetail(id);
  bids.value = await bidApi.listByRequirement(id);
}

async function acceptBid(bidId: number) {
  if (!req.value) return;
  await requirementApi.acceptBid(req.value.id, bidId);
  ElMessage.success('已采纳报价并生成合同');
  void load();
}

async function withdrawBid(bidId: number) {
  await bidApi.withdraw(bidId);
  ElMessage.success('已撤回');
  void load();
}

async function changeStatus(status: string) {
  if (!req.value) return;
  await requirementApi.updateStatus(req.value.id, status);
  ElMessage.success('状态已更新');
  void load();
}

async function submitBid() {
  if (!req.value) return;
  if (bidForm.value.proposal.length < 10) {
    ElMessage.warning('请填写完整提案');
    return;
  }
  submitting.value = true;
  try {
    await bidApi.create({ requirementId: req.value.id, ...bidForm.value, attachments: [] });
    ElMessage.success('报价已提交');
    bidForm.value = { amount: 10000, durationDays: 30, proposal: '' };
    void load();
  } finally {
    submitting.value = false;
  }
}

onMounted(() => void load());
</script>

<style scoped>
.detail-card { margin-bottom: 16px; }
.req-header { display: flex; justify-content: space-between; align-items: center; }
.req-desc { color: #606266; margin: 12px 0; line-height: 1.7; }
.req-meta { display: flex; gap: 20px; margin-bottom: 12px; }
.req-actions { margin-top: 16px; }
.bids-card { margin-bottom: 16px; }
.bid-form-card { margin-bottom: 24px; }
</style>
