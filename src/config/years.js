/**
 * Year configuration for the reading challenge.
 * The list of selectable years is dynamic (fetched from the API via
 * /years); this module only provides the default fallback.
 */

// Current year, used as the default selection when the URL has no year.
export const DEFAULT_YEAR = new Date().getFullYear()
