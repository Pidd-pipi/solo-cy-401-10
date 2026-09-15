<template>
  <el-steps :active="activeStep" align-center finish-status="success">
    <el-step v-for="stage in stages" :key="stage.name" :title="stage.name" :description="`${formatCurrency(stage.amount)} · ${stage.dueAt}`" />
  </el-steps>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { ContractStage } from '../../types';
import { formatCurrency } from '../../utils/formatCurrency';

const props = defineProps<{ stages: ContractStage[] }>();
const activeStep = computed(() => {
  let done = 0;
  for (const s of props.stages) {
    if (s.status === 'done') done++;
  }
  return done;
});
</script>
