import Box from "@mui/material/Box";
import Chip from "@mui/material/Chip";
import Grid from "@mui/material/Grid";
import InputAdornment from "@mui/material/InputAdornment";
import Link from "@mui/material/Link";
import MenuItem from "@mui/material/MenuItem";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import Search from "@mui/icons-material/Search";
import { DataGrid, type GridColDef } from "@mui/x-data-grid";
import { useMemo } from "react";
import { Link as RouterLink, useNavigate, useSearchParams } from "react-router";
import { workloadPath, type ResourceName, type Workload } from "../api";
import { Diff, ErrorAlert, ModeChip, SizingChip, Stat, StatusChip } from "../components";
import { sizing } from "../format";
import { useInfo, useWorkloads } from "../hooks";
import { t } from "../strings";

export const FILTERS = ["all", "under", "over", "learning", "disabled"] as const;
export type Filter = (typeof FILTERS)[number];

export function matches(w: Workload, f: Filter): boolean {
  switch (f) {
    case "all":
      return true;
    case "under":
      return sizing(w) === "under";
    case "over":
      return sizing(w) === "over";
    case "learning":
      return w.status === "Pending" || w.status === "LowConfidence";
    case "disabled":
      return w.mode === "Disabled";
  }
}

function Resource({ w, resource }: { w: Workload; resource: ResourceName }) {
  const multi = w.containers.length > 1;
  return (
    <Stack spacing={0.25} sx={{ py: 1 }}>
      {w.containers.map((c) => (
        <Box key={c.name} sx={{ display: "flex", gap: 1, alignItems: "baseline" }}>
          {multi && (
            <Typography variant="caption" color="text.secondary">
              {c.name}
            </Typography>
          )}
          <Diff resource={resource} current={c.requests[resource]} target={c.target?.[resource]} compact />
        </Box>
      ))}
    </Stack>
  );
}

const sizingOrder = { under: 0, over: 1, fit: 2, unknown: 3 };

const columns: GridColDef<Workload>[] = [
  {
    field: "name",
    headerName: t.list.columns.workload,
    flex: 1.4,
    minWidth: 200,
    valueGetter: (_, w) => `${w.namespace}/${w.name}`,
    renderCell: ({ row: w }) => (
      <Box sx={{ py: 1, minWidth: 0 }}>
        <Link
          component={RouterLink}
          to={workloadPath(w)}
          underline="hover"
          sx={{ fontWeight: 600 }}
          onClick={(e) => e.stopPropagation()}
        >
          {w.name}
        </Link>
        <Typography variant="caption" color="text.secondary" component="div" noWrap>
          {w.namespace} · {w.kind} · {w.replicas}
        </Typography>
      </Box>
    ),
  },
  { field: "mode", headerName: t.list.columns.mode, width: 140, renderCell: ({ row }) => <ModeChip mode={row.mode} /> },
  {
    field: "status",
    headerName: t.list.columns.status,
    width: 130,
    renderCell: ({ row }) => <StatusChip status={row.status} message={row.message} />,
  },
  {
    field: "sizing",
    headerName: t.list.columns.sizing,
    width: 180,
    valueGetter: (_, w) => sizingOrder[sizing(w)],
    renderCell: ({ row }) => <SizingChip sizing={sizing(row)} />,
  },
  {
    field: "cpu",
    headerName: t.list.columns.cpu,
    flex: 1,
    minWidth: 160,
    sortable: false,
    renderCell: ({ row }) => <Resource w={row} resource="cpu" />,
  },
  {
    field: "memory",
    headerName: t.list.columns.memory,
    flex: 1,
    minWidth: 160,
    sortable: false,
    renderCell: ({ row }) => <Resource w={row} resource="memory" />,
  },
];

export function WorkloadsPage() {
  const { data, error, isLoading, refetch } = useWorkloads();
  const info = useInfo();
  const navigate = useNavigate();
  const [params, setParams] = useSearchParams();
  const filter = (FILTERS as readonly string[]).includes(params.get("filter") ?? "")
    ? (params.get("filter") as Filter)
    : "all";
  const namespace = params.get("ns") ?? "";
  const search = params.get("q") ?? "";

  const set = (key: string, value: string, empty = "") =>
    setParams(
      (p) => {
        if (value === empty) p.delete(key);
        else p.set(key, value);
        return p;
      },
      { replace: true },
    );

  const all = useMemo(() => data ?? [], [data]);
  const namespaces = useMemo(() => [...new Set(all.map((w) => w.namespace))].sort(), [all]);
  const scoped = useMemo(() => {
    const q = search.trim().toLowerCase();
    return all.filter(
      (w) =>
        (!namespace || w.namespace === namespace) &&
        (!q || `${w.namespace}/${w.kind}/${w.name}`.toLowerCase().includes(q)),
    );
  }, [all, namespace, search]);
  const rows = useMemo(() => scoped.filter((w) => matches(w, filter)), [scoped, filter]);

  return (
    <Stack spacing={3}>
      <Box>
        <Typography variant="h1">{t.list.title}</Typography>
        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
          {t.list.subtitle(info.data?.selection ?? "all")}
        </Typography>
      </Box>

      <Grid container spacing={2}>
        <Grid size={{ xs: 6, md: 3 }}>
          <Stat label={t.list.stats.total} value={scoped.length} />
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <Stat label={t.list.stats.managed} value={scoped.filter((w) => w.mode !== "Disabled").length} />
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <Stat
            label={t.list.stats.under}
            value={scoped.filter((w) => matches(w, "under")).length}
            color="error.main"
          />
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <Stat
            label={t.list.stats.over}
            value={scoped.filter((w) => matches(w, "over")).length}
            color="success.main"
          />
        </Grid>
      </Grid>

      <ErrorAlert error={error} onRetry={() => refetch()} />

      <Paper>
        <Stack
          direction={{ xs: "column", md: "row" }}
          spacing={2}
          sx={{ p: 2, borderBottom: 1, borderColor: "divider", alignItems: { md: "center" } }}
        >
          <Stack
            direction="row"
            spacing={1}
            sx={{ flexWrap: "wrap", rowGap: 1, flex: 1 }}
            role="group"
            aria-label="Filters"
          >
            {FILTERS.map((f) => {
              const count = scoped.filter((w) => matches(w, f)).length;
              const active = filter === f;
              return (
                <Chip
                  key={f}
                  label={`${t.list.filters[f]} · ${count}`}
                  color={active ? "primary" : "default"}
                  variant={active ? "filled" : "outlined"}
                  onClick={() => set("filter", f, "all")}
                  aria-pressed={active}
                />
              );
            })}
          </Stack>
          <TextField
            select
            size="small"
            label={t.list.namespace}
            value={namespace}
            onChange={(e) => set("ns", e.target.value)}
            slotProps={{ select: { displayEmpty: true }, inputLabel: { shrink: true } }}
            sx={{ minWidth: 180 }}
          >
            <MenuItem value="">{t.list.allNamespaces}</MenuItem>
            {namespaces.map((n) => (
              <MenuItem key={n} value={n}>
                {n}
              </MenuItem>
            ))}
          </TextField>
          <TextField
            size="small"
            placeholder={t.list.search}
            value={search}
            onChange={(e) => set("q", e.target.value)}
            slotProps={{
              input: {
                startAdornment: (
                  <InputAdornment position="start">
                    <Search fontSize="small" />
                  </InputAdornment>
                ),
              },
              htmlInput: { "aria-label": t.list.search },
            }}
            sx={{ minWidth: 220 }}
          />
        </Stack>
        <DataGrid
          rows={rows}
          columns={columns}
          getRowId={(w) => `${w.namespace}/${w.kind}/${w.name}`}
          loading={isLoading}
          getRowHeight={() => "auto"}
          disableRowSelectionOnClick
          onRowClick={({ row }) => navigate(workloadPath(row))}
          initialState={{
            sorting: { sortModel: [{ field: "sizing", sort: "asc" }] },
            pagination: { paginationModel: { pageSize: 25 } },
          }}
          pageSizeOptions={[25, 50, 100]}
          localeText={{ noRowsLabel: t.list.empty }}
          sx={{ "& .MuiDataGrid-row": { cursor: "pointer" }, minHeight: 300 }}
          autoHeight
        />
      </Paper>
    </Stack>
  );
}
