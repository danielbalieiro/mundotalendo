/**
 * Color tier utilities for the Mundo Tá Lendo 2026 project
 * Handles mapping of progress percentages to color tiers
 */

/**
 * Determine which color tier to use based on progress percentage
 *
 * Tier boundaries (inclusive lower, exclusive upper):
 * - Tier 1: 0-20%
 * - Tier 2: 21-40%
 * - Tier 3: 41-60%
 * - Tier 4: 61-80%
 * - Tier 5: 81-100%
 *
 * @param {number} progress - Progress value from 0-100
 * @returns {1|2|3|4|5} Tier number (1 = lightest, 5 = vibrant full color)
 */
export function getColorTier(progress) {
  // Coerce to number and clamp to valid range
  const p = Math.max(0, Math.min(100, Number(progress) || 0))

  // Tier boundaries (values at exact boundaries belong to lower tier)
  if (p <= 20) return 1
  if (p <= 40) return 2
  if (p <= 60) return 3
  if (p <= 80) return 4
  return 5
}

/**
 * Get the color for a specific country based on its progress
 *
 * @param {string} iso - ISO 3166-1 Alpha-3 country code
 * @param {number} progress - Progress value from 0-100
 * @param {Array<Object>} monthsConfig - Months configuration array
 * @returns {string} Hex color code
 */
export function getCountryProgressColor(iso, progress, monthsConfig) {
  // Find the month this country belongs to
  const month = monthsConfig.find(m => m.countries.includes(iso))

  if (!month) {
    // Country not assigned to any month - use default gray
    console.warn(`Country ${iso} not assigned to any month`)
    return '#F5F5F5'
  }

  // Get tier based on progress
  const tier = getColorTier(progress)

  // Return the tier color
  return month.colors[`tier${tier}`]
}

/**
 * Get a human-readable tier label for display purposes
 *
 * @param {number} progress - Progress value from 0-100
 * @returns {string} Portuguese tier label
 */
export function getTierLabel(progress) {
  const tier = getColorTier(progress)

  const labels = {
    1: 'Iniciado (0-20%)',
    2: 'Em Progresso (21-40%)',
    3: 'No Meio (41-60%)',
    4: 'Quase Completo (61-80%)',
    5: 'Completo (81-100%)'
  }

  return labels[tier]
}
