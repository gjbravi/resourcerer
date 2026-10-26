import {
  ArrowDown,
  ArrowRight,
  ArrowUp,
  Ban,
  Check,
  CircleAlert,
  CircleCheck,
  Copy,
  Hourglass,
  Info,
  ThumbsUp,
  TriangleAlert,
  X,
  type LucideIcon,
} from "lucide-react";
import { useEffect, useState, type ReactNode } from "react";
import type { Phase, Resource, Summary } from "./api";
import { fmt, fmtCores, fmtMem, fmtMoney, Mi } from "./format";
import { dismiss, useToasts, workloadHash } from "./hooks";

const PHASES: Record<Phase, { tone: string; icon: LucideIcon; help: string }> = {
  Ready: { tone: "ok", icon: ThumbsUp, help: "Based on the full learning period." },
  Learning: { tone: "info", icon: Hourglass, help: "Preliminary: the learning period is not covered yet." },
  Pending: { tone: "muted", icon: Hourglass, help: "Collecting usage history." },
  Warning: { tone: "warn", icon: TriangleAlert, help: "Cannot be applied now." },
  Unsupported: { tone: "muted", icon: Ban, help: "This workload cannot be optimized." },
  Error: { tone: "bad", icon: CircleAlert, help: "Metrics could not be read." },
};
export const PHASE_NAMES = Object.keys(PHASES) as Phase[];
export const phaseTone = (p: Phase) => PHASES[p].tone;

export function StatusBadge({ phase, title }: { phase: Phase; title?: string }) {
  const s = PHASES[phase] ?? PHASES.Pending;
  return (
    <span className={`rz-badge t-${s.tone}`} title={title || s.help}>
      <s.icon size={13} aria-hidden /> {phase}
    </span>
  );
}

export function Ring({ value, size = 34, stroke = 4 }: { value?: number; size?: number; stroke?: number }) {
  const r = (size - stroke) / 2;
  const c = 2 * Math.PI * r;
  const v = value == null ? 0 : Math.max(0, Math.min(100, value));
  const tone = value == null ? "var(--faint)" : v >= 80 ? "var(--accent)" : v >= 50 ? "var(--warn)" : "var(--bad)";
  return (
    <span
      className="rz-ring"
      style={{ width: size, height: size }}
      role="img"
      aria-label={value == null ? "no score" : `score ${Math.round(v)}`}
    >
      <svg width={size} height={size}>
        <circle cx={size / 2} cy={size / 2} r={r} stroke="var(--track)" strokeWidth={stroke} fill="none" />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={r}
          stroke={tone}
          strokeWidth={stroke}
          fill="none"
          strokeLinecap="round"
          strokeDasharray={`${(v / 100) * c} ${c}`}
          transform={`rotate(-90 ${size / 2} ${size / 2})`}
        />
      </svg>
      <span className="rz-ringv" style={{ fontSize: size * 0.3 }}>
        {value == null ? "–" : Math.round(v)}
      </span>
    </span>
  );
}

/** Savings in the accent color with a down arrow; increases in red with an up arrow. */
export function Delta({ from, to, r, money }: { from: number; to: number; r?: Resource; money?: boolean }) {
  const d = from - to;
  const eps = money ? 0.005 : r === "cpu" ? 0.0005 : 0.5 * Mi;
  if (Math.abs(d) < eps) return <span className="rz-dim">—</span>;
  const saving = d > 0;
  const p = from ? Math.abs((d / from) * 100) : null;
  const val = money
    ? saving
      ? fmtMoney(d)
      : `+${fmtMoney(-d)}`
    : r === "cpu"
      ? `${saving ? "-" : "+"}${fmtCores(Math.abs(d))}`
      : `${saving ? "-" : "+"}${fmtMem(Math.abs(d))}`;
  return (
    <span className={`rz-delta ${saving ? "save" : "cost"}`}>
      {val}
      {p != null && <> {Math.round(p)}%</>} {saving ? <ArrowDown size={12} /> : <ArrowUp size={12} />}
    </span>
  );
}

/** Large currency figure: faded symbol, bold dollars, small cents and unit. */
export function Money({ value, unit = "/mo" }: { value: number; unit?: string }) {
  const abs = Math.abs(value);
  const dollars = Math.floor(abs).toLocaleString();
  const cents = (abs % 1).toFixed(2).slice(1);
  return (
    <div className="rz-money">
      {value < 0 && "−"}
      <span className="pre">$</span>
      {dollars}
      <span className="dec">
        {cents}
        {unit}
      </span>
    </div>
  );
}

export function Change({ r, a, b }: { r: Resource; a?: number; b?: number }) {
  if (a === b) return <span className="rz-dim">{a == null ? "none" : fmt(r, a)}</span>;
  return (
    <span className="rz-nowrap">
      {a == null ? "none" : fmt(r, a)} <ArrowRight size={11} className="rz-dim" />{" "}
      <b>{b == null ? "none" : fmt(r, b)}</b>
    </span>
  );
}

export function Modal({
  title,
  children,
  onClose,
  actions,
}: {
  title: string;
  children: ReactNode;
  onClose: () => void;
  actions: ReactNode;
}) {
  useEffect(() => {
    const k = (e: KeyboardEvent) => e.key === "Escape" && onClose();
    window.addEventListener("keydown", k);
    return () => window.removeEventListener("keydown", k);
  }, [onClose]);
  return (
    <div className="rz-backdrop" onClick={onClose}>
      <div className="rz-modal" role="dialog" aria-modal="true" aria-label={title} onClick={(e) => e.stopPropagation()}>
        <div className="rz-modalhd">
          <h3>{title}</h3>
          <button className="rz-icon" onClick={onClose} aria-label="Close">
            <X size={16} />
          </button>
        </div>
        <div className="rz-modalbody">{children}</div>
        <div className="rz-modalactions">{actions}</div>
      </div>
    </div>
  );
}

export function CopyButton({ text }: { text: string }) {
  const [done, setDone] = useState(false);
  return (
    <button
      className="rz-btn ghost small"
      onClick={() => {
        void navigator.clipboard?.writeText(text);
        setDone(true);
        setTimeout(() => setDone(false), 1500);
      }}
    >
      {done ? <Check size={14} /> : <Copy size={14} />} {done ? "Copied" : "Copy"}
    </button>
  );
}

export const SourceTag = ({ source }: { source: string }) => <span className={`rz-src s-${source}`}>{source}</span>;

export function ErrorBox({ error }: { error: unknown }) {
  if (!error) return null;
  return (
    <div className="rz-alert t-bad" role="alert">
      <CircleAlert size={16} />
      {error instanceof Error ? error.message : String(error)}
    </div>
  );
}

export const Logo = () => (
  <svg width="24" height="24" viewBox="0 0 24 24" aria-hidden>
    <rect x="1" y="1" width="22" height="22" rx="6" fill="var(--accent)" />
    <rect x="6" y="12.5" width="3" height="5.5" rx="1" fill="var(--bg)" />
    <rect x="10.5" y="8.5" width="3" height="9.5" rx="1" fill="var(--bg)" />
    <rect x="15" y="5.5" width="3" height="12.5" rx="1" fill="var(--bg)" />
  </svg>
);

export function Toasts() {
  const list = useToasts();
  return (
    <div className="rz-toasts" aria-live="polite">
      {list.map((t) => (
        <div key={t.id} className={`rz-toast t-${t.kind}`}>
          {t.kind === "ok" ? (
            <CircleCheck size={17} />
          ) : t.kind === "error" ? (
            <CircleAlert size={17} />
          ) : (
            <Info size={17} />
          )}
          <div className="tb">
            <b>{t.title}</b>
            {t.body && <div className="rz-dim rz-xs">{t.body}</div>}
            {t.action && (
              <button
                className="rz-link"
                onClick={() => {
                  t.action!.run();
                  dismiss(t.id);
                }}
              >
                {t.action.label}
              </button>
            )}
          </div>
          <button className="rz-icon" aria-label="Dismiss" onClick={() => dismiss(t.id)}>
            <X size={14} />
          </button>
        </div>
      ))}
    </div>
  );
}

export const WlCell = ({ w }: { w: Summary }) => (
  <a className="rz-wl" href={workloadHash(w)} onClick={(e) => e.stopPropagation()}>
    <b>{w.name}</b>
    <span className="rz-dim rz-xs">
      {w.namespace} · {w.kind}
    </span>
  </a>
);
