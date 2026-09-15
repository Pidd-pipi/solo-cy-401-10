<template>
  <el-form inline class="filter-bar">
    <el-form-item label="技能">
      <el-input v-model="skill" placeholder="技能关键词" clearable style="width: 160px" @keyup.enter="$emit('search', { skill })" />
    </el-form-item>
    <el-form-item label="最低预算">
      <el-input-number v-model="minBudget" :min="0" :step="5000" placeholder="最低" />
    </el-form-item>
    <el-form-item label="最高预算">
      <el-input-number v-model="maxBudget" :min="0" :step="5000" placeholder="最高" />
    </el-form-item>
    <el-form-item label="状态">
      <el-select v-model="status" clearable placeholder="全部" style="width: 130px">
        <el-option label="待报价" value="open" />
        <el-option label="报价中" value="bidding" />
        <el-option label="进行中" value="in_progress" />
        <el-option label="已完成" value="completed" />
      </el-select>
    </el-form-item>
    <el-form-item>
      <el-button type="primary" @click="$emit('search', { skill, minBudget, maxBudget, status })">筛选</el-button>
      <el-button @click="$emit('reset')">重置</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { ref } from 'vue';

defineEmits<{
  (e: 'search', payload: { skill?: string; minBudget?: number; maxBudget?: number; status?: string }): void;
  (e: 'reset'): void;
}>();

const skill = ref('');
const minBudget = ref<number | undefined>();
const maxBudget = ref<number | undefined>();
const status = ref<string | undefined>();
</script>
