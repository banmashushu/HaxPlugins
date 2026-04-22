/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        lol: {
          bg: '#0a1428',
          'bg-deep': '#060e1a',
          card: '#0f1d2e',
          'card-hover': '#142640',
          border: '#1e3450',
          'border-glow': '#1e4d8a',
          gold: '#c8aa6e',
          'gold-bright': '#f0d878',
          'gold-dim': '#8a7340',
          blue: '#0397ab',
          'blue-bright': '#0ac8e0',
          'blue-dim': '#025964',
          text: '#a8b4c2',
          'text-bright': '#e4eaf0',
          muted: '#5a6a7e',
          red: '#e74c5e',
          green: '#2dd4a0',
          purple: '#b48ef0',
        },
        tier: {
          s: '#2dd4a0',
          a: '#4da6e8',
          b: '#e8c44d',
          c: '#e88a3c',
          d: '#e74c5e',
          prismatic: '#b48ef0',
          gold: '#f0d878',
          silver: '#a8b4c2',
        },
      },
      boxShadow: {
        'glow-gold': '0 0 12px rgba(200, 170, 110, 0.3)',
        'glow-blue': '0 0 12px rgba(3, 151, 171, 0.3)',
        'glow-green': '0 0 12px rgba(45, 212, 160, 0.3)',
        'glow-red': '0 0 12px rgba(231, 76, 94, 0.3)',
        'card': '0 2px 8px rgba(0, 0, 0, 0.4)',
        'card-hover': '0 4px 20px rgba(0, 0, 0, 0.5)',
      },
      backgroundImage: {
        'card-gradient': 'linear-gradient(180deg, rgba(15, 29, 46, 0.95) 0%, rgba(10, 20, 40, 0.98) 100%)',
        'header-gradient': 'linear-gradient(90deg, #0a1428 0%, #0f2240 50%, #0a1428 100%)',
        'gold-shimmer': 'linear-gradient(90deg, #8a7340 0%, #f0d878 50%, #8a7340 100%)',
      },
    },
  },
  plugins: [],
}