<template>
  <section class="hero card product-hero">
    <div>
      <p class="eyebrow">Travel-Collab</p>
      <h1>Планируйте путешествия вместе</h1>
      <p class="lead">
        Собирайте маршрут на карте, обсуждайте поездку с друзьями, сохраняйте важные места
        и сразу считайте общие расходы без таблиц и путаницы.
      </p>
      <div v-if="auth.isAuthenticated" class="actions">
        <RouterLink class="button" to="/trips">Перейти к моим поездкам</RouterLink>
        <button class="button secondary" type="button" @click="scrollToFeatures">Что умеет сервис</button>
      </div>
      <div v-else class="actions">
        <RouterLink class="button" to="/register">Начать планирование</RouterLink>
        <RouterLink class="button secondary" to="/login">Войти</RouterLink>
      </div>
    </div>
    <div class="hero-panel product-panel">
      <span>🗺️ Маршрут на карте</span>
      <span>💬 Чат поездки</span>
      <span>💸 Деление расходов</span>
      <span>⚡ Обновления в реальном времени</span>
    </div>
  </section>

  <section v-if="auth.isAuthenticated" class="card welcome-card">
    <div>
      <p class="eyebrow">С возвращением</p>
      <h2>{{ auth.user?.display_name || 'Пользователь' }}, продолжим планирование?</h2>
      <p class="muted">Откройте список поездок, создайте новый маршрут или вступите в поездку по invite-коду.</p>
    </div>
    <RouterLink class="button" to="/trips">Открыть поездки</RouterLink>
  </section>

  <section ref="featuresEl" class="features-grid">
    <article class="card feature-card">
      <h3>Карта и точки интереса</h3>
      <p class="muted">Добавляйте отели, кафе, вокзалы, достопримечательности и другие места прямо кликом по карте.</p>
    </article>
    <article class="card feature-card">
      <h3>Совместная работа</h3>
      <p class="muted">Все участники видят изменения сразу: новые точки, сообщения, расходы и изменения ролей.</p>
    </article>
    <article class="card feature-card">
      <h3>Расходы без путаницы</h3>
      <p class="muted">Сервис показывает, кто сколько оплатил, какая доля у каждого и кто кому должен по итогу.</p>
    </article>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const featuresEl = ref<HTMLElement | null>(null)

function scrollToFeatures() {
  featuresEl.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}
</script>
