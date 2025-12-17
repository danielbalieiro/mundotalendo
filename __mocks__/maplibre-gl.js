class Map {
  constructor() {
    this.on = jest.fn()
    this.off = jest.fn()
    this.remove = jest.fn()
    this.addSource = jest.fn()
    this.addLayer = jest.fn()
    this.setPaintProperty = jest.fn()
    this.getSource = jest.fn()
    this.queryRenderedFeatures = jest.fn(() => [])
    this.getCanvas = jest.fn(() => ({
      style: { cursor: '' }
    }))
  }
}

module.exports = {
  Map,
  NavigationControl: jest.fn(),
  Popup: jest.fn(() => ({
    setLngLat: jest.fn().mockReturnThis(),
    setHTML: jest.fn().mockReturnThis(),
    addTo: jest.fn().mockReturnThis(),
    remove: jest.fn(),
  })),
}
