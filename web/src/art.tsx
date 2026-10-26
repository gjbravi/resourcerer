import { useId } from "react";

/** Soft "clay" illustrations: a light platform with a glowing object on top. */
export type ArtKind = "orb" | "potion" | "shield" | "scroll" | "note";

/** A floating stack of enchanted banknotes over a glowing platform. */
export function NoteArt({ size = 132 }: { size?: number }) {
  const id = useId().replace(/:/g, "");
  const g = (n: string) => `${id}-${n}`;
  // One banknote in isometric projection; dy lifts it above the one below.
  // One banknote in isometric projection: top face A(18,44) B(62,26) C(96,40) D(52,58); dy lifts it.
  const note = (dy: number, top = false) => (
    <g transform={`translate(0 ${dy})`}>
      <path d="M18 44 L52 58 L52 61.5 L18 47.5Z" fill="var(--art-deep)" opacity=".7" />
      <path d="M52 58 L96 40 L96 43.5 L52 61.5Z" fill="var(--art-deep)" opacity=".45" />
      <path d="M18 44 L62 26 L96 40 L52 58Z" fill={`url(#${g("paper")})`} />
      <path d="M24 44 L62 28.6 L90 40 L52 55.4Z" fill="none" stroke="#fff" strokeOpacity=".6" strokeWidth=".9" />
      {top && (
        <>
          <ellipse cx="69" cy="37.5" rx="6.5" ry="3.6" fill="#fff" opacity=".9" transform="rotate(-22 69 37.5)" />
          <path d="M69 35 l.9 1.7 1.9.6 -1.9.7 -.9 1.7 -.8 -1.7 -1.9 -.7 1.9 -.6z" fill="var(--accent)" />
          <ellipse cx="45" cy="47" rx="3.2" ry="1.8" fill="#fff" opacity=".55" transform="rotate(-22 45 47)" />
        </>
      )}
    </g>
  );
  return (
    <svg width={size} height={size * 0.9} viewBox="0 0 114 102" aria-hidden className="rz-art rz-note rz-save-art">
      <defs>
        <linearGradient id={g("paper")} x1="0" y1="0" x2="1" y2="1">
          <stop offset="0" stopColor="var(--art-light)" />
          <stop offset="1" stopColor="var(--art-deep)" />
        </linearGradient>
        <linearGradient id={g("plat")} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="var(--art-plat-top)" />
          <stop offset="1" stopColor="var(--art-plat-bottom)" />
        </linearGradient>
        <radialGradient id={g("beam")} cx=".5" cy="1" r=".9">
          <stop offset="0" stopColor="var(--art-deep)" stopOpacity=".35" />
          <stop offset="1" stopColor="var(--art-deep)" stopOpacity="0" />
        </radialGradient>
        <filter id={g("glow")} x="-60%" y="-60%" width="220%" height="220%">
          <feGaussianBlur stdDeviation="6" />
        </filter>
      </defs>
      <ellipse className="rz-note-shadow" cx="57" cy="92" rx="40" ry="8" fill="var(--art-shadow)" />
      <path d="M17 82 L57 96 L97 82 L97 75 L57 89 L17 75Z" fill="var(--art-plat-bottom)" />
      <path d="M17 75 L57 61 L97 75 L57 89Z" fill={`url(#${g("plat")})`} />
      <ellipse cx="57" cy="75" rx="24" ry="8" fill="var(--art-deep)" opacity=".35" filter={`url(#${g("glow")})`} />
      <path d="M33 75 L33 30 Q57 18 81 30 L81 75 Q57 86 33 75Z" fill={`url(#${g("beam")})`} />
      <g className="rz-note-float">
        {note(10)}
        {note(3)}
        {note(-4, true)}
        {/* Paper band across the stack, parallel to the short edges. */}
        <path d="M36.5 32.2 L43.5 29.3 L77.5 43.3 L70.5 46.2Z" fill="#fff" opacity=".95" />
        <path d="M70.5 46.2 L77.5 43.3 L77.5 57.8 L70.5 60.7Z" fill="#fff" opacity=".7" />
        <path d="M36.5 32.2 L36.5 35 L70.5 49 L70.5 46.2Z" fill="var(--art-deep)" opacity=".15" />
      </g>
      <g className="rz-sparkles" fill="var(--art-light)">
        <path d="M22 22 l1.2 2.6 2.6 1.2 -2.6 1.2 -1.2 2.6 -1.2 -2.6 -2.6 -1.2 2.6 -1.2z" />
        <path d="M96 16 l.9 2 2 .9 -2 .9 -.9 2 -.9 -2 -2 -.9 2 -.9z" />
        <circle cx="90" cy="58" r="1.6" />
      </g>
    </svg>
  );
}

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
        <g className="rz-orb-float">
          <ellipse cx="50" cy="60" rx="15" ry="4.5" fill="var(--art-deep)" opacity=".55" />
          <circle cx="50" cy="38" r="21" fill={`url(#${g("orb")})`} />
          <ellipse cx="42" cy="29" rx="7" ry="4" fill="#fff" opacity=".7" transform="rotate(-30 42 29)" />
          <path d="M56 40 l1.8 4.4 4.4 1.8 -4.4 1.8 -1.8 4.4 -1.8 -4.4 -4.4 -1.8 4.4 -1.8z" fill="#fff" opacity=".9" />
        </g>
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

/** A floating magic wand casting sparkles, for the fit score. */
export function WandArt({ size = 128 }: { size?: number }) {
  const id = useId().replace(/:/g, "");
  const g = (n: string) => `${id}-${n}`;
  return (
    <svg width={size} height={size * 0.9} viewBox="0 0 114 102" aria-hidden className="rz-art">
      <defs>
        <linearGradient id={g("plat")} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="var(--art-plat-top)" />
          <stop offset="1" stopColor="var(--art-plat-bottom)" />
        </linearGradient>
        <linearGradient id={g("stick")} x1="0" y1="0" x2="1" y2="0">
          <stop offset="0" stopColor="#3d3166" />
          <stop offset=".5" stopColor="#5a4a92" />
          <stop offset="1" stopColor="#2a1f4d" />
        </linearGradient>
        <radialGradient id={g("star")} cx=".5" cy=".45" r=".6">
          <stop offset="0" stopColor="#fff" />
          <stop offset=".45" stopColor="var(--art-light)" />
          <stop offset="1" stopColor="var(--art-deep)" />
        </radialGradient>
        <filter id={g("glow")} x="-80%" y="-80%" width="260%" height="260%">
          <feGaussianBlur stdDeviation="5" />
        </filter>
      </defs>
      <ellipse className="rz-note-shadow" cx="57" cy="92" rx="40" ry="8" fill="var(--art-shadow)" />
      <path d="M17 82 L57 96 L97 82 L97 75 L57 89 L17 75Z" fill="var(--art-plat-bottom)" />
      <path d="M17 75 L57 61 L97 75 L57 89Z" fill={`url(#${g("plat")})`} />
      <ellipse cx="57" cy="75" rx="22" ry="7" fill="var(--art-deep)" opacity=".3" filter={`url(#${g("glow")})`} />
      <g className="rz-wand-float">
        <g transform="rotate(-38 57 46)">
          <rect x="53.5" y="30" width="7" height="44" rx="3.5" fill={`url(#${g("stick")})`} />
          <rect x="53.5" y="30" width="7" height="8" rx="3" fill="#fff" />
          <rect x="53.5" y="64" width="7" height="3" fill="#fff" opacity=".35" />
        </g>
        <circle cx="70" cy="29" r="13" fill="var(--art-deep)" opacity=".55" filter={`url(#${g("glow")})`} />
        <path
          d="M70 17 l3.4 7.4 8 1.2 -5.8 5.6 1.4 8 -7 -3.8 -7 3.8 1.4 -8 -5.8 -5.6 8 -1.2z"
          fill={`url(#${g("star")})`}
        />
      </g>
      <g className="rz-sparkles" fill="var(--art-light)">
        <path d="M88 14 l1.2 2.6 2.6 1.2 -2.6 1.2 -1.2 2.6 -1.2 -2.6 -2.6 -1.2 2.6 -1.2z" />
        <path d="M92 40 l.9 2 2 .9 -2 .9 -.9 2 -.9 -2 -2 -.9 2 -.9z" />
        <circle cx="50" cy="16" r="1.6" />
        <circle cx="84" cy="54" r="1.2" />
      </g>
    </svg>
  );
}
