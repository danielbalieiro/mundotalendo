'use client'

import { useState } from 'react'
import { useYears } from '@/hooks/useYears'
import { useStats } from '@/hooks/useStats'
import { useUserLocations } from '@/hooks/useUserLocations'
import { getCountryName } from '@/config/countries'
import { getCountryProgressColor } from '@/utils/colorTiers'
import { months } from '@/config/months'
import { DEFAULT_YEAR } from '@/config/years'

/**
 * Sidebar / statistics panel with a global year filter.
 * @param {Object} props
 * @param {number} props.year - Currently selected year
 * @param {(year: number) => void} props.onYearChange - Callback when the year changes
 */
export default function Sidebar({ year, onYearChange }) {
  const [isOpen, setIsOpen] = useState(false)

  const { years } = useYears()
  const { countries, total: countryTotal, isLoading: statsLoading } = useStats(60000, year)
  const { users } = useUserLocations(year)

  // Always offer the selected year and the current year, even if they have no data yet.
  const options = [...new Set([...(years || []), year, DEFAULT_YEAR])].sort((a, b) => b - a)
  const sortedCountries = [...(countries || [])].sort((a, b) => b.progress - a.progress)

  return (
    <>
      {/* Toggle button (visible when closed) */}
      {!isOpen && (
        <button
          onClick={() => setIsOpen(true)}
          className="fixed top-4 right-4 z-40 bg-white/95 backdrop-blur-sm rounded-lg shadow-lg px-4 py-3 text-gray-700 hover:bg-white transition-colors"
          aria-label="Abrir painel de estatísticas"
        >
          <span className="flex items-center gap-2 text-sm font-semibold">
            ☰ Estatísticas
          </span>
        </button>
      )}

      {/* Sidebar */}
      <aside
        className={`fixed inset-y-0 right-0 z-30 w-80 max-w-[85vw] bg-white shadow-xl flex flex-col
          transform transition-transform duration-300 ease-in-out
          ${isOpen ? 'translate-x-0' : 'translate-x-full'}`}
      >
        {/* Header */}
        <div className="p-4 bg-gradient-to-r from-blue-600 to-purple-600 text-white flex justify-between items-center">
          <h2 className="font-bold text-lg">Estatísticas</h2>
          <button
            onClick={() => setIsOpen(false)}
            className="text-white hover:bg-white/20 rounded-full w-6 h-6 flex items-center justify-center"
            aria-label="Fechar painel de estatísticas"
          >
            ✕
          </button>
        </div>

        {/* Year filter */}
        <div className="p-4 border-b border-gray-200">
          <label
            htmlFor="year-select"
            className="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-2"
          >
            Ano
          </label>
          <select
            id="year-select"
            value={year}
            onChange={(e) => onYearChange(Number(e.target.value))}
            aria-label="Filtrar leituras por ano"
            className="w-full bg-white border border-gray-300 rounded-lg px-3 py-2 text-sm text-gray-800 font-semibold cursor-pointer focus:outline-none focus:ring-2 focus:ring-blue-400"
          >
            {options.map((y) => (
              <option key={y} value={y}>
                {y}
              </option>
            ))}
          </select>
        </div>

        {/* Summary stats */}
        <div className="p-4 grid grid-cols-2 gap-3 border-b border-gray-200">
          <div className="bg-gray-50 rounded-lg p-3">
            <div className="text-2xl font-bold text-gray-800">
              {statsLoading ? '—' : countryTotal}
            </div>
            <div className="text-xs text-gray-500 mt-1">Países explorados</div>
          </div>
          <div className="bg-gray-50 rounded-lg p-3">
            <div className="text-2xl font-bold text-gray-800">{users.length}</div>
            <div className="text-xs text-gray-500 mt-1">Leitores ativos</div>
          </div>
        </div>

        {/* Country list */}
        <div className="flex-1 overflow-y-auto p-4">
          <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-3">
            Países sendo lidos
          </h3>

          {statsLoading ? (
            <div className="space-y-3">
              {[0, 1, 2, 3, 4].map((i) => (
                <div key={i} className="animate-pulse">
                  <div className="h-3 bg-gray-200 rounded w-3/4 mb-2" />
                  <div className="h-1.5 bg-gray-200 rounded" />
                </div>
              ))}
            </div>
          ) : sortedCountries.length === 0 ? (
            <p className="text-sm text-gray-500">Nenhuma leitura registrada para este ano.</p>
          ) : (
            <ul className="space-y-3">
              {sortedCountries.map((c) => (
                <li key={c.iso3}>
                  <div className="flex items-center justify-between text-sm mb-1">
                    <span className="text-gray-800 font-medium">
                      {getCountryName(c.iso3) || c.iso3}
                    </span>
                    <span className="text-gray-500">{c.progress}%</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-1.5">
                    <div
                      className="h-1.5 rounded-full"
                      style={{
                        width: `${c.progress}%`,
                        backgroundColor: getCountryProgressColor(c.iso3, c.progress, months),
                      }}
                    />
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </aside>
    </>
  )
}
