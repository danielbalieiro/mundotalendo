'use client'

import { useState } from 'react'
import useCountryReadings from '@/hooks/useCountryReadings'
import { getCountryName } from '@/config/countries'
import { getCountryProgressColor } from '@/utils/colorTiers'
import { months } from '@/config/months'

/**
 * Collapsible country row. Fetches the readings for this country on first
 * expand (lazy), so the sidebar stays light even with many countries.
 * @param {Object} props
 * @param {string} props.iso3 - Country ISO3 code
 * @param {number} props.progress - Aggregated progress percentage
 * @param {number|string} props.year - Year being displayed
 */
export default function CountryItem({ iso3, progress, year }) {
  const [isOpen, setIsOpen] = useState(false)
  const [hasLoaded, setHasLoaded] = useState(false)
  const { fetchReadings, readings, loading, error } = useCountryReadings()

  const handleToggle = () => {
    const next = !isOpen
    setIsOpen(next)
    if (next && !hasLoaded) {
      setHasLoaded(true)
      fetchReadings(iso3, year)
    }
  }

  return (
    <li>
      <button
        onClick={handleToggle}
        className="w-full text-left cursor-pointer group"
        aria-expanded={isOpen}
      >
        <div className="flex items-center justify-between text-sm mb-1">
          <span className="text-gray-800 font-medium flex items-center gap-1.5 group-hover:text-blue-600 transition-colors">
            <span className="text-[10px] text-gray-400">{isOpen ? '▼' : '▶'}</span>
            {getCountryName(iso3) || iso3}
          </span>
          <span className="text-gray-500">{progress}%</span>
        </div>
      </button>

      <div className="w-full bg-gray-200 rounded-full h-1.5">
        <div
          className="h-1.5 rounded-full"
          style={{
            width: `${progress}%`,
            backgroundColor: getCountryProgressColor(iso3, progress, months),
          }}
        />
      </div>

      {isOpen && (
        <div className="mt-2 pl-5">
          {loading ? (
            <div className="space-y-1.5">
              {[0, 1, 2].map((i) => (
                <div key={i} className="animate-pulse h-3 bg-gray-200 rounded w-3/4" />
              ))}
            </div>
          ) : error ? (
            <p className="text-xs text-red-600">Erro ao carregar leituras.</p>
          ) : readings.length === 0 ? (
            <p className="text-xs text-gray-500">Nenhuma leitura encontrada.</p>
          ) : (
            <ul className="space-y-1">
              {readings.map((r) => (
                <li
                  key={`${r.user}-${r.livro}-${r.updatedAt}`}
                  className="flex items-center gap-2 text-xs text-gray-600"
                >
                  <span className="shrink-0">📖</span>
                  <span className="truncate">{r.livro || 'Sem título'}</span>
                  <span className="text-gray-400 ml-auto shrink-0">{r.user}</span>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </li>
  )
}
