import useSWR from 'swr'

const API_URL = process.env.NEXT_PUBLIC_API_URL
const API_KEY = process.env.NEXT_PUBLIC_API_KEY

/**
 * Fetcher for SWR with API key header
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
    throw new Error('Failed to fetch available years')
  }
  return res.json()
}

/**
 * Hook to fetch the list of years that have registered readings.
 * @returns {Object} { years, total, isLoading, error }
 */
export function useYears() {
  const { data, error, isLoading } = useSWR(
    `${API_URL}/years`,
    fetcher,
    {
      refreshInterval: 60000,
      revalidateOnFocus: false,
      dedupingInterval: 10000,
    }
  )

  return {
    years: data?.years || [],
    total: data?.years?.length || 0,
    isLoading,
    error,
  }
}
