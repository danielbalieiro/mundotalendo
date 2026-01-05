'use client'

import { useEffect, useRef } from 'react'

/**
 * Popup que exibe leitores de um país
 * @param {Object} props
 * @param {Array} props.readers - Array de { user, avatarURL, capaURL, livro }
 * @param {string} props.countryName - Nome do país
 * @param {{x: number, y: number}} props.position - Posição pixel no mapa
 * @param {Function} props.onClose - Callback para fechar popup
 */
export default function CountryPopup({ readers, countryName, position, onClose }) {
  const popupRef = useRef(null)

  // Fechar ao clicar fora
  useEffect(() => {
    function handleClickOutside(event) {
      if (popupRef.current && !popupRef.current.contains(event.target)) {
        onClose()
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [onClose])

  // Fechar com ESC
  useEffect(() => {
    function handleEscape(event) {
      if (event.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', handleEscape)
    return () => document.removeEventListener('keydown', handleEscape)
  }, [onClose])

  if (!readers || readers.length === 0) return null

  return (
    <div
      ref={popupRef}
      className="absolute bg-gradient-to-br from-blue-50 via-indigo-50 to-purple-50 rounded-lg shadow-2xl border border-blue-200/50 z-50 max-w-sm overflow-hidden"
      style={{
        left: `${position.x}px`,
        top: `${position.y}px`,
        transform: 'translate(-50%, -100%)', // Centraliza horizontalmente, aparece acima do click
        marginTop: '-10px', // Offset para não cobrir o marker
      }}
    >
      {/* Header */}
      <div className="flex items-start justify-between px-4 py-3 bg-gradient-to-r from-blue-500 to-purple-600">
        <h3 className="font-bold text-lg text-white flex-1 break-words pr-2">{countryName}</h3>
        <button
          onClick={onClose}
          className="text-white/80 hover:text-white transition-colors flex-shrink-0 mt-0.5"
          aria-label="Fechar"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      {/* Lista de leitores (scrollável) */}
      <div className="max-h-96 overflow-y-auto">
        {readers.map((reader, idx) => (
          <div
            key={idx}
            className="flex items-start gap-3 px-4 py-3 hover:bg-white/60 transition-colors"
          >
            {/* Capa do livro (esquerda) */}
            {reader.capaURL ? (
              <img
                src={reader.capaURL}
                alt={`Capa de ${reader.livro}`}
                className="w-12 h-16 object-cover rounded shadow-sm flex-shrink-0"
                onError={(e) => {
                  // Fallback se imagem falhar
                  e.target.style.display = 'none'
                }}
              />
            ) : (
              <div className="w-12 h-16 bg-gray-200 rounded flex items-center justify-center flex-shrink-0">
                <svg className="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                </svg>
              </div>
            )}

            {/* Informações do leitor (direita) */}
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 mb-1">
                {/* Avatar circular */}
                <img
                  src={reader.avatarURL || '/api/proxy-image?url=' + encodeURIComponent('https://i.pravatar.cc/150?img=1')}
                  alt={reader.user}
                  className="w-6 h-6 rounded-full flex-shrink-0"
                />
                <p className="font-semibold text-sm text-gray-900 truncate">
                  {reader.user}
                </p>
              </div>
              <p className="text-sm text-gray-600 line-clamp-2">
                {reader.livro}
              </p>
            </div>
          </div>
        ))}
      </div>

      {/* Footer com contagem */}
      <div className="px-4 py-2 bg-gradient-to-r from-blue-100/80 to-purple-100/80 border-t border-blue-200/50">
        <p className="text-xs text-indigo-700 font-medium text-center">
          {readers.length} {readers.length === 1 ? 'leitor' : 'leitores'}
        </p>
      </div>
    </div>
  )
}
