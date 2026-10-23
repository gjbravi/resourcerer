import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Chip, { type ChipProps } from "@mui/material/Chip";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogTitle from "@mui/material/DialogTitle";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import AutoMode from "@mui/icons-material/AutoMode";
import Block from "@mui/icons-material/Block";
import CheckCircleOutlined from "@mui/icons-material/CheckCircleOutlined";
import DragHandle from "@mui/icons-material/DragHandle";
import ErrorOutlined from "@mui/icons-material/ErrorOutlined";
import HelpOutline from "@mui/icons-material/HelpOutlineOutlined";
import HourglassEmpty from "@mui/icons-material/HourglassEmpty";
import Lightbulb from "@mui/icons-material/LightbulbOutlined";
import SyncProblem from "@mui/icons-material/SyncProblem";
import TrendingDown from "@mui/icons-material/TrendingDown";
import TrendingUp from "@mui/icons-material/TrendingUp";
import type { ReactElement, ReactNode } from "react";
import type { Mode, ResourceName, Status } from "./api";
import { change, format, type Sizing } from "./format";
import { t } from "./strings";
import { numeric } from "./theme";

type Spec = { color: ChipProps["color"]; icon: ReactElement };

const modeSpec: Record<Mode, Spec> = {
  Off: { color: "default", icon: <Lightbulb /> },
  Auto: { color: "primary", icon: <AutoMode /> },
  Disabled: { color: "default", icon: <Block /> },
};

/** Icon plus text, never color alone. */
export function ModeChip({ mode }: { mode: Mode }) {
  const s = modeSpec[mode];
  return (
    <Tooltip title={t.mode.help[mode]}>
      <Chip
        icon={s.icon}
        color={s.color}
        variant={mode === "Auto" ? "filled" : "outlined"}
        label={t.mode.label[mode]}
        sx={mode === "Disabled" ? { opacity: 0.7 } : undefined}
      />
    </Tooltip>
  );
}

const statusSpec: Record<Exclude<Status, "None">, Spec> = {
  Ready: { color: "success", icon: <CheckCircleOutlined /> },
  Pending: { color: "default", icon: <HourglassEmpty /> },
  LowConfidence: { color: "info", icon: <HourglassEmpty /> },
  Conflict: { color: "warning", icon: <SyncProblem /> },
  Error: { color: "error", icon: <ErrorOutlined /> },
};

export function StatusChip({ status, message }: { status: Status; message?: string }) {
  if (status === "None") return null;
  const s = statusSpec[status];
  return (
    <Tooltip title={message || t.status.help[status]}>
      <Chip icon={s.icon} color={s.color} variant="outlined" label={t.status.label[status]} />
    </Tooltip>
  );
}

const sizingSpec: Record<Exclude<Sizing, "unknown">, Spec> = {
  over: { color: "success", icon: <TrendingDown /> },
  under: { color: "error", icon: <TrendingUp /> },
  fit: { color: "default", icon: <DragHandle /> },
  unset: { color: "warning", icon: <HelpOutline /> },
};

export function SizingChip({ sizing }: { sizing: Sizing }) {
  if (sizing === "unknown") return null;
  const s = sizingSpec[sizing];
  return (
    <Tooltip title={t.sizing.help[sizing]}>
      <Chip icon={s.icon} color={s.color} variant="outlined" label={t.sizing.label[sizing]} />
    </Tooltip>
  );
}

/** request → target with the relative change; a single value when equal or no target. */
export function Diff({
  resource,
  current,
  target,
  compact,
}: {
  resource: ResourceName;
  current?: string;
  target?: string;
  compact?: boolean;
}) {
  const from = format(resource, current);
  const to = target ? format(resource, target) : from;
  if (from === to) {
    return (
      <Typography component="span" variant="body2" sx={numeric}>
        {from}
      </Typography>
    );
  }
  const pct = change(current, target);
  return (
    <Box
      component="span"
      sx={{ ...numeric, display: "inline-flex", gap: 0.75, alignItems: "baseline", whiteSpace: "nowrap" }}
    >
      <Typography component="span" variant="body2" color="text.secondary" sx={{ textDecoration: "line-through" }}>
        {from}
      </Typography>
      <Typography component="span" variant="body2" sx={{ fontWeight: 600 }}>
        → {to}
      </Typography>
      {!compact && pct !== undefined && (
        <Typography component="span" variant="caption" color={pct > 0 ? "error.main" : "success.main"}>
          {pct > 0 ? "+" : ""}
          {pct}%
        </Typography>
      )}
    </Box>
  );
}

export function Section({
  title,
  subtitle,
  actions,
  children,
  flush,
}: {
  title: ReactNode;
  subtitle?: ReactNode;
  actions?: ReactNode;
  children: ReactNode;
  flush?: boolean;
}) {
  return (
    <Paper component="section">
      <Stack
        direction="row"
        spacing={2}
        sx={{ px: 2.5, py: 1.5, borderBottom: 1, borderColor: "divider", alignItems: "center" }}
      >
        <Box sx={{ flex: 1, minWidth: 0 }}>
          <Typography variant="h2">{title}</Typography>
          {subtitle && (
            <Typography variant="body2" color="text.secondary">
              {subtitle}
            </Typography>
          )}
        </Box>
        {actions}
      </Stack>
      <Box sx={flush ? undefined : { p: 2.5 }}>{children}</Box>
    </Paper>
  );
}

export function Stat({ label, value, color }: { label: string; value: ReactNode; color?: string }) {
  return (
    <Paper sx={{ p: 2, height: "100%" }}>
      <Typography variant="overline" color="text.secondary" component="div">
        {label}
      </Typography>
      <Typography variant="h5" component="div" sx={{ ...numeric, fontWeight: 600, color }}>
        {value}
      </Typography>
    </Paper>
  );
}

export function ConfirmDialog(props: {
  open: boolean;
  title: string;
  children: ReactNode;
  confirmLabel: string;
  danger?: boolean;
  busy?: boolean;
  onConfirm: () => void;
  onClose: () => void;
}) {
  return (
    <Dialog open={props.open} onClose={props.busy ? undefined : props.onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{props.title}</DialogTitle>
      <DialogContent>{props.children}</DialogContent>
      <DialogActions sx={{ px: 3, pb: 2 }}>
        <Button onClick={props.onClose} disabled={props.busy}>
          {t.common.cancel}
        </Button>
        <Button
          variant="contained"
          color={props.danger ? "error" : "primary"}
          onClick={props.onConfirm}
          loading={props.busy}
          autoFocus
        >
          {props.confirmLabel}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export function ErrorAlert({ error, onRetry }: { error: unknown; onRetry?: () => void }) {
  if (!error) return null;
  return (
    <Alert
      severity="error"
      action={
        onRetry && (
          <Button color="inherit" size="small" onClick={onRetry}>
            {t.common.retry}
          </Button>
        )
      }
    >
      {error instanceof Error ? error.message : String(error)}
    </Alert>
  );
}
