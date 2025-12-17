import { months, getCountryColorMap, getCountryColor, getMonthByCountry } from '../months';

describe('months configuration', () => {
  describe('months array', () => {
    it('should have 12 months', () => {
      expect(months).toHaveLength(12);
    });

    it('should have all required properties for each month', () => {
      months.forEach((month, index) => {
        expect(month).toHaveProperty('name');
        expect(month).toHaveProperty('color');
        expect(month).toHaveProperty('countries');
        expect(typeof month.name).toBe('string');
        expect(typeof month.color).toBe('string');
        expect(Array.isArray(month.countries)).toBe(true);
      });
    });

    it('should have valid hex color codes', () => {
      const hexColorRegex = /^#[0-9A-F]{6}$/i;
      months.forEach(month => {
        expect(month.color).toMatch(hexColorRegex);
      });
    });

    it('should have Portuguese month names', () => {
      const expectedMonths = [
        'Janeiro', 'Fevereiro', 'Março', 'Abril', 'Maio', 'Junho',
        'Julho', 'Agosto', 'Setembro', 'Outubro', 'Novembro', 'Dezembro'
      ];

      months.forEach((month, index) => {
        expect(month.name).toBe(expectedMonths[index]);
      });
    });

    it('should have 3-letter ISO country codes', () => {
      months.forEach(month => {
        month.countries.forEach(iso => {
          expect(iso).toHaveLength(3);
          expect(iso).toMatch(/^[A-Z]{3}$/);
        });
      });
    });

    it('should not have duplicate countries across months', () => {
      const allCountries = months.flatMap(month => month.countries);
      const uniqueCountries = new Set(allCountries);
      expect(allCountries.length).toBe(uniqueCountries.size);
    });

    it('should have at least one country per month', () => {
      months.forEach(month => {
        expect(month.countries.length).toBeGreaterThan(0);
      });
    });
  });

  describe('getCountryColorMap()', () => {
    it('should return an object', () => {
      const map = getCountryColorMap();
      expect(typeof map).toBe('object');
      expect(map).not.toBeNull();
    });

    it('should map all countries to colors', () => {
      const map = getCountryColorMap();
      const totalCountries = months.reduce((sum, month) => sum + month.countries.length, 0);
      expect(Object.keys(map)).toHaveLength(totalCountries);
    });

    it('should map Brazil to Janeiro color', () => {
      const map = getCountryColorMap();
      const janeiroColor = months.find(m => m.name === 'Janeiro').color;
      expect(map['BRA']).toBe(janeiroColor);
      expect(map['BRA']).toBe('#FF1744');
    });

    it('should map USA to Julho color', () => {
      const map = getCountryColorMap();
      const julhoColor = months.find(m => m.name === 'Julho').color;
      expect(map['USA']).toBe(julhoColor);
      expect(map['USA']).toBe('#2979FF');
    });

    it('should have all colors as hex codes', () => {
      const map = getCountryColorMap();
      const hexColorRegex = /^#[0-9A-F]{6}$/i;
      Object.values(map).forEach(color => {
        expect(color).toMatch(hexColorRegex);
      });
    });
  });

  describe('getCountryColor()', () => {
    it('should return the correct color for Brazil', () => {
      expect(getCountryColor('BRA')).toBe('#FF1744');
    });

    it('should return the correct color for USA', () => {
      expect(getCountryColor('USA')).toBe('#2979FF');
    });

    it('should return the correct color for Japan', () => {
      expect(getCountryColor('JPN')).toBe('#00E5FF');
    });

    it('should return white for unknown country', () => {
      expect(getCountryColor('XXX')).toBe('#FFFFFF');
    });

    it('should return white for empty string', () => {
      expect(getCountryColor('')).toBe('#FFFFFF');
    });

    it('should return white for undefined', () => {
      expect(getCountryColor(undefined)).toBe('#FFFFFF');
    });

    it('should return white for null', () => {
      expect(getCountryColor(null)).toBe('#FFFFFF');
    });

    it('should be case-sensitive for ISO codes', () => {
      expect(getCountryColor('bra')).toBe('#FFFFFF');
      expect(getCountryColor('BRA')).toBe('#FF1744');
    });
  });

  describe('getMonthByCountry()', () => {
    it('should return Janeiro for Brazil', () => {
      const month = getMonthByCountry('BRA');
      expect(month).not.toBeNull();
      expect(month.name).toBe('Janeiro');
    });

    it('should return Julho for USA', () => {
      const month = getMonthByCountry('USA');
      expect(month).not.toBeNull();
      expect(month.name).toBe('Julho');
    });

    it('should return Fevereiro for Japan', () => {
      const month = getMonthByCountry('JPN');
      expect(month).not.toBeNull();
      expect(month.name).toBe('Fevereiro');
    });

    it('should return null for unknown country', () => {
      expect(getMonthByCountry('XXX')).toBeNull();
    });

    it('should return null for empty string', () => {
      expect(getMonthByCountry('')).toBeNull();
    });

    it('should return complete month object', () => {
      const month = getMonthByCountry('BRA');
      expect(month).toHaveProperty('name');
      expect(month).toHaveProperty('color');
      expect(month).toHaveProperty('countries');
      expect(Array.isArray(month.countries)).toBe(true);
    });

    it('should return same reference as months array', () => {
      const month = getMonthByCountry('BRA');
      const janeiro = months.find(m => m.name === 'Janeiro');
      expect(month).toBe(janeiro);
    });
  });

  describe('edge cases and integration', () => {
    it('should maintain consistency between getCountryColor and getMonthByCountry', () => {
      const testCountries = ['BRA', 'USA', 'JPN', 'FRA', 'AUS'];

      testCountries.forEach(iso => {
        const color = getCountryColor(iso);
        const month = getMonthByCountry(iso);
        expect(month.color).toBe(color);
      });
    });

    it('should have all months represented in color map', () => {
      const map = getCountryColorMap();
      const uniqueColors = new Set(Object.values(map));
      expect(uniqueColors.size).toBe(12);
    });
  });
});
