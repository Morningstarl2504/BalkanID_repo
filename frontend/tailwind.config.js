/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{js,jsx,ts,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'balkan-blue': '#3B82F6',
        'balkan-purple': '#8E44AD',
        'balkan-pink': '#E83E8C',
        'balkan-yellow': '#F4D03F',
        'balkan-green': '#2ECC71',
      },
    },
  },
  plugins: [],
}