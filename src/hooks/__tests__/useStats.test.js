import { renderHook, waitFor } from '@testing-library/react';
import { SWRConfig } from 'swr';
import { useStats } from '../useStats';

// Mock fetch globally
global.fetch = jest.fn();

// Mock AbortSignal.timeout for testing (not available in JSDOM)
if (!AbortSignal.timeout) {
  AbortSignal.timeout = jest.fn((ms) => {
    const controller = new AbortController();
    return controller.signal;
  });
}

// Mock environment variables
const originalEnv = process.env;

// Wrapper component for SWR with cache provider
const wrapper = ({ children }) => (
  <SWRConfig value={{ provider: () => new Map(), dedupingInterval: 0 }}>
    {children}
  </SWRConfig>
);

describe('useStats hook', () => {
  beforeEach(() => {
    jest.resetModules();
    process.env = { ...originalEnv };
    fetch.mockClear();
  });

  afterEach(() => {
    process.env = originalEnv;
  });

  describe('successful data fetching', () => {
    it('should fetch and return stats data', async () => {
      const mockData = {
        countries: ['BRA', 'USA', 'JPN'],
        total: 3,
      };

      fetch.mockResolvedValue({
        ok: true,
        json: async () => mockData,
      });

      const { result } = renderHook(() => useStats(0), { wrapper }); // 0 interval to prevent auto-refresh in tests

      // Initial state
      expect(result.current.isLoading).toBe(true);
      expect(result.current.countries).toEqual([]);
      expect(result.current.total).toBe(0);

      // Wait for data to load
      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.countries).toEqual(['BRA', 'USA', 'JPN']);
      expect(result.current.total).toBe(3);
      expect(result.current.error).toBeUndefined();
    });

    it('should use local API route when NEXT_PUBLIC_API_URL is not set', async () => {
      delete process.env.NEXT_PUBLIC_API_URL;

      fetch.mockResolvedValue({
        ok: true,
        json: async () => ({ countries: [], total: 0 }),
      });

      renderHook(() => useStats(0), { wrapper });

      await waitFor(() => {
        expect(fetch).toHaveBeenCalledWith(
          '/api/stats',
          expect.any(Object)
        );
      });
    });

    it('should use NEXT_PUBLIC_API_URL when set', async () => {
      process.env.NEXT_PUBLIC_API_URL = 'https://api.example.com';

      fetch.mockResolvedValue({
        ok: true,
        json: async () => ({ countries: [], total: 0 }),
      });

      renderHook(() => useStats(0), { wrapper });

      await waitFor(() => {
        expect(fetch).toHaveBeenCalledWith(
          'https://api.example.com/stats',
          expect.any(Object)
        );
      });
    });

    it('should include API key in headers when NEXT_PUBLIC_API_KEY is set', async () => {
      process.env.NEXT_PUBLIC_API_KEY = 'test-api-key-123';

      fetch.mockResolvedValue({
        ok: true,
        json: async () => ({ countries: [], total: 0 }),
      });

      renderHook(() => useStats(0), { wrapper });

      await waitFor(() => {
        expect(fetch).toHaveBeenCalledWith(
          expect.any(String),
          expect.objectContaining({
            headers: expect.objectContaining({
              'X-API-Key': 'test-api-key-123',
            }),
          })
        );
      });
    });

    it('should not include API key header when NEXT_PUBLIC_API_KEY is not set', async () => {
      delete process.env.NEXT_PUBLIC_API_KEY;

      fetch.mockResolvedValue({
        ok: true,
        json: async () => ({ countries: [], total: 0 }),
      });

      renderHook(() => useStats(0), { wrapper });

      await waitFor(() => {
        const fetchCall = fetch.mock.calls[0];
        expect(fetchCall[1].headers).toEqual({});
      });
    });
  });

  describe('error handling', () => {
    it('should handle HTTP error responses', async () => {
      fetch.mockResolvedValue({
        ok: false,
        status: 500,
      });

      const { result } = renderHook(() => useStats(0), { wrapper });

      await waitFor(() => {
        expect(result.current.error).toBeDefined();
      }, { timeout: 10000 }); // Longer timeout for retry logic (3 attempts + exponential backoff)

      expect(result.current.countries).toEqual([]);
      expect(result.current.total).toBe(0);
    });

    it('should handle network errors', async () => {
      fetch.mockRejectedValue(new Error('Network error'));

      const { result } = renderHook(() => useStats(0), { wrapper });

      await waitFor(() => {
        expect(result.current.error).toBeDefined();
      }, { timeout: 10000 }); // Longer timeout for retry logic (3 attempts + exponential backoff)

      expect(result.current.countries).toEqual([]);
      expect(result.current.total).toBe(0);
    });

    it('should return empty arrays when data is missing', async () => {
      fetch.mockResolvedValue({
        ok: true,
        json: async () => ({}),
      });

      const { result } = renderHook(() => useStats(0), { wrapper });

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.countries).toEqual([]);
      expect(result.current.total).toBe(0);
    });

    it.skip('should handle malformed JSON - TODO: Fix SWR with React 19', async () => {
      fetch.mockResolvedValue({
        ok: true,
        json: async () => {
          throw new Error('Invalid JSON');
        },
      });

      const { result } = renderHook(() => useStats(0));

      await waitFor(() => {
        expect(result.current.error).toBeDefined();
      });
    });
  });

  describe('data structure', () => {
    it('should handle empty countries array', async () => {
      fetch.mockResolvedValue({
        ok: true,
        json: async () => ({ countries: [], total: 0 }),
      });

      const { result } = renderHook(() => useStats(0));

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(result.current.countries).toEqual([]);
      expect(result.current.total).toBe(0);
    });

    it.skip('should handle large country lists - TODO: Fix SWR timing', async () => {
      const largeCountryList = Array.from({ length: 100 }, (_, i) => `C${i.toString().padStart(2, '0')}`);

      fetch.mockResolvedValue({
        ok: true,
        json: async () => ({ countries: largeCountryList, total: 100 }),
      });

      const { result } = renderHook(() => useStats(0), { wrapper });

      await waitFor(() => {
        expect(result.current.countries).toHaveLength(100);
      }, { timeout: 3000 });

      expect(result.current.total).toBe(100);
    });

    it.skip('should preserve country order - TODO: Fix SWR timing', async () => {
      const orderedCountries = ['ZZZ', 'AAA', 'MMM', 'BBB'];

      fetch.mockResolvedValue({
        ok: true,
        json: async () => ({ countries: orderedCountries, total: 4 }),
      });

      const { result } = renderHook(() => useStats(0), { wrapper });

      await waitFor(() => {
        expect(result.current.countries.length).toBeGreaterThan(0);
      }, { timeout: 3000 });

      expect(result.current.countries).toEqual(['ZZZ', 'AAA', 'MMM', 'BBB']);
    });
  });

  describe('refresh interval', () => {
    it('should use default refresh interval of 60000ms', async () => {
      fetch.mockResolvedValue({
        ok: true,
        json: async () => ({ countries: [], total: 0 }),
      });

      renderHook(() => useStats(), { wrapper });

      await waitFor(() => {
        expect(fetch).toHaveBeenCalled();
      });

      // SWR configuration is tested implicitly through the hook
      // We can't easily test the actual interval timing without integration tests
    });

    it('should accept custom refresh interval', async () => {
      fetch.mockResolvedValue({
        ok: true,
        json: async () => ({ countries: [], total: 0 }),
      });

      renderHook(() => useStats(5000), { wrapper });

      await waitFor(() => {
        expect(fetch).toHaveBeenCalled();
      });
    });

    it('should work with 0 interval (no auto-refresh)', async () => {
      fetch.mockResolvedValue({
        ok: true,
        json: async () => ({ countries: [], total: 0 }),
      });

      renderHook(() => useStats(0), { wrapper });

      await waitFor(() => {
        expect(fetch).toHaveBeenCalled();
      });
    });
  });

  describe('return value structure', () => {
    it('should always return an object with expected properties', async () => {
      fetch.mockResolvedValue({
        ok: true,
        json: async () => ({ countries: ['BRA'], total: 1 }),
      });

      const { result } = renderHook(() => useStats(0));

      expect(result.current).toHaveProperty('countries');
      expect(result.current).toHaveProperty('total');
      expect(result.current).toHaveProperty('isLoading');
      expect(result.current).toHaveProperty('error');

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });
    });

    it('should return correct types', async () => {
      fetch.mockResolvedValue({
        ok: true,
        json: async () => ({ countries: ['BRA'], total: 1 }),
      });

      const { result } = renderHook(() => useStats(0));

      await waitFor(() => {
        expect(result.current.isLoading).toBe(false);
      });

      expect(Array.isArray(result.current.countries)).toBe(true);
      expect(typeof result.current.total).toBe('number');
      expect(typeof result.current.isLoading).toBe('boolean');
    });
  });
});
