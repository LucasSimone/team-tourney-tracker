import { ref } from 'vue'
import { API_URL } from '@/config'

const teamImages = ref<{ [key: number]: string }>({})

export const useTeamImages = () => {
  const api = API_URL

  const loadTeamImage = async (teamId: number) => {
    try {
      const res = await fetch(`${api}/teams/${teamId}/image`)
      if (res.ok) {
        const blob = await res.blob()
        teamImages.value[teamId] = URL.createObjectURL(blob)
      }
    } catch (e) {
      console.error(`Failed to load team image for team ${teamId}`, e)
    }
  }

  const getCardStyle = (teamId: number) => {
    const imageUrl = teamImages.value[teamId]
    if (!imageUrl) {
      return {}
    }
    return {
      backgroundImage: `linear-gradient(135deg, rgba(26, 26, 46, 0.85) 0%, rgba(22, 33, 62, 0.85) 100%), url(${imageUrl})`,
      backgroundSize: 'cover',
      backgroundPosition: 'center'
    }
  }

  return {
    teamImages,
    loadTeamImage,
    getCardStyle
  }
}
