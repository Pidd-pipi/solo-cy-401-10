<template>
  <div class="bid-card">
    <div class="bid-head">
      <div>
        <b>{{ formatCurrency(bid.amount) }}</b>
        <span class="muted"> · {{ bid.durationDays }} 天</span>
      </div>
      <StatusBadge :status="bid.status" kind="bid" />
    </div>
    <p class="bid-proposal">{{ bid.proposal }}</p>
    <div class="bid-footer">
      <span class="muted">报价人：{{ bid.bidder?.name || '未知' }}</span>
      <el-button v-if="canAccept" type="primary" size="small" @click="$emit('accept', bid.id)">采纳</el-button>
      <el-button v-if="canWithdraw" size="small" @click="$emit('withdraw', bid.id)">撤回</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Bid } from '../../types';
import StatusBadge from './StatusBadge.vue';
import { formatCurrency } from '../../utils/formatCurrency';

defineProps<{ bid: Bid; canAccept?: boolean; canWithdraw?: boolean }>();
defineEmits<{ (e: 'accept', id: number): void; (e: 'withdraw', id: number): void }>();
</script>

<style scoped>
.bid-head { display: flex; justify-content: space-between; align-items: center; }
.bid-proposal { color: #606266; font-size: 13px; margin: 8px 0; }
.bid-footer { display: flex; justify-content: space-between; align-items: center; }
</style>
