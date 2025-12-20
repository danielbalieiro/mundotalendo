'use client'

import { useState } from 'react'
import dynamic from 'next/dynamic'
import Image from 'next/image'
import { months } from '@/config/months'

// Dynamically import Map component (client-side only)
const Map = dynamic(() => import('@/components/Map'), {
  ssr: false,
  loading: () => (
    <div className="flex items-center justify-center h-screen bg-gray-100">
      <div className="text-center">
        <div className="animate-spin rounded-full h-16 w-16 border-b-2 border-blue-600 mx-auto mb-4"></div>
        <p className="text-gray-600">Carregando mapa...</p>
      </div>
    </div>
  ),
})

/**
 * Home page component
 * @returns {JSX.Element}
 */
export default function Home() {
  const [isLegendVisible, setIsLegendVisible] = useState(false)

  return (
    <main className="relative h-screen w-screen overflow-hidden">
      {/* Header */}
      <div className="absolute top-0 left-0 right-0 z-10 p-6">
        <Image
          src="/mundotalendo.png"
          alt="Mundo Tá Lendo 2026"
          width={300}
          height={60}
          priority
          className="h-12 md:h-16 w-auto"
        />
      </div>

      {/* Map */}
      <Map />

      {/* Legend with Toggle */}
      <div className="absolute bottom-0 left-0 right-0 z-10">
        <div
          className={`bg-gradient-to-t from-black/70 to-transparent transition-all duration-300 ease-in-out ${
            isLegendVisible ? 'p-4' : 'p-2'
          }`}
        >
          <div className="max-w-6xl mx-auto">
            {/* Toggle Button */}
            <button
              onClick={() => setIsLegendVisible(!isLegendVisible)}
              className="flex items-center gap-2 text-white font-semibold mb-3 text-sm md:text-base hover:text-blue-300 transition-colors cursor-pointer"
              aria-label={isLegendVisible ? 'Ocultar meses' : 'Mostrar meses'}
              aria-expanded={isLegendVisible}
            >
              <span>Meses do Desafio</span>
              <span className="text-xs">
                {isLegendVisible ? '▼' : '▶'}
              </span>
            </button>

            {/* Legend Grid with Animation */}
            <div
              className={`grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-2 transition-all duration-300 ease-in-out overflow-hidden ${
                isLegendVisible
                  ? 'opacity-100 max-h-96'
                  : 'opacity-0 max-h-0'
              }`}
            >
              {months.map((month) => (
                <div
                  key={month.name}
                  className="flex items-center gap-2 bg-white/10 backdrop-blur-sm rounded px-3 py-2"
                >
                  <div
                    className="w-4 h-4 rounded-sm flex-shrink-0"
                    style={{ backgroundColor: month.color }}
                  />
                  <span className="text-white text-xs md:text-sm">
                    {month.name}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </main>
  )
}
