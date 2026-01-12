<template>
  <div class="team-poster-page">
    <div class="poster-header">
      <router-link :to="`/teams/${teamId}`" class="back-button">← Back to Team</router-link>
    </div>
    <div class="poster-container">
      <img v-if="teamImageUrl" :src="teamImageUrl" alt="Team Poster" class="poster-image" />
      <div v-else class="no-poster">
        No poster available for this team
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { API_URL } from '@/config'
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'

const api = API_URL
const route = useRoute()

const teamId = computed(() => Number(route.params.id))
const teamImageUrl = ref<string | null>(null)

const loadTeamImage = async (id: number) => {
  try {
    const res = await fetch(`${api}/teams/${id}/image/vertical`)
    if (res.ok) {
      const blob = await res.blob()
      teamImageUrl.value = URL.createObjectURL(blob)
    }
  } catch (e) {
    // 404 is expected for teams without images
  }
}

onMounted(() => {
  loadTeamImage(teamId.value)
})

watch(teamId, () => {
  loadTeamImage(teamId.value)
})
</script>

<style scoped>
.team-poster-page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background-color: var(--clr-surface-a0);
}

.poster-header {
  display: flex;
  align-items: center;
  margin-bottom: var(--spacing-xl);
}

.back-button {
  padding: var(--spacing-md) var(--spacing-lg);
  background-color: var(--clr-surface-a10);
  border-radius: var(--radius);
  color: var(--text-primary);
  text-decoration: none;
  transition: background-color 0.2s ease;
  font-size: 14px;
  font-weight: 500;
}

.back-button:hover {
  background-color: var(--clr-surface-a20);
}

.poster-container {
  display: flex;
  justify-content: center;
  align-items: center;
  flex: 1;
}

.poster-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  border-radius: var(--radius);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.no-poster {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  font-size: 16px;
}

@media (max-width: 768px) {
  .team-poster-page {
    min-height: 100vh;
  }

  .poster-image {
    max-width: 100%;
    height: auto;
  }
}
</style>
