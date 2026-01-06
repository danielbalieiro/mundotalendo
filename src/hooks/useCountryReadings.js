import { useState, useCallback } from 'react';

/**
 * Hook to fetch readings for a specific country
 * @returns {Object} { fetchReadings, readings, loading, error }
 */
export default function useCountryReadings() {
  const [readings, setReadings] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const fetchReadings = useCallback(async (iso3) => {
    setLoading(true);
    setError(null);
    setReadings([]);

    try {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || '/api';
      const response = await fetch(`${apiUrl}/readings/${iso3}`, {
        headers: {
          'X-API-Key': process.env.NEXT_PUBLIC_API_KEY || '',
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      setReadings(data.readings || []);
    } catch (err) {
      console.error('Error fetching country readings:', err);
      setError(err.message);
      setReadings([]);
    } finally {
      setLoading(false);
    }
  }, []);

  return { fetchReadings, readings, loading, error };
}
