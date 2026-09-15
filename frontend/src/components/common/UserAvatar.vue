<template>
  <el-avatar :size="size" :src="avatar">{{ initial }}</el-avatar>
</template>

<script setup lang="ts">
import { computed } from 'vue';

const props = withDefaults(defineProps<{ name?: string; avatar?: string; size?: number }>(), { size: 32 });
const initial = computed(() => (props.name || '?').slice(0, 1));
</script>
