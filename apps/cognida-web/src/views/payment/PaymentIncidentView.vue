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
      <section class="recommendation" aria-live="polite">
        <div class="recommendation-title"><h2>受证据约束的处置建议</h2><span>仅供人工核验</span></div>
        <dl>
          <dt>故障判断</dt><dd>{{ recommendation.judgment }}</dd>
          <dt>已知事实</dt><dd><ul><li v-for="fact in recommendation.knownFacts" :key="fact">{{ fact }}</li></ul></dd>
          <dt>待核验项</dt><dd><ul><li v-for="item in recommendation.pendingChecks" :key="item">{{ item }}</li></ul></dd>
          <dt>建议动作</dt><dd>{{ recommendation.recommendedAction }}</dd>
          <dt>风险提示</dt><dd class="risk-notice">{{ recommendation.riskNotice }}</dd>
          <dt>引用证据</dt><dd>{{ recommendation.evidenceReferences.length ? recommendation.evidenceReferences.join('、') : '无可追溯引用' }}</dd>
        </dl>
      </section>
      <section v-if="toolFacts || toolFactError" class="tool-facts">
        <div class="evidence-title"><h2>模拟只读工具事实</h2><span>仅本地演示数据</span></div>
        <p v-if="toolFactError" class="empty">{{ toolFactError }}</p>
        <template v-else-if="toolFacts">
          <dl><dt>订单</dt><dd>{{ toolFacts.order.id }} · {{ toolFacts.order.status }} · {{ toolFacts.order.amount_minor }} {{ toolFacts.order.currency }} · {{ toolFacts.order.payment_channel }}</dd><dt>支付流水</dt><dd>{{ toolFacts.payments.length ? toolFacts.payments.map(item => `${item.provider_trade_no}：${item.status}`).join('；') : '无记录' }}</dd><dt>回调日志</dt><dd><ul><li v-for="item in toolFacts.callbacks" :key="item.id">{{ item.event_type }} · {{ item.delivery_status }}：{{ item.payload_summary }}</li></ul></dd></dl>
        </template>
      </section>
      <section class="case-panel">
        <div class="case-title"><div><h2>人工案件</h2><p>创建后由人工负责状态流转；不会执行补单或回放。</p></div><button :disabled="creatingCase || !searched" @click="createCase">{{ creatingCase ? '创建中…' : '创建案件' }}</button></div>
        <p v-if="caseMessage" class="case-message">{{ caseMessage }}</p>
        <div class="case-title case-list-title"><h3>当前租户案件</h3><button class="secondary" :disabled="loadingCases" @click="loadCases">{{ loadingCases ? '刷新中…' : '刷新案件' }}</button></div>
        <p v-if="!incidents.length" class="empty">尚未创建案件。</p>
        <template v-for="item in incidents" :key="item.id">
          <article class="incident-card">
            <div><strong>{{ item.order_id || '未提供订单号' }}</strong><span>{{ incidentStatusLabel(item.status) }} · {{ item.priority }}</span></div>
            <p>{{ item.incident_type }} · {{ item.id }}</p>
            <div class="transition-actions"><button class="secondary" :disabled="detailLoadingId === item.id" @click="loadDetail(item.id)">{{ detailLoadingId === item.id ? '加载详情中…' : '查看详情' }}</button><button v-for="status in availableIncidentTransitions(item.status)" :key="status" class="secondary" :disabled="transitioningId === item.id" @click="transitionCase(item.id, status)">转为{{ incidentStatusLabel(status) }}</button></div>
          </article>
          <section v-if="selectedDetail?.incident.id === item.id" class="incident-detail">
          <div class="case-title"><h3>案件详情</h3><button class="secondary" @click="selectedDetail = null">收起</button></div>
          <dl><dt>订单号</dt><dd>{{ selectedDetail.incident.order_id || '未提供' }}</dd><dt>类型 / 优先级</dt><dd>{{ selectedDetail.incident.incident_type }} / {{ selectedDetail.incident.priority }}</dd><dt>状态 / 责任人</dt><dd>{{ incidentStatusLabel(selectedDetail.incident.status) }} / {{ selectedDetail.incident.assignee_id || '未分配' }}</dd><dt>检索请求</dt><dd>{{ selectedDetail.incident.assessment_request_id || '未关联' }}</dd><dt>建议版本</dt><dd>{{ selectedDetail.incident.recommendation_version || '未记录' }}</dd></dl>
          <label>责任人 ID<input v-model.number="assigneeId" type="number" min="1" placeholder="输入用户 ID" /></label><button class="secondary" @click="saveAssignee">保存责任人</button>
          <label>人工结论<textarea v-model="conclusion" rows="2" placeholder="记录人工核验结论" /></label><button class="secondary" @click="saveConclusion">保存结论</button>
          <label>时间线备注 / 附件引用<textarea v-model="timelineNote" rows="2" placeholder="备注内容（可附本地演示资料引用）" /></label><input v-model="attachmentRef" placeholder="附件引用（可选，不上传文件）" /><button class="secondary" @click="addNote">添加时间线</button>
          <h4>证据快照</h4><pre>{{ formattedEvidence }}</pre>
          <h4>处理时间线</h4><ol class="timeline"><li v-for="event in selectedDetail.timeline" :key="event.id"><strong>{{ event.event_type }}</strong> · {{ event.created_at }}<span v-if="event.from_status || event.to_status">：{{ event.from_status }} → {{ event.to_status }}</span><p v-if="event.content">{{ event.content }}</p><p v-if="event.attachment_ref">附件：{{ event.attachment_ref }}</p></li></ol>
          <h4>模拟处置草稿</h4><div class="draft-actions"><select v-model="draftAction"><option value="callback_review">建议回调复核</option><option value="manual_escalation">建议人工升级</option></select><input v-model="draftKey" placeholder="幂等键（至少 8 位）" /><button class="secondary" @click="createDraft">生成待审批草稿</button></div><p v-if="draftMessage" class="case-message">{{ draftMessage }}</p><div v-if="draft"><p>{{ draft.action_type }} · {{ draft.status }} · {{ draft.idempotency_key }}</p><button class="secondary" :disabled="draft.status !== 'pending_approval'" @click="decideDraft(true)">批准</button><button class="secondary" :disabled="draft.status !== 'pending_approval'" @click="decideDraft(false)">拒绝</button><button class="secondary" :disabled="draft.status !== 'approved'" @click="executeDraft">模拟执行</button><p v-if="draft.execution_result">{{ draft.execution_result }}</p></div>
          </section>
        </template>
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
import type { PaymentDispositionDraft, PaymentIncident, PaymentIncidentDetail, PaymentIncidentStatus, SimulatedCallbackLog, SimulatedPaymentOrder, SimulatedPaymentRecord } from '@/types'
import { buildEvidenceBoundRecommendation, buildInvestigationChecklist, paymentStarterQuestions, relativeMatchPercent, splitByQueryTerms } from '@/features/payment-incident/investigation'
import { availableIncidentTransitions, extractOrderID, incidentStatusLabel, inferIncidentType } from '@/features/payment-incident/cases'
import { paymentIncidentApi } from '@/api/payment-incident'

const store = useKnowledgeStore()
const { knowledgeBases } = storeToRefs(store)
const query = ref('')
const kbId = ref('')
const results = ref<SearchResult[]>([])
const loading = ref(false)
const searched = ref(false)
const error = ref('')
const incidents = ref<PaymentIncident[]>([])
const creatingCase = ref(false)
const loadingCases = ref(false)
const transitioningId = ref('')
const caseMessage = ref('')
const selectedDetail = ref<PaymentIncidentDetail | null>(null)
const detailLoadingId = ref('')
const assigneeId = ref<number | null>(null)
const conclusion = ref('')
const timelineNote = ref('')
const attachmentRef = ref('')
const toolFacts = ref<{ order: SimulatedPaymentOrder; payments: SimulatedPaymentRecord[]; callbacks: SimulatedCallbackLog[] } | null>(null)
const toolFactError = ref('')
const draftAction = ref('callback_review'), draftKey = ref(''), draft = ref<PaymentDispositionDraft | null>(null), draftMessage = ref('')
const checklist = computed(() => buildInvestigationChecklist(query.value, results.value))
const recommendation = computed(() => buildEvidenceBoundRecommendation(query.value, results.value))
const scopedKbIds = computed(() => kbId.value ? [kbId.value] : knowledgeBases.value.map(kb => kb.id))
const maxScore = computed(() => Math.max(0, ...results.value.map(item => item.score)))
const formattedEvidence = computed(() => {
  if (!selectedDetail.value?.incident.evidence_snapshot) return '未保存证据快照'
  try { return JSON.stringify(JSON.parse(selectedDetail.value.incident.evidence_snapshot), null, 2) } catch { return selectedDetail.value.incident.evidence_snapshot }
})

onMounted(() => { void store.loadKnowledgeBases(); void loadCases() })

async function loadCases() {
  loadingCases.value = true
  try { const response = await paymentIncidentApi.list({ page: 1, page_size: 20 }); incidents.value = response.data?.items ?? [] } finally { loadingCases.value = false }
}

async function createCase() {
  creatingCase.value = true; caseMessage.value = ''
  try {
    const searchResponseID = lastAssessmentRequestID.value
    const response = await paymentIncidentApi.create({ order_id: extractOrderID(query.value), incident_type: inferIncidentType(query.value), priority: 'P2', assessment_request_id: searchResponseID, recommendation_version: 'evidence-bound-v1', evidence_snapshot: { query: query.value.trim(), evidence: results.value, tool_facts: toolFacts.value, recommendation: recommendation.value } })
    caseMessage.value = `案件 ${response.data?.id ?? ''} 已创建，默认状态为待核验。`; await loadCases()
  } catch (e) { caseMessage.value = e instanceof Error ? e.message : '案件创建失败。' } finally { creatingCase.value = false }
}

const lastAssessmentRequestID = ref('')
async function loadDetail(id: string) {
  detailLoadingId.value = id
  caseMessage.value = ''
  try { const response = await paymentIncidentApi.get(id); selectedDetail.value = response.data ?? null; assigneeId.value = selectedDetail.value?.incident.assignee_id ?? null; conclusion.value = selectedDetail.value?.incident.conclusion ?? '' } catch (e) { caseMessage.value = e instanceof Error ? e.message : '案件详情加载失败。' } finally { detailLoadingId.value = '' }
}
async function saveAssignee() { if (!selectedDetail.value || !assigneeId.value) return; await paymentIncidentApi.update(selectedDetail.value.incident.id, { assignee_id: assigneeId.value }); await loadDetail(selectedDetail.value.incident.id) }
async function saveConclusion() { if (!selectedDetail.value) return; await paymentIncidentApi.update(selectedDetail.value.incident.id, { conclusion: conclusion.value }); await loadDetail(selectedDetail.value.incident.id) }
async function addNote() { if (!selectedDetail.value || !timelineNote.value.trim()) return; await paymentIncidentApi.addTimelineNote(selectedDetail.value.incident.id, { content: timelineNote.value, attachment_ref: attachmentRef.value }); timelineNote.value = ''; attachmentRef.value = ''; await loadDetail(selectedDetail.value.incident.id) }
async function createDraft() { if (!selectedDetail.value) return; try { const r = await paymentIncidentApi.createDispositionDraft(selectedDetail.value.incident.id, { action_type: draftAction.value, idempotency_key: draftKey.value }); draft.value = r.data ?? null; draftMessage.value = '已生成待审批草稿。' } catch (e) { draftMessage.value = e instanceof Error ? e.message : '草稿创建失败。' } }
async function decideDraft(approve: boolean) { if (!draft.value) return; try { const r = await paymentIncidentApi.decideDispositionDraft(draft.value.id, approve); draft.value = r.data ?? draft.value; draftMessage.value = approve ? '草稿已批准。' : '草稿已拒绝。' } catch (e) { draftMessage.value = e instanceof Error ? e.message : '审批失败。' } }
async function executeDraft() { if (!draft.value) return; try { const r = await paymentIncidentApi.executeDispositionDraft(draft.value.id); draft.value = r.data ?? draft.value; draftMessage.value = '模拟执行已记录。' } catch (e) { draftMessage.value = e instanceof Error ? e.message : '执行失败。' } }

async function transitionCase(id: string, status: PaymentIncidentStatus) {
  transitioningId.value = id; caseMessage.value = ''
  try { await paymentIncidentApi.transition(id, status); caseMessage.value = `案件已更新为${incidentStatusLabel(status)}。`; await loadCases() } catch (e) { caseMessage.value = e instanceof Error ? e.message : '案件状态更新失败。' } finally { transitioningId.value = '' }
}

async function search() {
  if (scopedKbIds.value.length === 0) {
    error.value = '当前没有可检索的知识库，请先创建并导入支付处置文档。'
    searched.value = false
    return
  }
  loading.value = true
  error.value = ''
  toolFacts.value = null
  toolFactError.value = ''
  try {
    const response = await knowledgeApi.search({
      query: query.value.trim(), kb_ids: scopedKbIds.value, top_k: 5, retrieval_mode: 'hybrid', save_assessment: true
    })
    results.value = response.data?.items ?? []
    lastAssessmentRequestID.value = response.request_id ?? ''
    const orderID = extractOrderID(query.value)
    if (orderID) {
      try {
        const [order, payments, callbacks] = await Promise.all([paymentIncidentApi.getSimulatedOrder(orderID), paymentIncidentApi.getSimulatedPayments(orderID), paymentIncidentApi.getSimulatedCallbackLogs(orderID)])
        if (order.data) toolFacts.value = { order: order.data, payments: payments.data?.items ?? [], callbacks: callbacks.data?.items ?? [] }
      } catch { toolFactError.value = '未找到该订单的模拟工具事实；请仅根据可追溯知识证据人工核验。' }
    }
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
.query-card, .checklist, .recommendation, .tool-facts, .case-panel, .evidence { padding: 20px; border: 1px solid var(--color-border-subtle); border-radius: var(--radius-md); background: var(--bg-tertiary); }
label { display: block; font-weight: 600; margin-bottom: 8px; } textarea, select { box-sizing: border-box; width: 100%; border: 1px solid var(--color-border-subtle); border-radius: 6px; padding: 10px; background: var(--bg-primary); color: var(--text-primary); font: inherit; }
.actions { display: flex; gap: 10px; margin-top: 12px; } .actions select { flex: 1; } button { border: 0; border-radius: 6px; padding: 9px 13px; cursor: pointer; background: var(--primary); color: var(--on-primary); font: inherit; } button:disabled { cursor: not-allowed; opacity: .6; }
.quick-questions { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 14px; } .quick-questions button { background: var(--bg-elevated); color: var(--text-secondary); font-size: 12px; }
.checklist, .recommendation, .tool-facts, .case-panel, .evidence { margin-top: 18px; } ol { margin-bottom: 0; padding-left: 20px; } li + li { margin-top: 8px; } .evidence-title, .recommendation-title, .case-title, .incident-card > div, .evidence-card > div { display: flex; justify-content: space-between; gap: 12px; } .evidence-title span, .recommendation-title span, .incident-card span, .evidence-card span { color: var(--text-muted); font: 12px var(--font-mono); }
.case-title h2, .case-title h3, .case-title p { margin: 0; } .case-title p, .incident-card p, .case-message { color: var(--text-muted); font-size: 13px; } .case-list-title { margin-top: 18px; align-items: center; } .secondary { background: var(--bg-elevated); color: var(--text-secondary); } .incident-card { margin-top: 10px; padding: 12px; background: var(--bg-primary); border-left: 3px solid var(--primary); } .incident-card p { margin: 6px 0; font-family: var(--font-mono); } .transition-actions { display: flex; gap: 8px; flex-wrap: wrap; } .transition-actions button { font-size: 12px; padding: 6px 9px; }
.recommendation dl, .tool-facts dl, .incident-detail dl { display: grid; grid-template-columns: 112px 1fr; gap: 12px 16px; margin: 16px 0 0; } .recommendation dt, .tool-facts dt, .incident-detail dt { color: var(--text-muted); font-weight: 600; } .recommendation dd, .tool-facts dd, .incident-detail dd { margin: 0; line-height: 1.6; } .recommendation ul, .tool-facts ul { margin: 0; padding-left: 20px; } .risk-notice { color: var(--danger); }
.incident-detail { margin-top: 12px; padding: 14px; background: var(--bg-primary); border-left: 3px solid var(--primary); } .incident-detail label { margin-top: 12px; } .incident-detail input, .incident-detail textarea { box-sizing: border-box; width: 100%; margin: 6px 0; border: 1px solid var(--color-border-subtle); border-radius: 6px; padding: 8px; background: var(--bg-primary); color: var(--text-primary); font: inherit; } .incident-detail pre { max-height: 220px; overflow: auto; white-space: pre-wrap; background: var(--bg-tertiary); padding: 10px; } .timeline { padding-left: 20px; } .timeline p { margin: 4px 0; color: var(--text-muted); }
.evidence-card { margin-top: 12px; padding: 14px; border-left: 3px solid var(--primary); background: var(--bg-primary); } .evidence-card p { margin-top: 8px; white-space: pre-wrap; line-height: 1.65; } mark { padding: 0 2px; border-radius: 2px; background: rgba(156, 180, 205, .24); color: inherit; } .empty, .error { margin-top: 16px; color: var(--text-muted); } .error { color: var(--danger); }
@media (max-width: 640px) { .actions { flex-direction: column; } .recommendation dl { grid-template-columns: 1fr; gap: 4px; } }
</style>
