import AppBar from "@mui/material/AppBar";
import Box from "@mui/material/Box";
import Chip from "@mui/material/Chip";
import Container from "@mui/material/Container";
import IconButton from "@mui/material/IconButton";
import LinearProgress from "@mui/material/LinearProgress";
import Toolbar from "@mui/material/Toolbar";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import { useColorScheme } from "@mui/material/styles";
import DarkModeOutlined from "@mui/icons-material/DarkModeOutlined";
import LightModeOutlined from "@mui/icons-material/LightModeOutlined";
import Tune from "@mui/icons-material/Tune";
import { useIsFetching } from "@tanstack/react-query";
import { Link, Outlet } from "react-router";
import { useInfo } from "./hooks";
import { t } from "./strings";

function ThemeToggle() {
  const { mode, systemMode, setMode } = useColorScheme();
  const current = mode === "system" ? systemMode : mode;
  if (!current) return null;
  const next = current === "dark" ? "light" : "dark";
  return (
    <Tooltip title={t.theme[next]}>
      <IconButton onClick={() => setMode(next)} aria-label={t.theme[next]} color="inherit">
        {current === "dark" ? <LightModeOutlined /> : <DarkModeOutlined />}
      </IconButton>
    </Tooltip>
  );
}

export function Layout() {
  const fetching = useIsFetching();
  const info = useInfo();
  return (
    <Box sx={{ minHeight: "100vh", bgcolor: "background.default" }}>
      <AppBar position="sticky" color="inherit" elevation={0} sx={{ borderBottom: 1, borderColor: "divider" }}>
        <Toolbar sx={{ gap: 1.5 }}>
          <Box
            component={Link}
            to="/"
            sx={{ display: "flex", alignItems: "center", gap: 1, color: "inherit", textDecoration: "none" }}
          >
            <Tune color="primary" />
            <Typography variant="subtitle1" sx={{ fontWeight: 700 }}>
              {t.app}
            </Typography>
          </Box>
          <Typography variant="caption" color="text.secondary" sx={{ display: { xs: "none", sm: "block" } }}>
            {t.tagline}
          </Typography>
          <Box sx={{ flex: 1 }} />
          {info.data?.readOnly && <Chip label={t.readOnly} variant="outlined" />}
          {info.data && (
            <Typography variant="caption" color="text.secondary">
              {info.data.version}
            </Typography>
          )}
          <ThemeToggle />
        </Toolbar>
        <LinearProgress sx={{ height: 2, visibility: fetching ? "visible" : "hidden" }} />
      </AppBar>
      <Container maxWidth="xl" sx={{ py: 4 }}>
        <Outlet />
      </Container>
    </Box>
  );
}
