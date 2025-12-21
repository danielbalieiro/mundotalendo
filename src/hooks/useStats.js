import useSWR from 'swr'

/**
 * Fetcher function for SWR with retry logic and timeout
 * @param {string} url - API endpoint URL
 * @returns {Promise<Object>} JSON response
 */
const fetcher = async (url) => {
  const maxRetries = 3
  const apiKey = process.env.NEXT_PUBLIC_API_KEY

  for (let i = 0; i < maxRetries; i++) {
    try {
      const headers = {}
      if (apiKey) {
        headers['X-API-Key'] = apiKey
      }

      const response = await fetch(url, {
        signal: AbortSignal.timeout(10000), // 10s timeout
        headers,
      })

      if (!response.ok) {
        // Rate limited - wait and retry
        if (response.status === 429) {
          await new Promise(r => setTimeout(r, 2000 * (i + 1)))
          continue
        }
        throw new Error(`HTTP ${response.status}`)
      }

      return response.json()
    } catch (err) {
      // Last attempt - throw error
      if (i === maxRetries - 1) throw err

      // Exponential backoff: 1s, 2s, 4s
      await new Promise(r => setTimeout(r, 1000 * Math.pow(2, i)))
    }
  }
}

/**
 * Hook to fetch stats from the API with auto-refresh
 * @param {number} [refreshInterval=60000] - Refresh interval in milliseconds (default: 60s)
 * @returns {Object} SWR response with data, error, and isLoading
 */
export function useStats(refreshInterval = 60000) {
  // Use local API route for development, or external API URL for production
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || '/api'

  const { data, error, isLoading } = useSWR(
    `${apiUrl}/stats`,
    fetcher,
    {
      refreshInterval, // Auto-refresh every 60 seconds (reduced from 15s)
      revalidateOnFocus: false, // Don't refetch when tab is focused (reduces unnecessary requests)
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
