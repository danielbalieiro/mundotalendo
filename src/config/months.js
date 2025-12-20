/**
 * @typedef {Object} TierColors
 * @property {string} tier1 - Color for 0-20% progress (lightest)
 * @property {string} tier2 - Color for 21-40% progress (light)
 * @property {string} tier3 - Color for 41-60% progress (medium)
 * @property {string} tier4 - Color for 61-80% progress (dark)
 * @property {string} tier5 - Color for 81-100% progress (vibrant full color)
 */

/**
 * @typedef {Object} MonthConfig
 * @property {string} name - Month name in Portuguese
 * @property {string} color - Hex color code for the month (tier5 for backward compatibility)
 * @property {TierColors} colors - 5-tier color gradient based on progress
 * @property {string[]} countries - List of country ISO codes for this month
 */

/**
 * Month configurations for the reading challenge
 * Each month has 5 color tiers based on reading progress
 * @type {MonthConfig[]}
 */
export const months = [
  {
    name: 'Janeiro',
    color: '#FF1744', // Vibrant Red (tier5)
    colors: {
      tier1: '#FFA3B5', // 0-20%
      tier2: '#FF869E', // 21-40%
      tier3: '#FF6885', // 41-60%
      tier4: '#FF4F71', // 61-80%
      tier5: '#FF1744'  // 81-100%
    },
    countries: ['BRA', 'GUF', 'SUR', 'GUY', 'VEN', 'COL', 'ECU', 'PER', 'BOL', 'CHL', 'PRY', 'ARG', 'URY']
  },
  {
    name: 'Fevereiro',
    color: '#00E5FF', // Bright Cyan (tier5)
    colors: {
      tier1: '#B9F8FF', // 0-20%
      tier2: '#92F4FF', // 21-40%
      tier3: '#6DF0FF', // 41-60%
      tier4: '#42ECFF', // 61-80%
      tier5: '#00E5FF'  // 81-100%
    },
    countries: ['CHN', 'JPN', 'KOR', 'PRK', 'PHL', 'IDN', 'BTN', 'MNG', 'LAO', 'NPL', 'VNM', 'BRN', 'MYS', 'TLS', 'KAZ', 'KHM', 'THA', 'MMR', 'SGP', 'TWN']
  },
  {
    name: 'Março',
    color: '#FFD600', // Lemon Yellow (tier5)
    colors: {
      tier1: '#FFF8D4', // 0-20%
      tier2: '#FFF2AA', // 21-40%
      tier3: '#FFEB81', // 41-60%
      tier4: '#FFE24A', // 61-80%
      tier5: '#FFD600'  // 81-100%
    },
    countries: ['PRT', 'ESP', 'FRA', 'AND', 'MCO', 'ITA', 'MLT', 'VAT', 'SMR']
  },
  {
    name: 'Abril',
    color: '#00E676', // Vibrant Green (tier5)
    colors: {
      tier1: '#D5FFEB', // 0-20%
      tier2: '#ACFFD7', // 21-40%
      tier3: '#71FFBA', // 41-60%
      tier4: '#1EFF92', // 61-80%
      tier5: '#00E676'  // 81-100%
    },
    countries: ['GNQ', 'GAB', 'COG', 'COD', 'UGA', 'KEN', 'RWA', 'BDI', 'TZA', 'AGO', 'ZMB', 'MWI', 'MOZ', 'ZWE', 'BWA', 'NAM', 'ZAF', 'LSO', 'SWZ', 'MDG', 'STP', 'MUS', 'SYC', 'COM']
  },
  {
    name: 'Maio',
    color: '#FF6F00', // Intense Orange (tier5)
    colors: {
      tier1: '#FFC99E', // 0-20%
      tier2: '#FFB67D', // 21-40%
      tier3: '#FFA159', // 41-60%
      tier4: '#FF8B31', // 61-80%
      tier5: '#FF6F00'  // 81-100%
    },
    countries: ['GTM', 'BLZ', 'SLV', 'HND', 'NIC', 'CRI', 'PAN', 'BHS', 'CUB', 'JAM', 'HTI', 'DOM', 'PRI', 'KNA', 'ATG', 'MSR', 'DMA', 'LCA', 'BRB', 'GRD', 'TTO', 'VCT']
  },
  {
    name: 'Junho',
    color: '#D500F9', // Vibrant Purple (tier5)
    colors: {
      tier1: '#F1A3FF', // 0-20%
      tier2: '#ED84FF', // 21-40%
      tier3: '#E85FFF', // 41-60%
      tier4: '#E236FF', // 61-80%
      tier5: '#D500F9'  // 81-100%
    },
    countries: ['GBR', 'IRL', 'ISL', 'NOR', 'SWE', 'FIN']
  },
  {
    name: 'Julho',
    color: '#2979FF', // Royal Blue (tier5)
    colors: {
      tier1: '#A7C8FF', // 0-20%
      tier2: '#8BB6FF', // 21-40%
      tier3: '#6CA3FF', // 41-60%
      tier4: '#4A8EFF', // 61-80%
      tier5: '#2979FF'  // 81-100%
    },
    countries: ['USA', 'CAN', 'MEX', 'GRL']
  },
  {
    name: 'Agosto',
    color: '#FF4081', // Vibrant Pink (tier5)
    colors: {
      tier1: '#FFC1D6', // 0-20%
      tier2: '#FF9EBF', // 21-40%
      tier3: '#FF80AB', // 41-60%
      tier4: '#FF6197', // 61-80%
      tier5: '#FF4081'  // 81-100%
    },
    countries: ['AUS', 'PNG', 'NZL', 'FJI', 'SLB', 'VUT', 'WSM', 'KIR', 'TON', 'FSM', 'PLW', 'MHL', 'NRU', 'TUV']
  },
  {
    name: 'Setembro',
    color: '#1DE9B6', // Bright Teal (tier5)
    colors: {
      tier1: '#CFFFF3', // 0-20%
      tier2: '#AAFFEA', // 21-40%
      tier3: '#7BFFDE', // 41-60%
      tier4: '#40FACC', // 61-80%
      tier5: '#1DE9B6'  // 81-100%
    },
    countries: ['CHE', 'BEL', 'LUX', 'NLD', 'DEU', 'DNK', 'POL', 'CZE', 'AUT', 'LIE']
  },
  {
    name: 'Outubro',
    color: '#FF9100', // Flaming Orange (tier5)
    colors: {
      tier1: '#FFDAAA', // 0-20%
      tier2: '#FFCB86', // 21-40%
      tier3: '#FFB758', // 41-60%
      tier4: '#FFA630', // 61-80%
      tier5: '#FF9100'  // 81-100%
    },
    countries: ['SVK', 'HUN', 'SVN', 'HRV', 'BIH', 'MNE', 'SRB', 'ALB', 'GRC', 'MKD', 'BGR', 'ROU', 'MDA', 'UKR', 'BLR', 'LTU', 'LVA', 'EST', 'RUS']
  },
  {
    name: 'Novembro',
    color: '#651FFF', // Deep Violet (tier5)
    colors: {
      tier1: '#AF8CFF', // 0-20%
      tier2: '#9F74FF', // 21-40%
      tier3: '#8D5AFF', // 41-60%
      tier4: '#7C41FF', // 61-80%
      tier5: '#651FFF'  // 81-100%
    },
    countries: ['MAR', 'DZA', 'TUN', 'ESH', 'MRT', 'SEN', 'GMB', 'GNB', 'GIN', 'SLE', 'LBR', 'CIV', 'MLI', 'BFA', 'GHA', 'TGO', 'BEN', 'NER', 'NGA', 'LBY', 'TCD', 'CMR', 'CAF', 'EGY', 'SDN', 'SSD', 'ETH', 'SOM', 'ERI', 'DJI', 'CPV']
  },
  {
    name: 'Dezembro',
    color: '#F50057', // Intense Magenta (tier5)
    colors: {
      tier1: '#FF8FB6', // 0-20%
      tier2: '#FF70A2', // 21-40%
      tier3: '#FF508E', // 41-60%
      tier4: '#FF2673', // 61-80%
      tier5: '#F50057'  // 81-100%
    },
    countries: ['TUR', 'CYP', 'LBN', 'ISR', 'PSE', 'JOR', 'SYR', 'IRQ', 'IRN', 'GEO', 'ARM', 'AZE', 'TKM', 'UZB', 'AFG', 'TJK', 'KGZ', 'PAK', 'SAU', 'KWT', 'BHR', 'QAT', 'ARE', 'OMN', 'YEM', 'IND', 'LKA', 'MDV', 'BGD']
  }
];

/**
 * Create a map of ISO codes to month colors
 * @returns {Object<string, string>} Map of ISO code to color
 */
export function getCountryColorMap() {
  const map = {};
  months.forEach(month => {
    month.countries.forEach(iso => {
      map[iso] = month.color;
    });
  });
  return map;
}

/**
 * Get the color for a specific country
 * @param {string} iso - ISO 3166-1 Alpha-3 code
 * @returns {string} Hex color code or default white
 */
export function getCountryColor(iso) {
  const map = getCountryColorMap();
  return map[iso] || '#FFFFFF';
}

/**
 * Get month info by country ISO code
 * @param {string} iso - ISO 3166-1 Alpha-3 code
 * @returns {MonthConfig|null} Month configuration or null
 */
export function getMonthByCountry(iso) {
  return months.find(month => month.countries.includes(iso)) || null;
}
