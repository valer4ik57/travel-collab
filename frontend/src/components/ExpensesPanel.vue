<template>
  <section class="card panel">
    <div class="section-head">
      <h2>Расходы</h2>
      <button class="button small secondary" @click="load">Обновить</button>
    </div>

    <form class="form compact" @submit.prevent="createExpense">
      <input v-model="form.description" placeholder="Описание" required />
      <div class="two-cols">
        <input v-model.number="form.amount" type="number" min="1" step="0.01" placeholder="Сумма" required />
        <select v-model="form.paid_by" required>
          <option disabled value="">Кто платил</option>
          <option v-for="m in members" :key="m.user_id" :value="m.user_id">{{ m.display_name }}</option>
        </select>
      </div>
      <div class="checkbox-list">
        <label v-for="m in members" :key="m.user_id">
          <input v-model="form.split_with" type="checkbox" :value="m.user_id" /> {{ m.display_name }}
        </label>
      </div>
      <button class="button small">Добавить расход</button>
    </form>

    <div v-if="summary?.settlements.length" class="subcard">
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
import { onMounted, reactive, ref, watch } from 'vue'
import { http } from '../api/http'
import type { ExpensesSummary, TripMember } from '../types'

const props = defineProps<{ tripId: string; members: TripMember[]; reloadKey: number }>()
const summary = ref<ExpensesSummary | null>(null)
const form = reactive({ description: '', amount: 0, paid_by: '', split_with: [] as string[] })

async function load() {
  const { data } = await http.get<ExpensesSummary>(`/trips/${props.tripId}/expenses`)
  summary.value = data
}

async function createExpense() {
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
}

watch(() => props.members, () => {
  if (!form.paid_by && props.members[0]) form.paid_by = props.members[0].user_id
  if (!form.split_with.length) form.split_with = props.members.map((m) => m.user_id)
}, { immediate: true })
watch(() => props.reloadKey, load)
onMounted(load)
</script>
