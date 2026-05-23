<template>
  <section class="card panel expenses-page optimized-expenses">
    <div class="section-head dense-head">
      <div>
        <h2>Расходы</h2>
        <p class="muted">Расход можно привязать к маршруту/точке, указать нескольких плательщиков и распределить сумму поровну или вручную.</p>
      </div>
      <button class="button small secondary" @click="load">Обновить</button>
    </div>

    <div v-if="preselectedLocation" class="success inline-notice compact-notice">
      Расход будет привязан к точке: <strong>{{ preselectedLocation.name }}</strong>
      <span v-if="preselectedLocation.route_title"> · {{ preselectedLocation.route_title }}</span>
    </div>

    <div v-if="summary" class="summary-grid compact-summary">
      <div class="subcard summary-card">
        <span class="muted">Всего расходов</span>
        <strong>{{ money(summary.total_amount, summary.currency) }}</strong>
      </div>
      <div class="subcard summary-card">
        <span class="muted">Оплачено</span>
        <strong>{{ money(summary.total_paid, summary.currency) }}</strong>
      </div>
      <div class="subcard summary-card">
        <span class="muted">Распределено</span>
        <strong>{{ money(summary.total_shares, summary.currency) }}</strong>
      </div>
      <div class="subcard summary-card">
        <span class="muted">Записей</span>
        <strong>{{ summary.expenses.length }}</strong>
      </div>
    </div>

    <div class="expense-layout expense-dashboard">
      <form v-if="canEdit" class="form compact expense-form wide-form expense-form-pro" @submit.prevent="createExpense">
        <div class="form-title-row">
          <h3>Добавить расход</h3>
          <span class="badge">{{ form.route_id ? 'Расход маршрута' : 'Общий расход' }}</span>
        </div>

        <div class="expense-form-grid">
          <div class="form-block expense-main-fields">
            <label>Описание расхода<input v-model="form.description" placeholder="Например, гостиница" required /></label>
            <div class="two-cols compact-fields">
              <label>Сумма<input v-model.number="form.amount" type="number" min="0.01" step="0.01" placeholder="Сумма" required @input="syncDefaultAmounts" /></label>
              <label>Время расхода<input v-model="form.expense_at" type="datetime-local" /></label>
            </div>
            <div class="two-cols compact-fields">
              <label>Маршрут / день
                <select v-model="form.route_id" @change="handleRouteChanged">
                  <option value="">Общий расход поездки</option>
                  <option v-for="route in sortedRoutes" :key="route.id" :value="route.id">
                    {{ route.title }}{{ route.route_date ? ` — ${formatDate(route.route_date)}` : '' }}
                  </option>
                </select>
              </label>
              <label>Связанная точка
                <select v-model="form.location_id" @change="handleLocationChanged">
                  <option value="">Не привязывать к точке</option>
                  <option v-for="location in availableLocations" :key="location.id" :value="location.id">
                    {{ location.name }}{{ location.visit_at ? ` — ${formatVisitShort(location.visit_at)}` : '' }}
                  </option>
                </select>
              </label>
            </div>

            <div v-if="error" class="error error-box">{{ error }}</div>
          </div>

          <div class="form-stack money-stack">
            <div class="subcard form-block compact-money-block">
              <div class="section-head tiny-head">
                <div>
                  <h3>Кто оплатил</h3>
                  <p class="hint">Сумма оплат должна совпасть с общей суммой расхода.</p>
                </div>
                <button type="button" class="button tiny secondary" @click="splitPaymentsEqually">Поровну</button>
              </div>
              <div class="money-rows compact-money-rows">
                <label v-for="row in paymentRows" :key="row.user_id" class="money-row compact-money-row">
                  <input v-model="row.enabled" type="checkbox" @change="syncPaymentRow(row)" />
                  <span>{{ row.name }}</span>
                  <input v-model.number="row.amount" type="number" min="0" step="0.01" :disabled="!row.enabled" @input="row.enabled = Number(row.amount || 0) > 0" />
                </label>
              </div>
              <p :class="paymentDelta === 0 ? 'hint' : 'error'">
                Оплачено: {{ money(paymentTotal) }} · Расход: {{ money(form.amount) }}
                <span v-if="paymentDelta !== 0"> · Разница: {{ money(Math.abs(paymentDelta)) }}</span>
              </p>
            </div>

            <div class="subcard form-block compact-money-block">
              <div class="section-head tiny-head">
                <div>
                  <h3>На кого делится</h3>
                  <p class="hint">Можно разделить поровну или вписать точные суммы вручную.</p>
                </div>
              </div>
              <div class="mode-switch compact-mode-switch">
                <label><input v-model="form.split_mode" type="radio" value="equal" @change="recalculateShares" /> Поровну</label>
                <label><input v-model="form.split_mode" type="radio" value="manual" @change="recalculateShares" /> Ручные суммы</label>
              </div>
              <div class="money-rows compact-money-rows">
                <label v-for="row in shareRows" :key="row.user_id" class="money-row compact-money-row">
                  <input v-model="row.enabled" type="checkbox" @change="syncShareRow(row)" />
                  <span>{{ row.name }}</span>
                  <input v-model.number="row.amount" type="number" min="0" step="0.01" :disabled="!row.enabled || form.split_mode !== 'manual'" @input="row.enabled = Number(row.amount || 0) > 0" />
                </label>
              </div>
              <p :class="shareDelta === 0 ? 'hint' : 'error'">
                Доли: {{ money(shareTotal) }} · Расход: {{ money(form.amount) }}
                <span v-if="shareDelta !== 0"> · Разница: {{ money(Math.abs(shareDelta)) }}</span>
              </p>
            </div>
          </div>
        </div>

        <div class="expense-submit-row">
          <p class="hint">Проверьте сумму оплат и распределения: обе суммы должны совпадать с общим расходом.</p>
          <button class="button full compact-submit" :disabled="saving">{{ saving ? 'Сохраняем...' : 'Добавить расход' }}</button>
        </div>
      </form>
      <div v-else class="subcard">
        <h3>Только просмотр</h3>
        <p class="muted">У вас роль viewer, поэтому расходы можно смотреть, но нельзя добавлять.</p>
      </div>

      <aside class="expense-side">
        <div class="subcard explainer compact-explainer">
          <h3>Как читать итоги</h3>
          <p><strong>Вложил</strong> — сколько человек реально оплатил.</p>
          <p><strong>Его доля</strong> — сколько расходов должно лечь на него.</p>
          <p><strong>Баланс</strong> — нажмите на сумму, чтобы увидеть детализацию переводов.</p>
        </div>

        <div v-if="summary?.settlements.length" class="subcard settlement-card compact-card">
          <h3>Кто кому должен</h3>
          <p v-for="s in summary.settlements" :key="`${s.from_user_id}-${s.to_user_id}`" class="settlement">
            {{ s.from_name }} → {{ s.to_name }}: <strong>{{ money(s.amount, s.currency) }}</strong>
          </p>
        </div>
        <p v-else-if="summary" class="empty compact-empty">По текущим расходам никто никому не должен.</p>
      </aside>
    </div>

    <details v-if="summary?.route_summaries.length" class="subcard table-card compact-details">
      <summary>Расходы по маршрутам</summary>
      <div class="expense-table compact-table">
        <div class="table-row table-head">
          <span>Маршрут</span>
          <span>Дата</span>
          <span>Записей</span>
          <span>Сумма</span>
        </div>
        <div v-for="route in summary.route_summaries" :key="route.route_id || 'general'" class="table-row">
          <strong>{{ route.route_title }}</strong>
          <span>{{ route.route_date ? formatDate(route.route_date) : '—' }}</span>
          <span>{{ route.expenses_count }}</span>
          <span>{{ money(route.total_amount, route.currency) }}</span>
        </div>
      </div>
    </details>

    <div v-if="summary?.participants.length" class="subcard table-card participants-card">
      <h3>Сводка по участникам</h3>
      <div class="expense-table">
        <div class="table-row table-head">
          <span>Участник</span>
          <span>Вложил в поездку</span>
          <span>Его доля расходов</span>
          <span>Баланс</span>
        </div>
        <div v-for="p in summary.participants" :key="p.user_id" class="table-row">
          <strong>{{ p.display_name }}</strong>
          <span>{{ money(p.paid_total, p.currency) }}</span>
          <span>{{ money(p.share_total, p.currency) }}</span>
          <button type="button" class="balance-button" :class="p.net_balance >= 0 ? 'positive' : 'negative'" @click="openDebtDetails(p)">
            {{ balanceLabel(p.net_balance, p.currency) }}
          </button>
        </div>
      </div>
    </div>

    <details v-if="summary" class="subcard expense-history compact-details">
      <summary>История расходов — {{ summary.expenses.length }}</summary>
      <p v-if="summary.expenses.length === 0" class="empty">Расходов пока нет.</p>
      <article v-for="expense in summary.expenses" :key="expense.id" class="expense-item">
        <div>
          <strong>{{ expense.description }}</strong>
          <span>{{ money(expense.amount, expense.currency) }}</span>
        </div>
        <small>Маршрут: {{ expense.route_title || 'общий расход поездки' }}</small>
        <small v-if="expense.location_id" class="location-link">Точка: {{ expense.location_name || locationNameById(expense.location_id) }}</small>
        <small v-if="expense.expense_at">Время: {{ formatVisitShort(expense.expense_at) }}</small>
        <small>Оплатили: {{ expense.payments.map((p) => `${p.display_name || nameById(p.user_id)} — ${money(p.amount, expense.currency)}`).join(', ') }}</small>
        <small>Доли: {{ expense.shares.map((p) => `${p.display_name || nameById(p.user_id)} — ${money(p.amount, expense.currency)}`).join(', ') }}</small>
      </article>
    </details>

    <div v-if="detailsParticipant && summary" class="modal-backdrop" @click.self="detailsParticipant = null">
      <div class="modal card debt-modal">
        <h2>{{ detailsParticipant.display_name }}</h2>
        <p class="muted">{{ balanceLabel(detailsParticipant.net_balance, detailsParticipant.currency) }}</p>
        <div v-if="participantSettlements(detailsParticipant.user_id).length" class="settlement-card">
          <p v-for="s in participantSettlements(detailsParticipant.user_id)" :key="`${s.from_user_id}-${s.to_user_id}`" class="settlement">
            <template v-if="s.from_user_id === detailsParticipant.user_id">
              {{ detailsParticipant.display_name }} должен {{ s.to_name }} — <strong>{{ money(s.amount, s.currency) }}</strong>
            </template>
            <template v-else>
              {{ s.from_name }} должен {{ detailsParticipant.display_name }} — <strong>{{ money(s.amount, s.currency) }}</strong>
            </template>
          </p>
        </div>
        <p v-else class="empty">По этому участнику нет переводов.</p>
        <div class="actions right">
          <button type="button" class="button secondary" @click="detailsParticipant = null">Закрыть</button>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { http } from '../api/http'
import type { ExpenseParticipantSummary, ExpensesSummary, LocationPoint, TripMember, TripRoute } from '../types'

interface MoneyRow {
  user_id: string
  name: string
  enabled: boolean
  amount: number
}

const props = defineProps<{
  tripId: string
  members: TripMember[]
  routes: TripRoute[]
  locations: LocationPoint[]
  selectedRouteId: string
  reloadKey: number
  canEdit: boolean
  preselectedLocationId?: string | null
}>()
const emit = defineEmits<{ created: [] }>()

const summary = ref<ExpensesSummary | null>(null)
const saving = ref(false)
const error = ref('')
const detailsParticipant = ref<ExpenseParticipantSummary | null>(null)
const form = reactive({ description: '', amount: 0, route_id: '', location_id: '', expense_at: '', split_mode: 'equal' as 'equal' | 'manual' })
const paymentRows = ref<MoneyRow[]>([])
const shareRows = ref<MoneyRow[]>([])
const sortedRoutes = computed(() => [...props.routes].sort(compareRoutes))
const preselectedLocation = computed(() => props.preselectedLocationId ? props.locations.find((location) => location.id === props.preselectedLocationId) || null : null)
const availableLocations = computed(() => {
  if (!form.route_id) return props.locations.filter((location) => !location.route_id).sort(compareLocations)
  return props.locations.filter((location) => location.route_id === form.route_id).sort(compareLocations)
})
const paymentTotal = computed(() => roundMoney(paymentRows.value.filter((row) => row.enabled).reduce((sum, row) => sum + Number(row.amount || 0), 0)))
const shareTotal = computed(() => roundMoney(shareRows.value.filter((row) => row.enabled).reduce((sum, row) => sum + Number(row.amount || 0), 0)))
const paymentDelta = computed(() => roundMoney(paymentTotal.value - Number(form.amount || 0)))
const shareDelta = computed(() => roundMoney(shareTotal.value - Number(form.amount || 0)))

async function load() {
  try {
    const { data } = await http.get<ExpensesSummary>(`/trips/${props.tripId}/expenses`)
    summary.value = data
  } catch (e: any) {
    error.value = formatApiError(e, 'Не удалось загрузить расходы')
  }
}

function validateForm() {
  if (!form.description.trim()) return 'Введите описание расхода'
  if (!form.amount || form.amount <= 0) return 'Сумма должна быть больше нуля'
  const payments = selectedPayments()
  const shares = selectedShares()
  if (payments.length === 0) return 'Укажите, кто и сколько оплатил'
  if (shares.length === 0) return 'Укажите, между кем делится расход'
  if (toCents(paymentTotal.value) !== toCents(form.amount)) {
    return `Сумма оплат не совпадает с общей суммой расхода. Общий расход: ${money(form.amount)}. Оплачено участниками: ${money(paymentTotal.value)}. Разница: ${money(Math.abs(paymentDelta.value))}.`
  }
  if (toCents(shareTotal.value) !== toCents(form.amount)) {
    return `Сумма долей не совпадает с общей суммой расхода. Общий расход: ${money(form.amount)}. Сумма долей: ${money(shareTotal.value)}. Разница: ${money(Math.abs(shareDelta.value))}.`
  }
  return ''
}

async function createExpense() {
  recalculateShares()
  error.value = validateForm()
  if (error.value) return
  saving.value = true
  try {
    await http.post(`/trips/${props.tripId}/expenses`, {
      description: form.description.trim(),
      amount: roundMoney(form.amount),
      currency: 'RUB',
      split_mode: form.split_mode,
      payments: selectedPayments(),
      shares: selectedShares(),
      route_id: form.route_id || null,
      location_id: form.location_id || null,
      expense_at: form.expense_at ? new Date(form.expense_at).toISOString() : null
    })
    form.description = ''
    form.amount = 0
    form.location_id = ''
    form.expense_at = ''
    initializeRows()
    await load()
    emit('created')
  } catch (e: any) {
    error.value = formatApiError(e, 'Не удалось добавить расход')
  } finally {
    saving.value = false
  }
}

function selectedPayments() {
  return paymentRows.value
    .filter((row) => row.enabled && Number(row.amount || 0) > 0)
    .map((row) => ({ user_id: row.user_id, amount: roundMoney(row.amount) }))
}

function selectedShares() {
  return shareRows.value
    .filter((row) => row.enabled && Number(row.amount || 0) > 0)
    .map((row) => ({ user_id: row.user_id, amount: roundMoney(row.amount) }))
}

function initializeRows() {
  const existingPayment = new Map(paymentRows.value.map((row) => [row.user_id, row]))
  const existingShare = new Map(shareRows.value.map((row) => [row.user_id, row]))
  paymentRows.value = props.members.map((member, index) => {
    const old = existingPayment.get(member.user_id)
    return old ? { ...old, name: member.display_name } : { user_id: member.user_id, name: member.display_name, enabled: index === 0, amount: index === 0 ? Number(form.amount || 0) : 0 }
  })
  shareRows.value = props.members.map((member) => {
    const old = existingShare.get(member.user_id)
    return old ? { ...old, name: member.display_name } : { user_id: member.user_id, name: member.display_name, enabled: true, amount: 0 }
  })
  recalculateShares()
}

function syncDefaultAmounts() {
  const selected = paymentRows.value.filter((row) => row.enabled)
  if (selected.length === 1) selected[0].amount = roundMoney(form.amount)
  recalculateShares()
}

function syncPaymentRow(row: MoneyRow) {
  if (!row.enabled) row.amount = 0
  if (row.enabled && paymentRows.value.filter((item) => item.enabled).length === 1) row.amount = roundMoney(form.amount)
}

function syncShareRow(row: MoneyRow) {
  if (!row.enabled) row.amount = 0
  if (form.split_mode === 'equal') recalculateShares()
}

function splitPaymentsEqually() {
  const selected = paymentRows.value.filter((row) => row.enabled)
  const rows = selected.length ? selected : paymentRows.value
  distributeAmount(form.amount, rows)
  for (const row of paymentRows.value) row.enabled = rows.some((item) => item.user_id === row.user_id)
}

function recalculateShares() {
  if (form.split_mode === 'manual') return
  const selected = shareRows.value.filter((row) => row.enabled)
  if (selected.length === 0) return
  distributeAmount(form.amount, selected)
  for (const row of shareRows.value) {
    if (!row.enabled) row.amount = 0
  }
}

function distributeAmount(amount: number, rows: MoneyRow[]) {
  if (!rows.length) return
  const totalCents = toCents(amount)
  const base = Math.floor(totalCents / rows.length)
  let rest = totalCents - base * rows.length
  rows.forEach((row) => {
    const cents = base + (rest > 0 ? 1 : 0)
    if (rest > 0) rest--
    row.amount = cents / 100
  })
}

function handleRouteChanged() {
  if (form.location_id && !availableLocations.value.some((location) => location.id === form.location_id)) form.location_id = ''
}

function handleLocationChanged() {
  const location = props.locations.find((item) => item.id === form.location_id)
  if (!location) return
  if (location.route_id) form.route_id = location.route_id
  if (location.visit_at && !form.expense_at) form.expense_at = toDateTimeLocal(location.visit_at)
  if (!form.description.trim()) form.description = `Расход по точке: ${location.name}`
}

function applyPreselectedLocation() {
  const location = preselectedLocation.value
  if (!location) return
  form.location_id = location.id
  form.route_id = location.route_id || ''
  if (location.visit_at) form.expense_at = toDateTimeLocal(location.visit_at)
  if (!form.description.trim()) form.description = `Расход по точке: ${location.name}`
}

function applySelectedRoute() {
  if (props.selectedRouteId !== 'all') form.route_id = props.selectedRouteId
}

function openDebtDetails(participant: ExpenseParticipantSummary) {
  detailsParticipant.value = participant
}

function participantSettlements(userID: string) {
  return summary.value?.settlements.filter((s) => s.from_user_id === userID || s.to_user_id === userID) || []
}

function balanceLabel(value: number, currency = 'RUB') {
  if (Math.abs(value) < 0.005) return 'Расчёт закрыт'
  if (value > 0) return `Ему должны ${money(value, currency)}`
  return `Должен ${money(Math.abs(value), currency)}`
}

function nameById(id: string) {
  return props.members.find((m) => m.user_id === id)?.display_name || 'Участник вне поездки'
}

function locationNameById(id: string) {
  return props.locations.find((location) => location.id === id)?.name || 'Точка удалена'
}

function money(value: number, currency = 'RUB') {
  return `${Number(value || 0).toFixed(2)} ${currency}`
}

function formatVisitShort(value: string) {
  return new Intl.DateTimeFormat('ru-RU', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' }).format(new Date(value))
}

function formatDate(value?: string | null) {
  if (!value) return '—'
  return new Intl.DateTimeFormat('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' }).format(new Date(value))
}

function toDateTimeLocal(value?: string | null) {
  if (!value) return ''
  const date = new Date(value)
  const offsetMs = date.getTimezoneOffset() * 60_000
  return new Date(date.getTime() - offsetMs).toISOString().slice(0, 16)
}

function compareRoutes(a: TripRoute, b: TripRoute) {
  const aDate = a.route_date ? new Date(a.route_date).getTime() : Number.POSITIVE_INFINITY
  const bDate = b.route_date ? new Date(b.route_date).getTime() : Number.POSITIVE_INFINITY
  if (aDate !== bDate) return aDate - bDate
  if (a.sort_order !== b.sort_order) return a.sort_order - b.sort_order
  return new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
}

function compareLocations(a: LocationPoint, b: LocationPoint) {
  const aOrder = a.route_order ?? Number.POSITIVE_INFINITY
  const bOrder = b.route_order ?? Number.POSITIVE_INFINITY
  if (a.route_id && b.route_id && a.route_id === b.route_id && aOrder !== bOrder) return aOrder - bOrder
  const aTime = a.visit_at ? new Date(a.visit_at).getTime() : Number.POSITIVE_INFINITY
  const bTime = b.visit_at ? new Date(b.visit_at).getTime() : Number.POSITIVE_INFINITY
  if (aTime !== bTime) return aTime - bTime
  return new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
}

function roundMoney(value: number) {
  return Math.round(Number(value || 0) * 100) / 100
}

function toCents(value: number) {
  return Math.round(Number(value || 0) * 100)
}

function formatApiError(e: any, fallback: string) {
  const data = e.response?.data
  if (!data?.error) return fallback
  if (!data.details) return data.error
  const labels: Record<string, string> = {
    total_amount: 'общий расход',
    payments_sum: 'сумма оплат',
    shares_sum: 'сумма долей',
    difference: 'разница',
    currency: 'валюта',
    locations_count: 'точек',
    expenses_count: 'расходов'
  }
  const details = Object.entries(data.details).map(([key, value]) => `${labels[key] || key}: ${value}`).join(', ')
  return `${data.error}. ${details}`
}

watch(() => props.members, initializeRows, { immediate: true, deep: true })
watch(() => form.amount, syncDefaultAmounts)
watch(() => form.split_mode, recalculateShares)
watch(() => props.preselectedLocationId, applyPreselectedLocation, { immediate: true })
watch(() => props.selectedRouteId, applySelectedRoute, { immediate: true })
watch(() => props.reloadKey, load)
onMounted(load)
</script>
