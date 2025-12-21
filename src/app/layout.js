import './globals.css'
import { ErrorBoundary } from '@/components/ErrorBoundary'

/** @type {import("next").Metadata} */
export const metadata = {
  title: 'Mundo Tá Lendo 2026',
  description: 'Dashboard de telemetria global do desafio de leitura Mundo Tá Lendo 2026',
}

/**
 * Root layout component
 * @param {Object} props
 * @param {React.ReactNode} props.children
 */
export default function RootLayout({ children }) {
  return (
    <html lang="pt-BR">
      <body>
        <ErrorBoundary>
          {children}
        </ErrorBoundary>
      </body>
    </html>
  )
}
