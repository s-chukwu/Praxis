/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./app/**/*.{ts,tsx}', './components/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        page: '#F7F8FA',
        surface: '#FFFFFF',
        ink: '#14171F',
        muted: '#5D6472',
        line: '#E5E7EB',
        brand: { DEFAULT: '#0F7B5F', hover: '#0B6650', soft: '#E7F4EF', ink: '#0A5642' },
        progress: { DEFAULT: '#C98417', soft: '#FBF1DF' },
        danger: { DEFAULT: '#C2412D', soft: '#FBEAE6' },
      },
      // Self-hosted from npm, so builds never depend on reaching Google Fonts.
      fontFamily: { sans: ['"Figtree Variable"', 'system-ui', 'sans-serif'] },
      borderRadius: { control: '10px', panel: '14px' },
      keyframes: {
        rise: { from: { opacity: '0', transform: 'translateY(8px)' }, to: { opacity: '1', transform: 'none' } },
        pop: { '0%': { transform: 'scale(0.6)' }, '60%': { transform: 'scale(1.12)' }, '100%': { transform: 'scale(1)' } },
        sweep: { from: { transform: 'translateX(-100%)' }, to: { transform: 'translateX(100%)' } },
        draw: { from: { strokeDashoffset: '48' }, to: { strokeDashoffset: '0' } },
      },
      animation: {
        rise: 'rise 420ms cubic-bezier(0.16, 1, 0.3, 1) both',
        pop: 'pop 320ms cubic-bezier(0.16, 1, 0.3, 1)',
        sweep: 'sweep 1.4s ease-in-out infinite',
        draw: 'draw 600ms cubic-bezier(0.16, 1, 0.3, 1) 150ms both',
      },
      boxShadow: {
        panel: '0 1px 2px rgba(20, 23, 31, 0.04), 0 4px 16px -6px rgba(20, 23, 31, 0.08)',
        lift: '0 2px 4px rgba(20, 23, 31, 0.06), 0 16px 40px -12px rgba(20, 23, 31, 0.18)',
      },
    },
  },
  plugins: [],
};
