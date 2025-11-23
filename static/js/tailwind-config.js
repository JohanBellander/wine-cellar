tailwind.config = {
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        "primary": "#bfa884",
        "background-light": "#fdfcf7",
        "background-dark": "#2a2621",
        "champagne-light": "#f5eadd",
        "champagne-dark": "#3c352a",
        "champagne-gold": "#C0A88F", 
        "prose-light": "#4a4137",
        "prose-dark": "#d1c7b8",
      },
      fontFamily: {
        "display": ["Playfair Display", "serif"],
        "body": ["Manrope", "sans-serif"],
        "script": ["Great Vibes", "cursive"],
        "serif": ["Playfair Display", "serif"],
        "sans": ["Manrope", "sans-serif"],
      },
      borderRadius: {"DEFAULT": "0.5rem", "lg": "0.75rem", "xl": "1rem", "full": "9999px"},
    },
  },
}
