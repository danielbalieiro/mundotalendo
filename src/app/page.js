'use client'

import dynamic from 'next/dynamic'
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
  return (
    <main className="relative h-screen w-screen overflow-hidden">
      {/* Header */}
      <div className="absolute top-0 left-0 right-0 z-10 bg-gradient-to-b from-black/70 to-transparent p-6">
        <h1 className="text-3xl md:text-4xl font-bold text-white">
          Mundo Tá Lendo 2026
        </h1>
      </div>

      {/* Map */}
      <Map />

      {/* Legend */}
      <div className="absolute bottom-0 left-0 right-0 z-10 bg-gradient-to-t from-black/70 to-transparent p-4">
        <div className="max-w-6xl mx-auto">
          <h2 className="text-white font-semibold mb-3 text-sm md:text-base">
            Meses do Desafio
          </h2>
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-2">
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
    </main>
  )
}
