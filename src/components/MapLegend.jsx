'use client'

import { useState } from 'react'
import { months } from '@/config/months'

/**
 * Expandable map legend showing 12 months and their colors
 * Toggle button to show/hide the full month list
 */
export default function MapLegend() {
  const [isExpanded, setIsExpanded] = useState(false)

  return (
    <div className="absolute bottom-6 left-6 z-10">
      {/* Toggle Button */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="bg-white/95 backdrop-blur-sm rounded-lg shadow-lg px-4 py-3
                   hover:bg-white transition-colors cursor-pointer"
        aria-label={isExpanded ? 'Ocultar legenda de meses' : 'Mostrar legenda de meses'}
        aria-expanded={isExpanded}
      >
        <div className="flex items-center gap-2">
          <span className="text-xs font-semibold text-gray-700">
            🗓️ Meses
          </span>
          <span className="text-xs text-gray-400">
            {isExpanded ? '▼' : '▶'}
          </span>
        </div>
      </button>

      {/* Expanded Content */}
      {isExpanded && (
        <div className="mt-2 bg-white/95 backdrop-blur-sm rounded-lg shadow-lg p-4 max-w-xs">
          <div className="space-y-2">
            {months.map((month) => (
              <div key={month.name} className="flex items-center gap-3">
                {/* Color swatch (tier5 - vibrant color) */}
                <div
                  className="w-6 h-6 rounded shadow-sm"
                  style={{ backgroundColor: month.colors.tier5 }}
                  title={`Cor de ${month.name}`}
                />

                {/* Month name */}
                <span className="text-sm text-gray-700">
                  {month.name}
                </span>

                {/* Country count */}
                <span className="text-xs text-gray-400 ml-auto">
                  {month.countries.length} {month.countries.length === 1 ? 'país' : 'países'}
                </span>
              </div>
            ))}
          </div>

          {/* Footer note */}
          <div className="mt-3 pt-3 border-t border-gray-200">
            <p className="text-xs text-gray-400 italic">
              Cada mês tem uma cor específica
            </p>
          </div>
        </div>
      )}
    </div>
  )
}
