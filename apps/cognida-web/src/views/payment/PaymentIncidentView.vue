<template>
  <section class="payment-workbench">
    <header>
      <p class="eyebrow">PAYMENT INCIDENT</p>
      <h1>支付故障核验台</h1>
      <p>先检索可追溯证据，再由人工决定补单、回放或升级处理。</p>
    </header>

    <form class="query-card" @submit.prevent="search">
      <label for="incident-query">订单号或故障现象</label>
      <textarea id="incident-query" v-model="query" rows="3" placeholder="例如：ORD-1020 支付成功但订单仍为 PAYING，日志显示 db_timeout" />
      <div class="actions">
        <select v-model="kbId" aria-label="选择知识库">
          <option value="">全部可访问知识库</option>
          <option v-for="kb in knowledgeBases" :key="kb.id" :value="kb.id">{{ kb.name }}</option>
        </select>
        <button :disabled="loading || !query.trim()">{{ loading ? '检索中…' : '检索证据' }}</button>
      </div>
      <div class="quick-questions">
        <button v-for="item in paymentStarterQuestions" :key="item" type="button" @click="query = item">{{ item }}</button>
      </div>
    </form>

    <p v-if="error" class="error">{{ error }}</p>
    <template v-if="searched">
      <section class="checklist">
        <h2>建议核验顺序</h2>
        <ol><li v-for="item in checklist" :key="item">{{ item }}</li></ol>
      </section>
      <section class="evidence">
        <div class="evidence-title"><h2>检索证据</h2><span>{{ results.length }} 条</span></div>
        <p v-if="!results.length" class="empty">未找到可直接支持本次判断的证据，请转人工处理。</p>
        <article v-for="item in results" :key="item.chunk_id" class="evidence-card">
          <div><strong>{{ item.knowledge_title || '知识库片段' }}</strong><span>相对匹配 {{ relativeMatchPercent(item.score, maxScore) }}%</span></div>
          <p><template v-for="(segment, index) in splitByQueryTerms(item.content, query)" :key="index"><mark v-if="segment.matched">{{ segment.text }}</mark><template v-else>{{ segment.text }}</template></template></p>
        </article>
      </section>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { knowledgeApi } from '@/api/knowledge'
import { useKnowledgeStore } from '@/stores/knowledge'
import type { SearchResult } from '@/types'
import { buildInvestigationChecklist, paymentStarterQuestions, relativeMatchPercent, splitByQueryTerms } from '@/features/payment-incident/investigation'

const store = useKnowledgeStore()
const { knowledgeBases } = storeToRefs(store)
const query = ref('')
const kbId = ref('')
const results = ref<SearchResult[]>([])
const loading = ref(false)
const searched = ref(false)
const error = ref('')
const checklist = computed(() => buildInvestigationChecklist(query.value, results.value))
const scopedKbIds = computed(() => kbId.value ? [kbId.value] : knowledgeBases.value.map(kb => kb.id))
const maxScore = computed(() => Math.max(0, ...results.value.map(item => item.score)))

onMounted(() => { void store.loadKnowledgeBases() })

async function search() {
  if (scopedKbIds.value.length === 0) {
    error.value = '当前没有可检索的知识库，请先创建并导入支付处置文档。'
    searched.value = false
    return
  }
  loading.value = true
  error.value = ''
  try {
    const response = await knowledgeApi.search({
      query: query.value.trim(), kb_ids: scopedKbIds.value, top_k: 5, retrieval_mode: 'hybrid'
    })
    results.value = response.data?.items ?? []
    searched.value = true
  } catch (e) {
    error.value = e instanceof Error ? e.message : '检索失败，请稍后重试。'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.payment-workbench { max-width: 960px; margin: 0 auto; color: var(--text-primary); }
header { margin-bottom: 24px; } h1, h2, p { margin: 0; } h1 { font-size: 28px; margin: 4px 0 8px; } header p:last-child { color: var(--text-muted); }
.eyebrow { color: var(--primary); font: 12px var(--font-mono); letter-spacing: .12em; }
.query-card, .checklist, .evidence { padding: 20px; border: 1px solid var(--color-border-subtle); border-radius: var(--radius-md); background: var(--bg-tertiary); }
label { display: block; font-weight: 600; margin-bottom: 8px; } textarea, select { box-sizing: border-box; width: 100%; border: 1px solid var(--color-border-subtle); border-radius: 6px; padding: 10px; background: var(--bg-primary); color: var(--text-primary); font: inherit; }
.actions { display: flex; gap: 10px; margin-top: 12px; } .actions select { flex: 1; } button { border: 0; border-radius: 6px; padding: 9px 13px; cursor: pointer; background: var(--primary); color: var(--on-primary); font: inherit; } button:disabled { cursor: not-allowed; opacity: .6; }
.quick-questions { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 14px; } .quick-questions button { background: var(--bg-elevated); color: var(--text-secondary); font-size: 12px; }
.checklist, .evidence { margin-top: 18px; } ol { margin-bottom: 0; padding-left: 20px; } li + li { margin-top: 8px; } .evidence-title, .evidence-card > div { display: flex; justify-content: space-between; gap: 12px; } .evidence-title span, .evidence-card span { color: var(--text-muted); font: 12px var(--font-mono); }
.evidence-card { margin-top: 12px; padding: 14px; border-left: 3px solid var(--primary); background: var(--bg-primary); } .evidence-card p { margin-top: 8px; white-space: pre-wrap; line-height: 1.65; } mark { padding: 0 2px; border-radius: 2px; background: rgba(156, 180, 205, .24); color: inherit; } .empty, .error { margin-top: 16px; color: var(--text-muted); } .error { color: var(--danger); }
@media (max-width: 640px) { .actions { flex-direction: column; } }
</style>
