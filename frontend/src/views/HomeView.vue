<template>
  <section class="hero card">
    <div>
      <p class="eyebrow">Дипломный проект</p>
      <h1>Совместное планирование путешествий</h1>
      <p class="lead">
        Travel-Collab помогает группе собрать маршрут на карте, вести список точек,
        обсуждать поездку в чате и считать общие расходы.
      </p>
      <div class="actions">
        <RouterLink class="button" to="/trips">Открыть поездки</RouterLink>
        <RouterLink class="button secondary" to="/register">Создать аккаунт</RouterLink>
      </div>
    </div>
    <div class="hero-panel">
      <span>🗺️ Карта</span>
      <span>💬 Чат</span>
      <span>💸 Расходы</span>
      <span>⚡ Real-time</span>
    </div>
  </section>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>
