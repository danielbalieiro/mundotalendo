import useSWR from 'swr'

const API_URL = process.env.NEXT_PUBLIC_API_URL
const API_KEY = process.env.NEXT_PUBLIC_API_KEY

/**
 * Fetcher function for SWR
 * @param {string} url - API endpoint URL
 * @returns {Promise<Object>} JSON response
 */
const fetcher = async (url) => {
  const res = await fetch(url, {
    headers: {
      'X-API-Key': API_KEY,
    },
  })
  if (!res.ok) {
    throw new Error('Failed to fetch user locations')
  }
  return res.json()
}

/**
 * Hook to fetch user locations from API
 * Returns the most recent location for each user
 * @param {number|string} [year] - Year to fetch locations for (defaults to the current year server-side)
 * @returns {Object} { users, total, isLoading, error }
 */
export function useUserLocations(year) {
  const yearQuery = year ? `?year=${encodeURIComponent(year)}` : ''

  const { data, error, isLoading } = useSWR(
    `${API_URL}/users/locations${yearQuery}`,
    fetcher,
    {
      refreshInterval: 60000, // 60s polling (same as useStats)
      revalidateOnFocus: false,
      dedupingInterval: 10000,
    }
  )

  return {
    users: data?.users || [],
    total: data?.total || 0,
    isLoading,
    error,
  }
}
