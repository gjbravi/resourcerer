import { createTheme } from "@mui/material/styles";
import type {} from "@mui/x-data-grid/themeAugmentation";

export const theme = createTheme({
  cssVariables: { colorSchemeSelector: "class" },
  colorSchemes: {
    light: {
      palette: {
        primary: { main: "#4f46e5" },
        background: { default: "#f8fafc", paper: "#ffffff" },
      },
    },
    dark: {
      palette: {
        primary: { main: "#a5b4fc" },
        background: { default: "#0b1120", paper: "#111827" },
      },
    },
  },
  shape: { borderRadius: 10 },
  typography: {
    fontFamily: 'Inter, ui-sans-serif, system-ui, -apple-system, "Segoe UI", sans-serif',
    h1: { fontSize: "1.5rem", fontWeight: 600 },
    h2: { fontSize: "1rem", fontWeight: 600 },
    overline: { fontWeight: 600, letterSpacing: "0.06em" },
    button: { textTransform: "none", fontWeight: 600 },
  },
  components: {
    MuiPaper: { defaultProps: { variant: "outlined" } },
    MuiButton: { defaultProps: { disableElevation: true } },
    MuiChip: { defaultProps: { size: "small" } },
    MuiTooltip: { defaultProps: { arrow: true } },
    MuiDataGrid: {
      styleOverrides: {
        root: { border: 0, "--DataGrid-containerBackground": "transparent" },
        cell: { display: "flex", alignItems: "center" },
      },
    },
  },
});

/** Tabular numbers keep resource columns aligned. */
export const numeric = { fontVariantNumeric: "tabular-nums" } as const;
