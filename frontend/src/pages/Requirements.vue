<template>
  <div class="page">
    <div class="page-title">
      <h2>需求大厅</h2>
      <el-button v-if="userStore.isRequester" type="primary" @click="showCreate = true">
        <el-icon><Plus /></el-icon>&nbsp;发布需求
      </el-button>
    </div>
    <FilterBar @search="onSearch" @reset="onReset" />
    <div v-loading="store.loading" class="card-grid">
      <RequirementCard v-for="req in store.list" :key="req.id" :requirement="req" />
    </div>
    <div class="pager">
      <el-pagination
        layout="prev, pager, next, total"
        :total="store.total"
        :page-size="pageSize"
        :current-page="page"
        @current-change="onPageChange"
      />
    </div>

    <el-dialog v-model="showCreate" title="发布需求" width="560px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="标题"><el-input v-model="form.title" placeholder="需求标题" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="4" placeholder="需求详细描述（至少 10 字）" /></el-form-item>
        <el-form-item label="预算范围">
          <el-input-number v-model="form.minBudget" :min="0" :step="1000" /> ~
          <el-input-number v-model="form.maxBudget" :min="0" :step="1000" />
        </el-form-item>
        <el-form-item label="截止日期"><el-date-picker v-model="form.deadline" type="date" value-format="YYYY-MM-DD" /></el-form-item>
        <el-form-item label="技能标签">
          <el-select v-model="form.skills" multiple filterable allow-create default-first-option placeholder="输入技能后回车" style="width: 100%">
            <el-option v-for="s in commonSkills" :key="s" :label="s" :value="s" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreate = false">取消</el-button>
        <el-button type="primary" @click="createRequirement">发布</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { useRequirementStore } from '../stores/requirement';
import { useUserStore } from '../stores/user';
import { requirementApi } from '../api/requirement';
import RequirementCard from '../components/common/RequirementCard.vue';
import FilterBar from '../components/common/FilterBar.vue';

const store = useRequirementStore();
const userStore = useUserStore();
const page = ref(1);
const pageSize = 12;
const filters = ref<Record<string, unknown>>({});
const showCreate = ref(false);
const commonSkills = ['Go', 'Vue', 'React', 'UI设计', 'ECharts', 'MySQL', 'WebSocket', 'Figma'];

const form = ref({
  title: '',
  description: '',
  minBudget: 10000,
  maxBudget: 30000,
  deadline: '',
  skills: [] as string[]
});

async function load() {
  await store.fetchList({ ...filters.value, page: page.value, page_size: pageSize });
}

function onSearch(payload: { skill?: string; minBudget?: number; maxBudget?: number; status?: string }) {
  filters.value = payload;
  page.value = 1;
  void load();
}

function onReset() {
  filters.value = {};
  page.value = 1;
  void load();
}

function onPageChange(p: number) {
  page.value = p;
  void load();
}

async function createRequirement() {
  if (!form.value.title || form.value.description.length < 10) {
    ElMessage.warning('请填写标题与详细描述');
    return;
  }
  await requirementApi.create(form.value);
  ElMessage.success('发布成功');
  showCreate.value = false;
  form.value = { title: '', description: '', minBudget: 10000, maxBudget: 30000, deadline: '', skills: [] };
  void load();
}

onMounted(() => void load());
</script>

<style scoped>
.pager { display: flex; justify-content: center; margin-top: 20px; }
</style>
