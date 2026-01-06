'use client'

import { useEffect, useRef } from 'react'

/**
 * Popup que exibe todas as leituras de um país
 * @param {Object} props
 * @param {Array} props.readers - Array de { user, avatarURL, capaURL, livro, progresso, categoria, updatedAt }
 * @param {string} props.countryName - Nome do país em português
 * @param {{x: number, y: number}} props.position - Posição pixel no mapa
 * @param {boolean} props.loading - Estado de carregamento
 * @param {string|null} props.error - Mensagem de erro se houver
 * @param {Function} props.onClose - Callback para fechar popup
 */
export default function CountryPopup({ readers = [], countryName, position, loading = false, error = null, onClose }) {
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

  return (
    <div
      ref={popupRef}
      className="absolute bg-white rounded-lg shadow-2xl z-50 w-96 max-h-[500px] flex flex-col"
      style={{
        left: `${position.x + 10}px`,
        top: `${position.y + 10}px`,
      }}
    >
      {/* Header */}
      <div className="bg-gradient-to-r from-blue-600 to-purple-600 text-white px-4 py-3 rounded-t-lg flex justify-between items-center">
        <h3 className="font-bold text-lg leading-tight break-words pr-2">
          {countryName}
        </h3>
        <button
          onClick={onClose}
          className="text-white hover:bg-white/20 rounded-full w-6 h-6 flex items-center justify-center flex-shrink-0"
          aria-label="Fechar"
        >
          ✕
        </button>
      </div>

      {/* Loading State */}
      {loading && (
        <div className="p-8 flex flex-col items-center justify-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mb-4"></div>
          <p className="text-gray-600">Carregando leituras...</p>
        </div>
      )}

      {/* Error State */}
      {error && !loading && (
        <div className="p-6 text-center">
          <p className="text-red-600 mb-2">Erro ao carregar leituras</p>
          <p className="text-sm text-gray-500">{error}</p>
        </div>
      )}

      {/* Content: List of Readings */}
      {!loading && !error && readers.length > 0 && (
        <>
          <div className="overflow-y-auto max-h-96 px-2 py-2">
            <ul className="space-y-3">
              {readers.map((reader, index) => (
                <li
                  key={`${reader.user}-${reader.livro}-${index}`}
                  className="flex items-start gap-3 p-2 hover:bg-gray-50 rounded-lg transition-colors"
                >
                  {/* Avatar */}
                  <img
                    src={reader.avatarURL || '/api/proxy-image?url=' + encodeURIComponent('https://i.pravatar.cc/150?img=1')}
                    alt={reader.user}
                    className="w-12 h-12 rounded-full object-cover flex-shrink-0"
                    onError={(e) => {
                      e.target.src = '/api/proxy-image?url=' + encodeURIComponent('https://i.pravatar.cc/150?img=1')
                    }}
                  />

                  {/* Book Cover */}
                  <div className="flex-shrink-0 w-16 h-24 bg-gray-100 rounded overflow-hidden relative">
                    {reader.capaURL ? (
                      <img
                        src={`/api/proxy-image?url=${encodeURIComponent(reader.capaURL)}`}
                        alt={reader.livro}
                        className="w-full h-full object-cover"
                        onError={(e) => {
                          e.target.onerror = null
                          e.target.style.display = 'none'
                          e.target.nextSibling.style.display = 'flex'
                        }}
                      />
                    ) : null}
                    <div
                      className="w-full h-full flex items-center justify-center text-gray-400 text-2xl absolute top-0 left-0"
                      style={{ display: reader.capaURL ? 'none' : 'flex' }}
                    >
                      📚
                    </div>
                  </div>

                  {/* Reading Details */}
                  <div className="flex-1 min-w-0">
                    <p className="font-semibold text-sm text-gray-800 truncate">
                      {reader.user}
                    </p>
                    <p className="text-xs text-gray-600 line-clamp-2 leading-tight mt-1">
                      {reader.livro}
                    </p>

                    {/* Progress Bar */}
                    <div className="mt-2">
                      <div className="flex justify-between items-center text-xs text-gray-500 mb-1">
                        <span>{reader.progresso}%</span>
                        {reader.progresso === 100 && (
                          <span className="text-green-600 font-semibold">✓ Completo</span>
                        )}
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-1.5">
                        <div
                          className={`h-1.5 rounded-full ${
                            reader.progresso === 100 ? 'bg-green-500' : 'bg-blue-500'
                          }`}
                          style={{ width: `${reader.progresso}%` }}
                        ></div>
                      </div>
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          </div>

          {/* Footer */}
          <div className="border-t border-gray-200 px-4 py-2 bg-gray-50 rounded-b-lg">
            <p className="text-sm text-gray-600 text-center">
              {readers.length} {readers.length === 1 ? 'leitura' : 'leituras'}
            </p>
          </div>
        </>
      )}

      {/* Empty State */}
      {!loading && !error && readers.length === 0 && (
        <div className="p-6 text-center text-gray-500">
          <p>Nenhuma leitura encontrada neste país.</p>
        </div>
      )}
    </div>
  )
}
