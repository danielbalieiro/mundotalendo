'use client'

import { months } from '@/config/months'
import { getColorTier, getTierLabel } from '@/utils/colorTiers'

/**
 * Test page to visually validate all 60 color combinations
 * (12 months × 5 tiers)
 *
 * Access at: http://localhost:3000/test-colors
 */
export default function TestColorsPage() {
  // Sample progress values representing each tier
  const progressSamples = [
    { value: 10, label: '10% (Tier 1)' },
    { value: 30, label: '30% (Tier 2)' },
    { value: 50, label: '50% (Tier 3)' },
    { value: 70, label: '70% (Tier 4)' },
    { value: 90, label: '90% (Tier 5)' }
  ]

  return (
    <div className="p-8 bg-gray-50 min-h-screen">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-gray-800 mb-2">
            🎨 Teste de Cores - Sistema de 5 Tiers
          </h1>
          <p className="text-gray-600">
            Validação visual das 60 combinações de cores (12 meses × 5 níveis de progresso)
          </p>
        </div>

        {/* Legend */}
        <div className="bg-white rounded-lg shadow-sm p-6 mb-8">
          <h2 className="text-lg font-semibold text-gray-800 mb-4">Legenda</h2>
          <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
            <div className="text-sm">
              <span className="font-semibold">Tier 1:</span> 0-20% (Tom mais claro)
            </div>
            <div className="text-sm">
              <span className="font-semibold">Tier 2:</span> 21-40% (Tom claro)
            </div>
            <div className="text-sm">
              <span className="font-semibold">Tier 3:</span> 41-60% (Tom médio)
            </div>
            <div className="text-sm">
              <span className="font-semibold">Tier 4:</span> 61-80% (Tom escuro)
            </div>
            <div className="text-sm">
              <span className="font-semibold">Tier 5:</span> 81-100% (Cor vibrante total)
            </div>
          </div>
        </div>

        {/* Month sections */}
        {months.map((month) => (
          <div key={month.name} className="mb-12">
            <div className="bg-white rounded-lg shadow-md p-6">
              {/* Month header */}
              <h2 className="text-2xl font-bold text-gray-800 mb-6 flex items-center gap-3">
                <span>{month.name}</span>
                <span className="text-sm font-normal text-gray-500">
                  ({month.countries.length} países)
                </span>
              </h2>

              {/* Color grid */}
              <div className="grid grid-cols-2 md:grid-cols-5 gap-6">
                {progressSamples.map((sample, index) => {
                  const tier = getColorTier(sample.value)
                  const color = month.colors[`tier${tier}`]
                  const tierLabel = getTierLabel(sample.value)

                  return (
                    <div key={index} className="text-center">
                      {/* Color swatch */}
                      <div
                        className="w-full h-32 rounded-lg shadow-md mb-3 border border-gray-200"
                        style={{ backgroundColor: color }}
                      />

                      {/* Progress info */}
                      <div className="space-y-1">
                        <p className="font-mono text-sm font-semibold text-gray-700">
                          {sample.label}
                        </p>
                        <p className="text-xs text-gray-500">
                          {tierLabel}
                        </p>
                        <p className="text-xs font-mono text-gray-400">
                          {color}
                        </p>
                      </div>
                    </div>
                  )
                })}
              </div>

              {/* Example countries */}
              <div className="mt-6 pt-6 border-t border-gray-200">
                <p className="text-sm text-gray-600">
                  <span className="font-semibold">Países exemplo:</span>{' '}
                  {month.countries.slice(0, 5).join(', ')}
                  {month.countries.length > 5 && ` (+${month.countries.length - 5} mais)`}
                </p>
              </div>
            </div>
          </div>
        ))}

        {/* Boundary tests */}
        <div className="bg-white rounded-lg shadow-md p-6 mb-8">
          <h2 className="text-2xl font-bold text-gray-800 mb-6">
            🧪 Teste de Limites (Boundaries)
          </h2>
          <p className="text-sm text-gray-600 mb-6">
            Validar que valores exatos nos limites pertencem à faixa correta
          </p>

          <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
            {[0, 20, 21, 40, 41, 60, 61, 80, 81, 100].map((progress) => {
              const tier = getColorTier(progress)
              const color = months[0].colors[`tier${tier}`] // Use Janeiro as example

              return (
                <div key={progress} className="text-center">
                  <div
                    className="w-full h-24 rounded shadow-sm border border-gray-300"
                    style={{ backgroundColor: color }}
                  />
                  <p className="mt-2 font-mono text-sm font-semibold">
                    {progress}% → Tier {tier}
                  </p>
                </div>
              )
            })}
          </div>
        </div>

        {/* Footer */}
        <div className="bg-blue-50 rounded-lg p-6 text-center">
          <p className="text-sm text-gray-700">
            ✅ Se todas as cores aparecem corretamente e os gradientes são visualmente distintos,
            o sistema de 5 tiers está funcionando perfeitamente!
          </p>
          <p className="text-xs text-gray-500 mt-2">
            Teste também em modo daltônico no DevTools para garantir acessibilidade
          </p>
        </div>
      </div>
    </div>
  )
}
