<template>
  <section class="card panel">
    <div class="section-head">
      <h2>Расходы</h2>
      <button class="button small secondary" @click="load">Обновить</button>
    </div>

    <form v-if="canEdit" class="form compact" @submit.prevent="createExpense">
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
        <p class="hint">Сумма делится поровну между выбранными участниками.</p>
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
    <p v-else class="empty">У вас роль viewer, поэтому расходы доступны только для просмотра.</p>

    <div v-if="summary?.settlements.length" class="subcard settlement-card">
      <h3>Кто кому должен</h3>
      <p v-for="s in summary.settlements" :key="`${s.from_user_id}-${s.to_user_id}`" class="settlement">
        {{ s.from_name }} → {{ s.to_name }}: <strong>{{ s.amount.toFixed(2) }} {{ s.currency }}</strong>
      </p>
    </div>

    <div v-if="summary" class="balances">
      <h3>Балансы</h3>
      <p v-for="b in summary.balances" :key="b.user_id" :class="b.amount >= 0 ? 'positive' : 'negative'">
        {{ b.display_name }}: {{ b.amount.toFixed(2) }} {{ b.currency }}
      </p>
    </div>

    <div class="expense-list">
      <p v-if="!summary?.expenses.length" class="empty">Расходов пока нет.</p>
      <article v-for="expense in summary?.expenses" :key="expense.id" class="expense-item">
        <strong>{{ expense.description }}</strong>
        <span>{{ expense.amount.toFixed(2) }} {{ expense.currency }}</span>
        <small>Платил: {{ expense.paid_by_name || 'участник' }}</small>
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

async function load() {
  const { data } = await http.get<ExpensesSummary>(`/trips/${props.tripId}/expenses`)
  summary.value = data
}

async function createExpense() {
  error.value = ''
  if (!form.description.trim()) {
    error.value = 'Укажи описание расхода'
    return
  }
  if (!form.amount || form.amount <= 0) {
    error.value = 'Сумма должна быть больше нуля'
    return
  }
  if (!form.paid_by) {
    error.value = 'Выбери, кто оплатил расход'
    return
  }
  if (!form.split_with.length) {
    error.value = 'Выбери хотя бы одного участника для разделения суммы'
    return
  }
  saving.value = true
  try {
    await http.post(`/trips/${props.tripId}/expenses`, {
      description: form.description,
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
  if (!form.split_with.length) form.split_with = props.members.map((m) => m.user_id)
}, { immediate: true })
watch(() => props.reloadKey, load)
onMounted(load)
</script>
