import { useState, useRef, useCallback } from 'react'

/**
 * @typedef {Object} ImageQueueItem
 * @property {string} url - Image URL to load
 * @property {string} name - Unique identifier for MapLibre sprite
 */

/**
 * Batched async image loader for MapLibre sprites
 * Loads images in queue with concurrency limit to avoid overwhelming network
 *
 * @param {import('maplibre-gl').Map | null} map - MapLibre map instance
 * @param {number} concurrency - Max parallel image loads (default: 5)
 * @returns {{
 *   loadImages: (items: ImageQueueItem[]) => void,
 *   loadedCount: number,
 *   totalCount: number,
 *   isLoading: boolean
 * }}
 */
export default function useAsyncImages(map, concurrency = 5) {
  const [loadedCount, setLoadedCount] = useState(0)
  const [totalCount, setTotalCount] = useState(0)
  const [isLoading, setIsLoading] = useState(false)
  const queueRef = useRef([])
  const activeRef = useRef(0)

  const processQueue = useCallback(() => {
    if (!map || queueRef.current.length === 0 || activeRef.current >= concurrency) {
      if (activeRef.current === 0 && queueRef.current.length === 0 && isLoading) {
        setIsLoading(false)
        console.log('Image queue completed:', loadedCount, 'loaded')
      }
      return
    }

    const item = queueRef.current.shift()
    activeRef.current++

    const img = new Image()
    img.crossOrigin = 'anonymous'

    img.onload = () => {
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

        // Replace placeholder with circular avatar
        if (map.hasImage(item.name)) {
          map.removeImage(item.name)
        }
        map.addImage(item.name, imageData, { pixelRatio: 1 })

        setLoadedCount(prev => prev + 1)
        activeRef.current--
        processQueue() // Load next in queue
      } catch (error) {
        console.warn('Failed to process avatar image:', item.name, error)
        setLoadedCount(prev => prev + 1)
        activeRef.current--
        processQueue()
      }
    }

    img.onerror = (err) => {
      console.warn('Failed to load avatar:', item.url, err)
      // Fallback: create gray circle placeholder
      if (!map.hasImage(item.name)) {
        const size = 48
        const canvas = document.createElement('canvas')
        canvas.width = size
        canvas.height = size
        const ctx = canvas.getContext('2d')
        ctx.fillStyle = '#D1D5DB'
        ctx.beginPath()
        ctx.arc(size / 2, size / 2, size / 2 - 2, 0, Math.PI * 2)
        ctx.fill()
        // Convert to ImageData for MapLibre
        const imageData = ctx.getImageData(0, 0, size, size)
        map.addImage(item.name, imageData, { pixelRatio: 1 })
      }
      setLoadedCount(prev => prev + 1)
      activeRef.current--
      processQueue()
    }

    img.src = item.url
  }, [map, concurrency, isLoading, loadedCount])

  const loadImages = useCallback((items) => {
    if (items.length === 0) return

    console.log('Queueing', items.length, 'avatar images for background load')
    queueRef.current = [...items]
    setTotalCount(items.length)
    setLoadedCount(0)
    setIsLoading(true)

    // Start processing queue (will batch based on concurrency)
    for (let i = 0; i < concurrency; i++) {
      processQueue()
    }
  }, [concurrency, processQueue])

  return { loadImages, loadedCount, totalCount, isLoading }
}
