/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/server/web/**/*.{html,js}",
    "./web/**/*.{html,js}"
  ],
  theme: {
    extend: {
      colors: {
        'bg-primary': '#0f0f23',
        'bg-surface': '#1a1a3e',
        'text-primary': '#e5e7eb',
        'accent': '#60a5fa',
        'success': '#34d399',
        'error': '#f87171',
        'warning': '#fbbf24',
      },
      fontFamily: {
        'sans': ['Inter', 'system-ui', 'sans-serif'],
        'mono': ['JetBrains Mono', 'monospace'],
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
      },
    },
  },
  plugins: [
    require('daisyui'),
  ],
  daisyui: {
    themes: ["dark"], // We'll use dark theme only
    darkTheme: "dark",
    base: true,
    styled: true,
    utils: true,
  },
}
