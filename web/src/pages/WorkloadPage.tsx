import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import Breadcrumbs from "@mui/material/Breadcrumbs";
import Card from "@mui/material/Card";
import CardActionArea from "@mui/material/CardActionArea";
import Grid from "@mui/material/Grid";
import Link from "@mui/material/Link";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import Radio from "@mui/material/Radio";
import Skeleton from "@mui/material/Skeleton";
import Stack from "@mui/material/Stack";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Typography from "@mui/material/Typography";
import { useState } from "react";
import { Link as RouterLink, useParams } from "react-router";
import type { Mode, WorkloadDetail, WorkloadRef } from "../api";
import { ConfirmDialog, Diff, ErrorAlert, ModeChip, Section, SizingChip, StatusChip } from "../components";
import { format, relativeTime, sizing } from "../format";
import { useInfo, useSetMode, useWorkload } from "../hooks";
import { t } from "../strings";
import { numeric } from "../theme";

const MODES: Mode[] = ["Off", "Auto", "Disabled"];
const RESOURCES = ["cpu", "memory"] as const;

export function WorkloadPage() {
  const p = useParams();
  const ref: WorkloadRef = { namespace: p.namespace!, kind: p.kind!, name: p.name! };
  const { data: w, error, isLoading, refetch } = useWorkload(ref);

  if (isLoading) {
    return (
      <Stack spacing={2}>
        <Skeleton width={240} height={40} />
        <Skeleton variant="rounded" height={160} />
        <Skeleton variant="rounded" height={240} />
      </Stack>
    );
  }
  if (error || !w) return <ErrorAlert error={error ?? t.detail.notFound} onRetry={() => refetch()} />;

  return (
    <Stack spacing={3}>
      <Box>
        <Breadcrumbs sx={{ mb: 1 }}>
          <Link component={RouterLink} to="/" underline="hover" color="inherit">
            {t.detail.back}
          </Link>
          <Typography color="text.secondary">{w.namespace}</Typography>
        </Breadcrumbs>
        <Stack direction="row" spacing={1.5} sx={{ alignItems: "center", flexWrap: "wrap", rowGap: 1 }}>
          <Typography variant="h1">{w.name}</Typography>
          <ModeChip mode={w.mode} />
          <StatusChip status={w.status} message={w.message} />
          <SizingChip sizing={sizing(w)} />
        </Stack>
        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
          {w.kind} · {t.detail.replicas(w.replicas)}
          {w.vpa && (
            <>
              {" · VPA "}
              <code>{w.vpa}</code>
            </>
          )}
        </Typography>
      </Box>

      {w.message && w.status !== "Ready" && (
        <Alert severity={w.status === "Error" || w.status === "Conflict" ? "warning" : "info"}>{w.message}</Alert>
      )}

      <ModeSection w={w} workloadRef={ref} />

      {w.mode === "Disabled" ? <Alert severity="info">{t.detail.disabled}</Alert> : <Recommendation w={w} />}

      <Grid container spacing={2}>
        <Grid size={{ xs: 12, lg: 7 }}>
          <Pods w={w} />
        </Grid>
        <Grid size={{ xs: 12, lg: 5 }}>
          <Events w={w} />
        </Grid>
      </Grid>
    </Stack>
  );
}

function ModeSection({ w, workloadRef }: { w: WorkloadDetail; workloadRef: WorkloadRef }) {
  const setMode = useSetMode(workloadRef);
  const info = useInfo();
  const readOnly = info.data?.readOnly ?? false;
  const [pending, setPending] = useState<Mode | null>(null);

  return (
    <Section title={t.detail.sections.mode} subtitle={w.mode !== "Disabled" ? t.mode.source(w.modeSource) : undefined}>
      <ErrorAlert error={setMode.error} />
      <Grid container spacing={2} role="radiogroup" aria-label={t.detail.sections.mode}>
        {MODES.map((m) => {
          const active = w.mode === m;
          return (
            <Grid key={m} size={{ xs: 12, md: 4 }}>
              <Card
                variant="outlined"
                sx={{ height: "100%", borderColor: active ? "primary.main" : undefined, borderWidth: active ? 2 : 1 }}
              >
                <CardActionArea
                  role="radio"
                  aria-checked={active}
                  disabled={readOnly || setMode.isPending}
                  onClick={() => !active && setPending(m)}
                  sx={{ p: 2, height: "100%", alignItems: "flex-start" }}
                >
                  <Stack direction="row" spacing={1} sx={{ alignItems: "center" }}>
                    <Radio checked={active} size="small" sx={{ p: 0 }} tabIndex={-1} />
                    <Typography sx={{ fontWeight: 600 }}>{t.mode.label[m]}</Typography>
                  </Stack>
                  <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                    {t.mode.help[m]}
                  </Typography>
                </CardActionArea>
              </Card>
            </Grid>
          );
        })}
      </Grid>
      <Typography variant="caption" color="text.secondary" component="p" sx={{ mt: 2 }}>
        {t.mode.gitops("resourcerer.io/enabled", "resourcerer.io/mode")}
      </Typography>
      <ConfirmDialog
        open={pending !== null}
        title={pending ? t.mode.confirmTitle(pending) : ""}
        confirmLabel={t.mode.confirm}
        danger={pending === "Disabled"}
        busy={setMode.isPending}
        onClose={() => setPending(null)}
        onConfirm={() => pending && setMode.mutate(pending, { onSuccess: () => setPending(null) })}
      >
        <Typography>{pending && t.mode.confirmBody[pending]}</Typography>
      </ConfirmDialog>
    </Section>
  );
}

function Recommendation({ w }: { w: WorkloadDetail }) {
  return (
    <Section title={t.detail.sections.recommendation} subtitle={t.detail.recommendationHelp} flush>
      <Box sx={{ overflowX: "auto" }}>
        <Table aria-label={t.detail.sections.recommendation}>
          <TableHead>
            <TableRow>
              <TableCell rowSpan={2}>{t.detail.container}</TableCell>
              <TableCell colSpan={3}>CPU</TableCell>
              <TableCell colSpan={3}>Memory</TableCell>
            </TableRow>
            <TableRow>
              {RESOURCES.flatMap((r) => [
                <TableCell key={`${r}-t`} sx={{ color: "text.secondary", fontSize: 12 }}>
                  {t.detail.request} → {t.detail.target}
                </TableCell>,
                <TableCell key={`${r}-r`} sx={{ color: "text.secondary", fontSize: 12 }}>
                  {t.detail.range}
                </TableCell>,
                <TableCell key={`${r}-l`} sx={{ color: "text.secondary", fontSize: 12 }}>
                  {t.detail.limit}
                </TableCell>,
              ])}
            </TableRow>
          </TableHead>
          <TableBody>
            {w.containers.map((c) => (
              <TableRow key={c.name}>
                <TableCell sx={{ fontWeight: 600 }}>{c.name}</TableCell>
                {RESOURCES.flatMap((r) => [
                  <TableCell key={`${r}-t`}>
                    {c.target ? (
                      <Diff resource={r} current={c.requests[r]} target={c.target[r]} />
                    ) : (
                      <Typography variant="body2" color="text.secondary">
                        {t.detail.noRecommendation}
                      </Typography>
                    )}
                  </TableCell>,
                  <TableCell key={`${r}-r`} sx={{ ...numeric, color: "text.secondary" }}>
                    {c.lowerBound ? `${format(r, c.lowerBound[r])} – ${format(r, c.upperBound?.[r])}` : "–"}
                  </TableCell>,
                  <TableCell key={`${r}-l`} sx={{ ...numeric, color: "text.secondary" }}>
                    {format(r, c.limits[r])}
                  </TableCell>,
                ])}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Box>
    </Section>
  );
}

function Pods({ w }: { w: WorkloadDetail }) {
  return (
    <Section title={t.detail.sections.pods} subtitle={t.detail.podsHelp} flush>
      <Box sx={{ overflowX: "auto" }}>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell>{t.detail.pod}</TableCell>
              <TableCell>{t.detail.container}</TableCell>
              <TableCell>CPU</TableCell>
              <TableCell>Memory</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {w.pods.flatMap((p) =>
              p.containers.map((c) => (
                <TableRow key={p.name + c.name}>
                  <TableCell>{p.name}</TableCell>
                  <TableCell sx={{ color: "text.secondary" }}>{c.name}</TableCell>
                  <TableCell sx={numeric}>
                    {format("cpu", c.requests.cpu)} / {format("cpu", c.limits.cpu)}
                  </TableCell>
                  <TableCell sx={numeric}>
                    {format("memory", c.requests.memory)} / {format("memory", c.limits.memory)}
                  </TableCell>
                </TableRow>
              )),
            )}
          </TableBody>
        </Table>
      </Box>
    </Section>
  );
}

function Events({ w }: { w: WorkloadDetail }) {
  return (
    <Section title={t.detail.sections.events}>
      {w.events.length === 0 ? (
        <Typography variant="body2" color="text.secondary">
          {t.detail.eventsEmpty}
        </Typography>
      ) : (
        <List dense disablePadding>
          {w.events.slice(0, 10).map((e, i) => (
            <ListItem key={i} disableGutters>
              <ListItemText
                primary={
                  <>
                    <Typography
                      component="span"
                      variant="body2"
                      sx={{ fontWeight: 600, color: e.type === "Warning" ? "error.main" : undefined }}
                    >
                      {e.reason}
                    </Typography>
                    <Typography component="span" variant="caption" color="text.secondary" sx={{ ml: 1 }}>
                      {relativeTime(e.time)}
                    </Typography>
                  </>
                }
                secondary={e.message}
              />
            </ListItem>
          ))}
        </List>
      )}
    </Section>
  );
}
