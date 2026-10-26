import { useId } from "react";

/** Soft "clay" illustrations: a light platform with a glowing object on top. */
export type ArtKind = "orb" | "potion" | "shield" | "scroll";

export function Art({ kind, size = 104 }: { kind: ArtKind; size?: number }) {
  const id = useId().replace(/:/g, "");
  const g = (n: string) => `${id}-${n}`;
  return (
    <svg width={size} height={size * 0.92} viewBox="0 0 100 92" aria-hidden className="rz-art">
      <defs>
        <linearGradient id={g("plat")} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="var(--art-plat-top)" />
          <stop offset="1" stopColor="var(--art-plat-bottom)" />
        </linearGradient>
        <radialGradient id={g("orb")} cx=".38" cy=".32" r=".75">
          <stop offset="0" stopColor="#fff" />
          <stop offset=".35" stopColor="var(--art-light)" />
          <stop offset="1" stopColor="var(--art-deep)" />
        </radialGradient>
        <linearGradient id={g("liq")} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="var(--art-light)" />
          <stop offset="1" stopColor="var(--art-deep)" />
        </linearGradient>
        <filter id={g("glow")} x="-60%" y="-60%" width="220%" height="220%">
          <feGaussianBlur stdDeviation="7" />
        </filter>
      </defs>
      <ellipse cx="50" cy="80" rx="40" ry="10" fill="var(--art-shadow)" />
      <path d="M12 68 L50 80 L88 68 L88 61 L50 73 L12 61Z" fill="var(--art-plat-bottom)" />
      <path d="M12 61 L50 49 L88 61 L50 73Z" fill={`url(#${g("plat")})`} />
      <ellipse cx="50" cy="44" rx="24" ry="16" fill="var(--art-deep)" filter={`url(#${g("glow")})`} opacity=".45" />
      {kind === "orb" && (
        <>
          <ellipse cx="50" cy="60" rx="15" ry="4.5" fill="var(--art-deep)" opacity=".55" />
          <circle cx="50" cy="38" r="21" fill={`url(#${g("orb")})`} />
          <ellipse cx="42" cy="29" rx="7" ry="4" fill="#fff" opacity=".7" transform="rotate(-30 42 29)" />
          <path d="M56 40 l1.8 4.4 4.4 1.8 -4.4 1.8 -1.8 4.4 -1.8 -4.4 -4.4 -1.8 4.4 -1.8z" fill="#fff" opacity=".9" />
        </>
      )}
      {kind === "potion" && (
        <>
          <rect x="44" y="9" width="12" height="9" rx="2.5" fill="var(--art-cork)" />
          <path
            d="M45.5 17 h9 v9 Q71 32 70 47 Q69 63 50 63 Q31 63 30 47 Q29 32 45.5 26z"
            fill="#ffffffd9"
            stroke="#fff"
            strokeWidth="1.2"
          />
          <path d="M31.5 44 Q50 38 68.5 44 Q68 61 50 61 Q32 61 31.5 44z" fill={`url(#${g("liq")})`} />
          <circle cx="45" cy="50" r="2.4" fill="#fff" opacity=".75" />
          <circle cx="55" cy="55" r="1.6" fill="#fff" opacity=".6" />
          <path
            d="M36 31 Q39 27 44 26.5"
            stroke="#fff"
            strokeWidth="2.6"
            fill="none"
            opacity=".85"
            strokeLinecap="round"
          />
        </>
      )}
      {kind === "shield" && (
        <>
          <path d="M50 14 L68 21 V37 Q68 52 50 60 Q32 52 32 37 V21Z" fill={`url(#${g("liq")})`} />
          <path d="M50 14 L68 21 V37 Q68 52 50 60Z" fill="#000" opacity=".08" />
          <path
            d="M42 37 l6 6 11-12"
            stroke="#fff"
            strokeWidth="3.4"
            fill="none"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </>
      )}
      {kind === "scroll" && (
        <>
          <rect x="33" y="20" width="34" height="38" rx="4" fill="#fff" />
          <rect x="29" y="16" width="42" height="8" rx="4" fill={`url(#${g("liq")})`} />
          <rect x="29" y="54" width="42" height="8" rx="4" fill={`url(#${g("liq")})`} />
          <g stroke="var(--art-deep)" strokeWidth="2" strokeLinecap="round" opacity=".55">
            <line x1="38" x2="62" y1="32" y2="32" />
            <line x1="38" x2="56" y1="39" y2="39" />
            <line x1="38" x2="60" y1="46" y2="46" />
          </g>
        </>
      )}
    </svg>
  );
}

export const HatLogo = ({ size = 26 }: { size?: number }) => (
  <svg width={size} height={size} viewBox="0 0 26 26" aria-hidden>
    <rect width="26" height="26" rx="8" fill="var(--accent)" />
    <path d="M5.5 19.5 h15 l-2.6-1.8 L13 5 8.1 17.7z" fill="#fff" />
    <path d="M12 11.2l.8 1.7 1.7.5-1.7.6-.8 1.7-.6-1.7-1.7-.6 1.7-.5z" fill="var(--accent)" />
  </svg>
);

/** Half-circle gauge for the fit score. */
export function Gauge({ value, size = 180 }: { value?: number; size?: number }) {
  const v = value == null ? 0 : Math.max(0, Math.min(100, value));
  const r = 76;
  const len = Math.PI * r;
  const tone = value == null ? "var(--faint)" : v >= 80 ? "var(--accent)" : v >= 50 ? "var(--warn)" : "var(--bad)";
  return (
    <svg
      viewBox="0 0 200 104"
      width={size}
      height={size * 0.52}
      role="img"
      aria-label={value == null ? "no fit score" : `fit score ${Math.round(v)}`}
    >
      <path d="M24 92 A76 76 0 0 1 176 92" stroke="var(--track)" strokeWidth="14" fill="none" strokeLinecap="round" />
      {v > 0 && (
        <path
          d="M24 92 A76 76 0 0 1 176 92"
          stroke={tone}
          strokeWidth="14"
          fill="none"
          strokeLinecap="round"
          strokeDasharray={`${(v / 100) * len} ${len}`}
        />
      )}
      <text x="100" y="86" textAnchor="middle" className="rz-gaugev">
        {value == null ? "–" : Math.round(v)}
      </text>
    </svg>
  );
}
