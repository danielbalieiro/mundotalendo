'use client'

import { useEffect, useRef, useState, useCallback } from 'react'
import maplibregl from 'maplibre-gl'
import { useStats } from '@/hooks/useStats'
import { useUserLocations } from '@/hooks/useUserLocations'
import useCountryReadings from '@/hooks/useCountryReadings'
import { getMonthByCountry, months } from '@/config/months'
import { getCountryName, countryNames } from '@/config/countries'
import { countryCentroids } from '@/config/countryCentroids'
import { getCountryProgressColor, getTierLabel } from '@/utils/colorTiers'
import { logger } from '@/utils/logger'
import CountryPopup from './CountryPopup'

// Feature flag for user markers
const SHOW_USER_MARKERS = process.env.NEXT_PUBLIC_SHOW_USER_MARKERS === 'true'

// Configuration for concentric rings around country centroids
const RING_BASE_RADIUS = 1.2       // degrees - first ring radius
const RING_INCREMENT = 0.9          // degrees - increment between rings
const MIN_SPACING_DEGREES = 0.35    // degrees - minimum spacing between user avatars

/**
 * Distribute users into concentric rings around a country centroid
 * @param {Array} users - Array of user objects to distribute
 * @returns {Array} Array of ring objects {radius, users, count}
 */
function distributeUsersInRings(users) {
  const rings = []
  let remainingUsers = users.length
  let ringIndex = 0
  let userOffset = 0

  while (remainingUsers > 0) {
    const radius = RING_BASE_RADIUS + (ringIndex * RING_INCREMENT)
    const circumference = 2 * Math.PI * radius
    const ringCapacity = Math.floor(circumference / MIN_SPACING_DEGREES)
    const usersInRing = Math.min(remainingUsers, ringCapacity)

    rings.push({
      radius: radius,
      users: users.slice(userOffset, userOffset + usersInRing),
      count: usersInRing
    })

    userOffset += usersInRing
    remainingUsers -= usersInRing
    ringIndex++
  }

  return rings
}

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
 * Build GeoJSON for user markers with offset for multiple users in same country
 * @param {Array} users - Array of user location objects
 * @param {Object} centroids - Country centroids mapping (ISO -> [lng, lat])
 * @returns {Object} GeoJSON FeatureCollection
 */
export function buildUserMarkersGeoJSON(users, centroids) {
  // Group users by country
  const usersByCountry = {}
  users.forEach(user => {
    if (!user.avatarURL || !centroids[user.iso3]) return // Filter out users without avatar or invalid country
    if (!usersByCountry[user.iso3]) {
      usersByCountry[user.iso3] = []
    }
    usersByCountry[user.iso3].push(user)
  })

  // Create features with concentric ring positioning
  const features = []
  Object.entries(usersByCountry).forEach(([iso, countryUsers]) => {
    const baseCoords = centroids[iso]

    // Distribute users into concentric rings
    const rings = distributeUsersInRings(countryUsers)

    // Log ring distribution in development
    if (process.env.NODE_ENV === 'development' && rings.length > 1) {
      logger.debug(`${iso}: ${countryUsers.length} users in ${rings.length} rings`)
    }

    // Process each ring
    rings.forEach((ring, ringIndex) => {
      const angleStep = (2 * Math.PI) / ring.count

      ring.users.forEach((user, indexInRing) => {
        // Calculate angle for this user in the ring (360° distribution)
        const angle = indexInRing * angleStep

        // Convert polar coordinates to cartesian (lng/lat offsets)
        const offsetLng = ring.radius * Math.cos(angle)
        const offsetLat = ring.radius * Math.sin(angle)

        features.push({
          type: 'Feature',
          geometry: {
            type: 'Point',
            coordinates: [baseCoords[0] + offsetLng, baseCoords[1] + offsetLat],
          },
          properties: {
            user: user.user,
            avatarURL: user.avatarURL,
            country: user.pais,
            book: user.livro || user.pais, // Fallback to country if no book
            timestamp: user.timestamp,
          },
        })
      })
    })
  })

  return {
    type: 'FeatureCollection',
    features,
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
  const [hoveredUser, setHoveredUser] = useState(null)
  const [cursorPosition, setCursorPosition] = useState({ x: 0, y: 0 })
  const [popup, setPopup] = useState(null) // { iso3, countryName, position: {x, y}, readers }

  const { countries, total, isLoading, error } = useStats()
  const { users } = useUserLocations()
  const { fetchReadings, readings, loading: readingsLoading, error: readingsError } = useCountryReadings()

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

  // Handle country click to show popup with readers
  const handleCountryClick = useCallback(async (e) => {
    if (!map.current) return

    const features = map.current.queryRenderedFeatures(e.point, {
      layers: ['country-fills']
    })

    if (features.length > 0) {
      const iso3 = features[0].properties.ADM0_A3
      const countryName = getCountryName(iso3) || features[0].properties.name || iso3

      // Check if country is colored (has stats data with progress >= 1%)
      const countryStats = countries.find(c => c.iso3 === iso3)

      // Only show popup if country is being read (progress >= 1%)
      if (countryStats && countryStats.progress >= 1) {
        // Set popup immediately with loading state
        setPopup({
          iso3,
          countryName,
          position: { x: e.point.x, y: e.point.y },
          readers: [],
          loading: true,
          error: null
        })

        // Fetch readings data asynchronously
        await fetchReadings(iso3)
      }
    }
  }, [countries, fetchReadings])

  // Close popup
  const handleClosePopup = useCallback(() => {
    setPopup(null)
  }, [])

  // Update popup when readings data arrives
  useEffect(() => {
    if (!readingsLoading) {
      setPopup(prev => {
        if (!prev || prev.loading === false) return prev
        return {
          ...prev,
          readers: readings,
          loading: false,
          error: readingsError
        }
      })
    }
  }, [readings, readingsLoading, readingsError])

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
      minZoom: 2.0, // Limit zoom out to prevent excessive distance
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

      // Add user markers (only if feature flag is enabled)
      if (SHOW_USER_MARKERS && !map.current.getSource('user-markers')) {
        map.current.addSource('user-markers', {
          type: 'geojson',
          data: buildUserMarkersGeoJSON(users, countryCentroids),
        })

        // Add layer with white circle background
        map.current.addLayer({
          id: 'user-marker-bg',
          type: 'circle',
          source: 'user-markers',
          paint: {
            'circle-radius': 12,
            'circle-color': '#ffffff',
            'circle-stroke-width': 2,
            'circle-stroke-color': '#2563eb', // GPS blue
          },
        })

        // Add layer with user avatar images
        map.current.addLayer({
          id: 'user-marker-images',
          type: 'symbol',
          source: 'user-markers',
          layout: {
            'icon-image': ['get', 'user'], // Use username as sprite ID
            'icon-size': 0.5, // 48px sprite * 0.5 = 24px (matches circle diameter)
            'icon-allow-overlap': true,
            'icon-ignore-placement': true,
          },
        })

        logger.debug('User markers layers added!')
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

    // User marker mouse handlers (only if feature flag is enabled)
    if (SHOW_USER_MARKERS) {
      map.current.on('mouseenter', 'user-marker-images', (e) => {
        if (e.features && e.features.length > 0) {
          const feature = e.features[0]
          setHoveredUser({
            user: feature.properties.user,
            country: feature.properties.country,
            book: feature.properties.book,
            avatarURL: feature.properties.avatarURL,
            timestamp: feature.properties.timestamp,
          })
          setCursorPosition({ x: e.point.x, y: e.point.y })
          map.current.getCanvas().style.cursor = 'pointer'
        }
      })

      map.current.on('mouseleave', 'user-marker-images', () => {
        setHoveredUser(null)
        map.current.getCanvas().style.cursor = ''
      })
    }

    // Click handler for countries to show popup with readers
    map.current.on('click', 'country-fills', handleCountryClick)

    return () => {
      if (map.current) {
        map.current.off('click', 'country-fills', handleCountryClick)
        map.current.remove()
        map.current = null
      }
    }
  }, [handleCountryClick])

  // Update colors when countries change
  useEffect(() => {
    applyCountryColors()
  }, [countries, applyCountryColors])

  // Load user avatar sprites (only if feature flag is enabled)
  useEffect(() => {
    if (!SHOW_USER_MARKERS || !map.current || !map.current.hasImage) return

    users.forEach(user => {
      if (!user.avatarURL) return

      // Check if sprite already loaded
      if (map.current.hasImage(user.user)) return

      // Load avatar image via proxy to bypass CORS in development
      const img = new Image()
      img.crossOrigin = 'anonymous'
      img.onload = () => {
        if (map.current && !map.current.hasImage(user.user)) {
          try {
            // Create circular clipped version of the image
            const size = 48 // Size of the circular sprite
            const canvas = document.createElement('canvas')
            canvas.width = size
            canvas.height = size
            const ctx = canvas.getContext('2d')

            // Draw circular clip path
            ctx.beginPath()
            ctx.arc(size / 2, size / 2, size / 2, 0, Math.PI * 2)
            ctx.closePath()
            ctx.clip()

            // Draw image to fill the circle (cover the entire area)
            const imgAspect = img.width / img.height
            let drawWidth, drawHeight, offsetX, offsetY

            if (imgAspect > 1) {
              // Image is wider - fit height and crop sides
              drawHeight = size
              drawWidth = size * imgAspect
              offsetX = -(drawWidth - size) / 2
              offsetY = 0
            } else {
              // Image is taller - fit width and crop top/bottom
              drawWidth = size
              drawHeight = size / imgAspect
              offsetX = 0
              offsetY = -(drawHeight - size) / 2
            }

            ctx.drawImage(img, offsetX, offsetY, drawWidth, drawHeight)

            // Convert canvas to ImageData for MapLibre
            const imageData = ctx.getImageData(0, 0, size, size)

            // Add the ImageData as sprite (pixelRatio: 1 so 48px sprite = 48px on map)
            map.current.addImage(user.user, imageData, { pixelRatio: 1 })
            logger.debug(`Loaded circular avatar sprite for ${user.user}`)
          } catch (error) {
            logger.warn(`Failed to add sprite for ${user.user}:`, error)
          }
        }
      }
      img.onerror = () => {
        logger.warn(`Failed to load avatar for ${user.user}: ${user.avatarURL}`)
      }
      // Always use proxy to bypass CORS for external images
      const imageUrl = `/api/proxy-image?url=${encodeURIComponent(user.avatarURL)}`
      img.src = imageUrl
    })
  }, [users])

  // Update user markers GeoJSON when users change
  useEffect(() => {
    if (!SHOW_USER_MARKERS) return
    if (map.current && map.current.getSource('user-markers')) {
      map.current.getSource('user-markers').setData(
        buildUserMarkersGeoJSON(users, countryCentroids)
      )
      logger.debug(`Updated user markers: ${users.length} users`)
    }
  }, [users])

  return (
    <div className="relative w-full h-full">
      <div ref={mapContainer} className="w-full h-full" />

      {/* Hover Tooltip - Country */}
      {hoveredCountry && !hoveredUser && (
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

      {/* Hover Tooltip - User Marker */}
      {SHOW_USER_MARKERS && hoveredUser && (
        <div
          className="absolute z-30 bg-blue-600/95 text-white px-4 py-2 rounded-lg shadow-lg pointer-events-none"
          style={{
            left: cursorPosition.x + 15,
            top: cursorPosition.y + 15,
          }}
        >
          <div className="font-semibold text-sm">📍 {hoveredUser.user}</div>
          <div className="text-xs opacity-90 mt-1">
            Lendo: {hoveredUser.book}
          </div>
        </div>
      )}

      {/* Country Popup - Readers List */}
      {popup && (
        <CountryPopup
          readers={popup.readers}
          countryName={popup.countryName}
          position={popup.position}
          loading={popup.loading || false}
          error={popup.error || null}
          onClose={handleClosePopup}
        />
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
