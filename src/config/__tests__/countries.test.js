import { countryNames, getCountryName } from '../countries';

describe('countries configuration', () => {
  describe('countryNames object', () => {
    it('should be an object', () => {
      expect(typeof countryNames).toBe('object');
      expect(countryNames).not.toBeNull();
    });

    it('should have ISO codes as keys', () => {
      Object.keys(countryNames).forEach(iso => {
        expect(iso).toHaveLength(3);
        expect(iso).toMatch(/^[A-Z]{3}$/);
      });
    });

    it('should have Portuguese names as values', () => {
      Object.values(countryNames).forEach(name => {
        expect(typeof name).toBe('string');
        expect(name.length).toBeGreaterThan(0);
      });
    });

    it('should contain Brazil', () => {
      expect(countryNames).toHaveProperty('BRA', 'Brasil');
    });

    it('should contain USA', () => {
      expect(countryNames).toHaveProperty('USA', 'Estados Unidos');
    });

    it('should contain common countries', () => {
      expect(countryNames['CAN']).toBe('Canadá');
      expect(countryNames['FRA']).toBe('França');
      expect(countryNames['DEU']).toBe('Alemanha');
      expect(countryNames['JPN']).toBe('Japão');
      expect(countryNames['CHN']).toBe('China');
    });

    it('should have unique country names', () => {
      const names = Object.values(countryNames);
      const uniqueNames = new Set(names);
      expect(names.length).toBe(uniqueNames.size);
    });

    it('should have a substantial number of countries', () => {
      const count = Object.keys(countryNames).length;
      expect(count).toBeGreaterThan(150); // Should have most countries
    });
  });

  describe('getCountryName()', () => {
    it('should return Portuguese name for valid ISO code', () => {
      expect(getCountryName('BRA')).toBe('Brasil');
      expect(getCountryName('USA')).toBe('Estados Unidos');
      expect(getCountryName('FRA')).toBe('França');
    });

    it('should return ISO code for unknown country', () => {
      expect(getCountryName('XXX')).toBe('XXX');
      expect(getCountryName('ZZZ')).toBe('ZZZ');
    });

    it('should return ISO code for empty string', () => {
      expect(getCountryName('')).toBe('');
    });

    it('should handle undefined gracefully', () => {
      expect(getCountryName(undefined)).toBe(undefined);
    });

    it('should handle null gracefully', () => {
      expect(getCountryName(null)).toBe(null);
    });

    it('should be case-sensitive', () => {
      // Should not find lowercase
      expect(getCountryName('bra')).toBe('bra');
      expect(getCountryName('usa')).toBe('usa');

      // Should find uppercase
      expect(getCountryName('BRA')).toBe('Brasil');
      expect(getCountryName('USA')).toBe('Estados Unidos');
    });

    it('should handle all countries in countryNames', () => {
      Object.keys(countryNames).forEach(iso => {
        const name = getCountryName(iso);
        expect(name).toBe(countryNames[iso]);
        expect(name).not.toBe(iso); // Should return name, not ISO
      });
    });

    it('should return consistent results', () => {
      const iso = 'BRA';
      const name1 = getCountryName(iso);
      const name2 = getCountryName(iso);
      expect(name1).toBe(name2);
    });

    it('should handle special characters in names', () => {
      // Countries with accents and special characters
      expect(getCountryName('CAN')).toContain('á'); // Canadá
      expect(getCountryName('PRT')).toContain('Port'); // Portugal
      expect(getCountryName('ESP')).toContain('Esp'); // Espanha
    });
  });

  describe('data integrity', () => {
    it('should have all ISO codes in uppercase', () => {
      Object.keys(countryNames).forEach(iso => {
        expect(iso).toBe(iso.toUpperCase());
      });
    });

    it('should not have empty names', () => {
      Object.values(countryNames).forEach(name => {
        expect(name.trim()).toBe(name);
        expect(name.length).toBeGreaterThan(0);
      });
    });

    it('should have valid Portuguese characters', () => {
      Object.values(countryNames).forEach(name => {
        // Should only contain valid characters (letters, spaces, hyphens, accents)
        expect(name).toMatch(/^[A-Za-zÀ-ÿ\s\-']+$/);
      });
    });
  });
});
