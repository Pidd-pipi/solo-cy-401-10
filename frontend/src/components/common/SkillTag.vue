<template>
  <el-tag v-for="skill in skills" :key="skill" size="small" class="skill-tag" effect="plain">{{ skill }}</el-tag>
</template>

<script setup lang="ts">
defineProps<{ skills?: string[] }>();
</script>

<style scoped>
.skill-tag { margin-right: 6px; margin-bottom: 4px; }
</style>
