import useSWR from 'swr'

/**
 * Fetcher function for SWR
 * @param {string} url - API endpoint URL
 * @returns {Promise<Object>} JSON response
 */
const fetcher = async (url) => {
  const response = await fetch(url)
  if (!response.ok) {
    throw new Error('Failed to fetch stats')
  }
  return response.json()
}

/**
 * Hook to fetch stats from the API with auto-refresh
 * @param {number} [refreshInterval=15000] - Refresh interval in milliseconds (default: 15s)
 * @returns {Object} SWR response with data, error, and isLoading
 */
export function useStats(refreshInterval = 15000) {
  // Use local API route for development, or external API URL for production
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || '/api'

  const { data, error, isLoading } = useSWR(
    `${apiUrl}/stats`,
    fetcher,
    {
      refreshInterval, // Auto-refresh every 15 seconds
      revalidateOnFocus: true,
      revalidateOnReconnect: true,
      dedupingInterval: 10000, // Prevent duplicate requests within 10s
    }
  )

  return {
    countries: data?.countries || [],
    total: data?.total || 0,
    isLoading,
    error,
  }
}
