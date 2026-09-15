<template>
  <div class="page">
    <div class="page-title">
      <h2>我的工作台</h2>
      <span class="muted">你好，{{ userStore.user?.name }}</span>
    </div>

    <el-row :gutter="16" class="stats">
      <el-col :span="8">
        <el-card><div class="stat"><b>{{ data?.counts.requirements || 0 }}</b><span class="muted">已发布需求</span></div></el-card>
      </el-col>
      <el-col :span="8">
        <el-card><div class="stat"><b>{{ data?.counts.bids || 0 }}</b><span class="muted">已提交报价</span></div></el-card>
      </el-col>
      <el-col :span="8">
        <el-card><div class="stat"><b>{{ data?.counts.contracts || 0 }}</b><span class="muted">进行中合同</span></div></el-card>
      </el-col>
    </el-row>

    <el-tabs v-model="tab">
      <el-tab-pane label="我发布的需求" name="reqs">
        <div class="card-grid">
          <RequirementCard v-for="r in data?.myRequirements || []" :key="r.id" :requirement="r" />
        </div>
        <el-empty v-if="(data?.myRequirements || []).length === 0" description="暂无发布需求" />
      </el-tab-pane>
      <el-tab-pane label="我的报价" name="bids">
        <BidCard v-for="b in data?.myBids || []" :key="b.id" :bid="b" :can-withdraw="b.status === 'pending'" @withdraw="withdrawBid" />
        <el-empty v-if="(data?.myBids || []).length === 0" description="暂无报价" />
      </el-tab-pane>
      <el-tab-pane label="我的合同" name="contracts">
        <div class="card-grid">
          <ContractCard v-for="c in data?.myContracts || []" :key="c.id" :contract="c" />
        </div>
        <el-empty v-if="(data?.myContracts || []).length === 0" description="暂无合同" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { getData } from '../api/request';
import { bidApi } from '../api/bid';
import type { DashboardData } from '../types';
import { useUserStore } from '../stores/user';
import RequirementCard from '../components/common/RequirementCard.vue';
import BidCard from '../components/common/BidCard.vue';
import ContractCard from '../components/common/ContractCard.vue';

const userStore = useUserStore();
const data = ref<DashboardData | null>(null);
const tab = ref('reqs');

async function load() {
  data.value = await getData<DashboardData>('/dashboard');
}

async function withdrawBid(id: number) {
  await bidApi.withdraw(id);
  ElMessage.success('已撤回');
  void load();
}

onMounted(() => void load());
</script>

<style scoped>
.stats { margin-bottom: 20px; }
.stat { display: flex; flex-direction: column; align-items: center; padding: 8px; }
.stat b { font-size: 28px; color: #243b53; }
</style>
