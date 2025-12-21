'use client'

import { useEffect, useRef, useState, useCallback } from 'react'
import maplibregl from 'maplibre-gl'
import { useStats } from '@/hooks/useStats'
import { getMonthByCountry, months } from '@/config/months'
import { getCountryName, countryNames } from '@/config/countries'
import { countryCentroids } from '@/config/countryCentroids'
import { getCountryProgressColor, getTierLabel } from '@/utils/colorTiers'
import { logger } from '@/utils/logger'

/**
 * Build GeoJSON FeatureCollection with country centroids and Portuguese names
 * @returns {Object} GeoJSON FeatureCollection
 */
export function buildCountryLabelsGeoJSON() {
  const features = Object.entries(countryCentroids).map(([iso, coordinates]) => ({
    type: 'Feature',
    geometry: {
      type: 'Point',
      coordinates: coordinates,
    },
    properties: {
      iso: iso,
      name: countryNames[iso] || iso,
    },
  }))

  return {
    type: 'FeatureCollection',
    features: features,
  }
}

/**
 * Map component with MapLibre GL JS
 * @returns {JSX.Element}
 */
export default function Map() {
  const mapContainer = useRef(null)
  const map = useRef(null)
  const [hoveredCountry, setHoveredCountry] = useState(null)
  const [cursorPosition, setCursorPosition] = useState({ x: 0, y: 0 })

  const { countries, total, isLoading, error } = useStats()

  // Function to apply country colors to the map (memoized with useCallback)
  const applyCountryColors = useCallback(() => {
    if (!map.current || !map.current.getLayer('country-fills')) {
      return
    }

    // If no countries, just use a solid color
    if (countries.length === 0) {
      map.current.setPaintProperty('country-fills', 'fill-color', '#F5F5F5')
      map.current.setPaintProperty('country-fills', 'fill-opacity', 0.9)
      return
    }

    const colorExpression = ['match', ['get', 'ADM0_A3']]

    // Add tier-based colors for countries being explored
    countries.forEach((countryData) => {
      const iso = countryData.iso3
      const progress = countryData.progress

      // Get tier-based color instead of applying opacity
      const tierColor = getCountryProgressColor(iso, progress, months)
      colorExpression.push(iso, tierColor)
    })

    // Default color for non-explored countries
    colorExpression.push('#F5F5F5')

    // Apply solid colors (no opacity variation)
    map.current.setPaintProperty('country-fills', 'fill-color', colorExpression)
    map.current.setPaintProperty('country-fills', 'fill-opacity', 0.9) // Solid opacity for all
  }, [countries]) // Re-create function when countries changes

  // Initialize map
  useEffect(() => {
    if (map.current) return // Initialize map only once

    map.current = new maplibregl.Map({
      container: mapContainer.current,
      style: {
        version: 8,
        sources: {},
        layers: [
          {
            id: 'background',
            type: 'background',
            paint: {
              'background-color': '#6BB6FF', // Lighter ocean blue
            },
          },
        ],
      },
      center: [-40, 10],
      zoom: 1.5,
      minZoom: 1,
      maxZoom: 6, // Limit maximum zoom to avoid state divisions
    })

    map.current.on('load', () => {
      // Hide all existing country and place labels from base map
      try {
        const layers = map.current.getStyle().layers
        layers.forEach((layer) => {
          if (layer.type === 'symbol' &&
              layer.id &&
              (layer.id.includes('country') || layer.id.includes('place'))) {
            map.current.setLayoutProperty(layer.id, 'visibility', 'none')
          }
        })
      } catch (error) {
        // Silently ignore errors
      }

      // Add countries source for fill and borders
      if (!map.current.getSource('countries')) {
        map.current.addSource('countries', {
          type: 'vector',
          url: 'https://demotiles.maplibre.org/tiles/tiles.json',
        })

        // Add fill layer with initial white color
        map.current.addLayer({
          id: 'country-fills',
          type: 'fill',
          source: 'countries',
          'source-layer': 'countries',
          paint: {
            'fill-color': '#FFFFFF',
            'fill-opacity': 0.9,
          },
        })

        // Apply colors immediately if we already have countries data
        setTimeout(() => {
          applyCountryColors()
        }, 100)

        // Add border layer
        map.current.addLayer({
          id: 'country-borders',
          type: 'line',
          source: 'countries',
          'source-layer': 'countries',
          paint: {
            'line-color': '#334155',
            'line-width': 0.5,
            'line-opacity': 0.3,
          },
        })
      }

      // Add GeoJSON source for Portuguese country labels (one point per country)
      if (!map.current.getSource('country-labels-source')) {
        map.current.addSource('country-labels-source', {
          type: 'geojson',
          data: buildCountryLabelsGeoJSON(),
        })

        // Add label layer with clean, simple configuration
        map.current.addLayer({
          id: 'country-labels-pt',
          type: 'symbol',
          source: 'country-labels-source',
          minzoom: 2,
          maxzoom: 6,
          layout: {
            'text-field': ['get', 'name'],
            'text-font': ['Open Sans Regular'],
            'text-size': [
              'interpolate',
              ['linear'],
              ['zoom'],
              2, 10,
              4, 13,
              6, 16
            ],
            'text-allow-overlap': false,
            'text-ignore-placement': false,
          },
          paint: {
            'text-color': '#1f2937',
            'text-halo-color': '#ffffff',
            'text-halo-width': 2,
            'text-halo-blur': 1,
          },
        })
        logger.debug('Portuguese country labels added from GeoJSON!')
      }
    })

    // Mouse move handler
    map.current.on('mousemove', 'country-fills', (e) => {
      if (e.features && e.features.length > 0) {
        const feature = e.features[0]
        const iso = feature.properties.iso_a3

        const countryData = countries.find(c => c.iso3 === iso)
        if (countryData) {
          setHoveredCountry({
            name: getCountryName(iso),
            iso: iso,
            progress: countryData.progress,
          })
          setCursorPosition({ x: e.point.x, y: e.point.y })
          map.current.getCanvas().style.cursor = 'pointer'
        } else {
          setHoveredCountry(null)
          map.current.getCanvas().style.cursor = ''
        }
      }
    })

    // Mouse leave handler
    map.current.on('mouseleave', 'country-fills', () => {
      setHoveredCountry(null)
      map.current.getCanvas().style.cursor = ''
    })

    return () => {
      if (map.current) {
        map.current.remove()
        map.current = null
      }
    }
  }, [])

  // Update colors when countries change
  useEffect(() => {
    applyCountryColors()
  }, [countries])

  return (
    <div className="relative w-full h-full">
      <div ref={mapContainer} className="w-full h-full" />

      {/* Hover Tooltip */}
      {hoveredCountry && (
        <div
          className="absolute z-20 bg-black/90 text-white px-4 py-2 rounded-lg shadow-lg pointer-events-none"
          style={{
            left: cursorPosition.x + 15,
            top: cursorPosition.y + 15,
          }}
        >
          <div className="font-semibold">{hoveredCountry.name}</div>
          <div className="text-sm text-gray-300">
            {getMonthByCountry(hoveredCountry.iso)?.name || 'Sem categoria'}
          </div>
          <div className="text-sm text-blue-300">
            <span className="font-mono">{hoveredCountry.progress}%</span>
            <span className="text-xs ml-2 opacity-75">
              • {getTierLabel(hoveredCountry.progress)}
            </span>
          </div>
        </div>
      )}

      {/* Error Message */}
      {error && (
        <div className="absolute top-24 left-1/2 transform -translate-x-1/2 z-10 bg-red-500 text-white px-6 py-3 rounded-lg shadow-lg">
          Erro ao carregar dados. Tentando novamente...
        </div>
      )}
    </div>
  )
}
