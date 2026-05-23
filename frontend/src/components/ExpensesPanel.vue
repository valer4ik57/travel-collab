<template>
  <section class="card panel expenses-page">
    <div class="section-head">
      <div>
        <h2>Расходы</h2>
        <p class="muted">Здесь считаются не личные траты, а взаиморасчёты: кто оплатил общий расход и кто кому должен после деления.</p>
      </div>
      <button class="button small secondary" @click="load">Обновить</button>
    </div>

    <div class="expense-layout">
      <form v-if="canEdit" class="form compact expense-form" @submit.prevent="createExpense">
        <h3>Добавить расход</h3>
        <label>Описание расхода<input v-model="form.description" placeholder="Например, гостиница" required /></label>
        <div class="two-cols">
          <label>Сумма<input v-model.number="form.amount" type="number" min="0.01" step="0.01" placeholder="Сумма" required /></label>
          <label>Кто оплатил
            <select v-model="form.paid_by" required>
              <option disabled value="">Кто платил</option>
              <option v-for="m in editableMembers" :key="m.user_id" :value="m.user_id">{{ m.display_name }}</option>
            </select>
          </label>
        </div>
        <div>
          <strong>Разделить между:</strong>
          <p class="hint">Если выбрать только плательщика, это личная трата: долг не появится. Если выбрать нескольких участников, сумма делится поровну между ними.</p>
          <div class="checkbox-list better">
            <label v-for="m in editableMembers" :key="m.user_id" class="check-card">
              <input v-model="form.split_with" type="checkbox" :value="m.user_id" />
              <span>{{ m.display_name }}</span>
            </label>
          </div>
        </div>
        <p v-if="error" class="error">{{ error }}</p>
        <button class="button small" :disabled="saving">Добавить расход</button>
      </form>
      <div v-else class="subcard">
        <h3>Только просмотр</h3>
        <p class="muted">У вас роль viewer, поэтому расходы можно смотреть, но нельзя добавлять.</p>
      </div>

      <div class="subcard explainer">
        <h3>Как читать итоги</h3>
        <p><strong>Оплатил</strong> — сколько человек реально заплатил.</p>
        <p><strong>Доля</strong> — сколько на него приходится после деления общих расходов.</p>
        <p><strong>Итог</strong> — разница между оплатой и долей. Плюс — человеку должны, минус — он должен.</p>
      </div>
    </div>

    <div v-if="summary" class="summary-grid">
      <div class="subcard summary-card">
        <span class="muted">Всего расходов</span>
        <strong>{{ money(summary.total_amount, summary.currency) }}</strong>
      </div>
      <div class="subcard summary-card">
        <span class="muted">Записей</span>
        <strong>{{ summary.expenses.length }}</strong>
      </div>
      <div class="subcard summary-card">
        <span class="muted">Расчётов между людьми</span>
        <strong>{{ summary.settlements.length }}</strong>
      </div>
    </div>

    <div v-if="summary?.participants.length" class="subcard table-card">
      <h3>Сводка по участникам</h3>
      <div class="expense-table">
        <div class="table-row table-head">
          <span>Участник</span>
          <span>Оплатил</span>
          <span>Доля</span>
          <span>Итог</span>
        </div>
        <div v-for="p in summary.participants" :key="p.user_id" class="table-row">
          <strong>{{ p.display_name }}</strong>
          <span>{{ money(p.paid_total, p.currency) }}</span>
          <span>{{ money(p.share_total, p.currency) }}</span>
          <span :class="p.net_balance >= 0 ? 'positive' : 'negative'">
            {{ money(p.net_balance, p.currency) }}
          </span>
        </div>
      </div>
    </div>

    <div v-if="summary?.settlements.length" class="subcard settlement-card">
      <h3>Кто кому должен</h3>
      <p v-for="s in summary.settlements" :key="`${s.from_user_id}-${s.to_user_id}`" class="settlement">
        {{ s.from_name }} → {{ s.to_name }}: <strong>{{ money(s.amount, s.currency) }}</strong>
      </p>
    </div>
    <p v-else-if="summary" class="empty">По текущим расходам никто никому не должен.</p>

    <div v-if="summary" class="expense-list wide">
      <h3>История расходов</h3>
      <p v-if="summary.expenses.length === 0" class="empty">Расходов пока нет.</p>
      <article v-for="expense in summary.expenses" :key="expense.id" class="expense-item">
        <div>
          <strong>{{ expense.description }}</strong>
          <span>{{ money(expense.amount, expense.currency) }}</span>
        </div>
        <small>Платил: {{ expense.paid_by_name || nameById(expense.paid_by) }}</small>
        <small>Делится между: {{ expense.split_with.map(nameById).join(', ') || 'не указано' }}</small>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { http } from '../api/http'
import type { ExpensesSummary, TripMember } from '../types'

const props = defineProps<{ tripId: string; members: TripMember[]; reloadKey: number; canEdit: boolean }>()
const summary = ref<ExpensesSummary | null>(null)
const saving = ref(false)
const error = ref('')
const form = reactive({ description: '', amount: 0, paid_by: '', split_with: [] as string[] })
const editableMembers = computed(() => props.members)

function nameById(id: string) {
  return props.members.find((m) => m.user_id === id)?.display_name || 'Участник вне поездки'
}

function money(value: number, currency = 'RUB') {
  return `${Number(value || 0).toFixed(2)} ${currency}`
}

async function load() {
  const { data } = await http.get<ExpensesSummary>(`/trips/${props.tripId}/expenses`)
  summary.value = data
}

function validateForm() {
  if (!form.description.trim()) return 'Введите описание расхода'
  if (!form.amount || form.amount <= 0) return 'Сумма должна быть больше нуля'
  if (!form.paid_by) return 'Выберите, кто оплатил расход'
  if (form.split_with.length === 0) return 'Выберите, между кем делится расход'
  return ''
}

async function createExpense() {
  error.value = validateForm()
  if (error.value) return
  saving.value = true
  try {
    await http.post(`/trips/${props.tripId}/expenses`, {
      description: form.description.trim(),
      amount: form.amount,
      currency: 'RUB',
      paid_by: form.paid_by,
      split_with: form.split_with
    })
    form.description = ''
    form.amount = 0
    await load()
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Не удалось добавить расход'
  } finally {
    saving.value = false
  }
}

watch(() => props.members, () => {
  if (!form.paid_by && props.members[0]) form.paid_by = props.members[0].user_id
  if (form.split_with.length === 0) form.split_with = props.members.map((m) => m.user_id)
}, { immediate: true })
watch(() => props.reloadKey, load)
onMounted(load)
</script>
