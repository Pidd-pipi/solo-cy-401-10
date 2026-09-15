<template>
  <div class="page">
    <el-page-header content="合同详情" @back="$router.push('/dashboard')" />
    <el-card v-if="contract" class="detail-card">
      <div class="c-head">
        <h2>{{ contract.contractNo }}</h2>
        <StatusBadge :status="contract.status" kind="contract" />
      </div>
      <p class="muted">需求：{{ contract.requirement?.title }}</p>
      <p><b>{{ formatCurrency(contract.totalAmount) }}</b>
        <span class="muted"> · {{ contract.paymentType === 'installments' ? '分阶段付款' : '一次性付款' }}</span>
      </p>
      <div class="c-parties">
        <span>甲方：{{ contract.partyA?.name }}</span>
        <span>乙方：{{ contract.partyB?.name }}</span>
      </div>
      <div class="actions">
        <el-button v-if="contract.status === 'pending_signature'" type="primary" @click="sign">签署确认</el-button>
        <el-button v-if="contract.status === 'in_progress' && isPartyA" type="success" @click="complete">确认完成</el-button>
      </div>
    </el-card>

    <el-card v-if="contract" class="stages-card">
      <template #header><b>阶段进度</b></template>
      <ProgressSteps :stages="contract.stages" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';
import type { Contract } from '../types';
import { useContractStore } from '../stores/contract';
import { useUserStore } from '../stores/user';
import { contractApi } from '../api/contract';
import StatusBadge from '../components/common/StatusBadge.vue';
import ProgressSteps from '../components/common/ProgressSteps.vue';
import { formatCurrency } from '../utils/formatCurrency';

const route = useRoute();
const store = useContractStore();
const userStore = useUserStore();
const contract = ref<Contract | null>(null);

const isPartyA = computed(() => contract.value?.partyAId === userStore.user?.id);

async function load() {
  const id = Number(route.params.id);
  contract.value = await store.fetchDetail(id);
}

async function sign() {
  if (!contract.value) return;
  await contractApi.sign(contract.value.id);
  ElMessage.success('签署成功');
  void load();
}

async function complete() {
  if (!contract.value) return;
  await contractApi.complete(contract.value.id);
  ElMessage.success('已确认完成');
  void load();
}

onMounted(() => void load());
</script>

<style scoped>
.detail-card { margin-bottom: 16px; }
.c-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.c-parties { display: flex; gap: 24px; margin: 12px 0; }
.actions { margin-top: 16px; }
.stages-card { margin-bottom: 24px; }
</style>
