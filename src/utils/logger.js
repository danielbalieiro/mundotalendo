/**
 * Logger utility for conditional logging
 * Only logs in development mode
 */
export const logger = {
  debug: (...args) => {
    if (process.env.NODE_ENV === 'development') {
      console.log(...args)
    }
  },
  warn: (...args) => {
    if (process.env.NODE_ENV === 'development') {
      console.warn(...args)
    }
  },
  error: (...args) => {
    console.error(...args)
  }
}
