/**
 * @typedef {Object} MonthConfig
 * @property {string} name - Month name in Portuguese
 * @property {string} color - Hex color code for the month
 * @property {string[]} countries - List of country ISO codes for this month
 */

/**
 * Month configurations for the reading challenge
 * Each month has a color and a list of countries (ISO 3166-1 Alpha-3)
 * @type {MonthConfig[]}
 */
export const months = [
  {
    name: 'Janeiro',
    color: '#FF1744', // Vibrant Red
    countries: ['BRA', 'GUF', 'SUR', 'GUY', 'VEN', 'COL', 'ECU', 'PER', 'BOL', 'CHL', 'PRY', 'ARG', 'URY']
  },
  {
    name: 'Fevereiro',
    color: '#00E5FF', // Bright Cyan
    countries: ['CHN', 'JPN', 'KOR', 'PRK', 'PHL', 'IDN', 'BTN', 'MNG', 'LAO', 'NPL', 'VNM', 'BRN', 'MYS', 'TLS', 'KAZ', 'KHM', 'THA', 'MMR', 'SGP', 'TWN']
  },
  {
    name: 'Março',
    color: '#FFD600', // Vivid Yellow
    countries: ['PRT', 'ESP', 'FRA', 'AND', 'MCO', 'ITA', 'MLT', 'VAT', 'SMR']
  },
  {
    name: 'Abril',
    color: '#00E676', // Bright Green
    countries: ['GNQ', 'GAB', 'COG', 'COD', 'UGA', 'KEN', 'RWA', 'BDI', 'TZA', 'AGO', 'ZMB', 'MWI', 'MOZ', 'ZWE', 'BWA', 'NAM', 'ZAF', 'LSO', 'SWZ', 'MDG', 'STP', 'MUS', 'SYC', 'COM']
  },
  {
    name: 'Maio',
    color: '#FF6F00', // Vibrant Orange
    countries: ['GTM', 'BLZ', 'SLV', 'HND', 'NIC', 'CRI', 'PAN', 'BHS', 'CUB', 'JAM', 'HTI', 'DOM', 'PRI', 'KNA', 'ATG', 'MSR', 'DMA', 'LCA', 'BRB', 'GRD', 'TTO', 'VCT']
  },
  {
    name: 'Junho',
    color: '#D500F9', // Bright Purple
    countries: ['GBR', 'IRL', 'ISL', 'NOR', 'SWE', 'FIN']
  },
  {
    name: 'Julho',
    color: '#2979FF', // Vivid Blue
    countries: ['USA', 'CAN', 'MEX', 'GRL']
  },
  {
    name: 'Agosto',
    color: '#FF4081', // Hot Pink
    countries: ['AUS', 'PNG', 'NZL', 'FJI', 'SLB', 'VUT', 'WSM', 'KIR', 'TON', 'FSM', 'PLW', 'MHL', 'NRU', 'TUV']
  },
  {
    name: 'Setembro',
    color: '#1DE9B6', // Bright Teal
    countries: ['CHE', 'BEL', 'LUX', 'NLD', 'DEU', 'DNK', 'POL', 'CZE', 'AUT', 'LIE']
  },
  {
    name: 'Outubro',
    color: '#FF9100', // Bright Amber
    countries: ['SVK', 'HUN', 'SVN', 'HRV', 'BIH', 'MNE', 'SRB', 'ALB', 'GRC', 'MKD', 'BGR', 'ROU', 'MDA', 'UKR', 'BLR', 'LTU', 'LVA', 'EST', 'RUS']
  },
  {
    name: 'Novembro',
    color: '#651FFF', // Vivid Indigo
    countries: ['MAR', 'DZA', 'TUN', 'ESH', 'MRT', 'SEN', 'GMB', 'GNB', 'GIN', 'SLE', 'LBR', 'CIV', 'MLI', 'BFA', 'GHA', 'TGO', 'BEN', 'NER', 'NGA', 'LBY', 'TCD', 'CMR', 'CAF', 'EGY', 'SDN', 'SSD', 'ETH', 'SOM', 'ERI', 'DJI', 'CPV']
  },
  {
    name: 'Dezembro',
    color: '#F50057', // Vivid Rose
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
