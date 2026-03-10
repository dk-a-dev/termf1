/* Top-down F1 car SVG – nose pointing DOWN (car approaching viewer)
   ViewBox 0 0 140 370   ·   rear at top, front at bottom */
export default function F1Car({ width = 200, glow = true }) {
  const h = Math.round(width * (370 / 140))
  return (
    <svg
      viewBox="0 0 140 370"
      width={width}
      height={h}
      xmlns="http://www.w3.org/2000/svg"
      style={
        glow
          ? {
              filter:
                'drop-shadow(0 0 28px rgba(232,0,45,0.75)) drop-shadow(0 0 8px rgba(232,0,45,0.4))',
            }
          : undefined
      }
    >
      {/* ── REAR WING (top) ─────────────────────────────── */}
      <rect x="9" y="12" width="122" height="15" rx="3.5" fill="#111" stroke="#2a2a2a" strokeWidth="0.6" />
      {/* DRS highlights */}
      <rect x="14" y="14" width="48" height="4" rx="1.5" fill="#E8B000" opacity="0.85" />
      <rect x="78" y="14" width="48" height="4" rx="1.5" fill="#E8B000" opacity="0.85" />
      {/* End plates */}
      <rect x="4"   y="7"  width="11" height="29" rx="2.5" fill="#181818" stroke="#2a2a2a" strokeWidth="0.5" />
      <rect x="125" y="7"  width="11" height="29" rx="2.5" fill="#181818" stroke="#2a2a2a" strokeWidth="0.5" />
      {/* Support struts */}
      <rect x="62" y="24" width="7"  height="38" fill="#1e1e1e" />
      <rect x="71" y="24" width="7"  height="38" fill="#1e1e1e" />

      {/* ── REAR TYRES ───────────────────────────────────── */}
      <rect x="0"   y="56" width="36" height="72" rx="9"   fill="#0e0e0e" />
      <rect x="104" y="56" width="36" height="72" rx="9"   fill="#0e0e0e" />
      {/* Rims */}
      <rect x="6"   y="63" width="24" height="58" rx="5"   fill="#1c1c1c" />
      <rect x="110" y="63" width="24" height="58" rx="5"   fill="#1c1c1c" />
      {/* Pirelli stripe */}
      <rect x="7"   y="83" width="22" height="5"  rx="1"   fill="#555" opacity="0.55" />
      <rect x="111" y="83" width="22" height="5"  rx="1"   fill="#555" opacity="0.55" />
      {/* Brake calipers – gold */}
      <rect x="13"  y="71" width="6"  height="26" rx="1.5" fill="#E8B000" />
      <rect x="121" y="71" width="6"  height="26" rx="1.5" fill="#E8B000" />
      {/* Rear pull-rods */}
      <line x1="36"  y1="74"  x2="52"  y2="74"  stroke="#3a3a3a" strokeWidth="3.5" strokeLinecap="round" />
      <line x1="36"  y1="114" x2="52"  y2="114" stroke="#3a3a3a" strokeWidth="3.5" strokeLinecap="round" />
      <line x1="104" y1="74"  x2="88"  y2="74"  stroke="#3a3a3a" strokeWidth="3.5" strokeLinecap="round" />
      <line x1="104" y1="114" x2="88"  y2="114" stroke="#3a3a3a" strokeWidth="3.5" strokeLinecap="round" />

      {/* ── MAIN BODY ────────────────────────────────────── */}
      <path
        d="M52,40 L52,335 Q52,348 70,352 Q88,348 88,335 L88,40 Z"
        fill="#C40000"
      />
      {/* Centre spine highlight */}
      <path
        d="M63,42 L63,334 L70,350 L77,334 L77,42 Z"
        fill="#E8002D"
        opacity="0.45"
      />
      {/* Hair-line glow on spine */}
      <line x1="70" y1="46" x2="70" y2="334" stroke="#FF6060" strokeWidth="0.8" opacity="0.25" />

      {/* ── SIDE PODS ────────────────────────────────────── */}
      <path d="M52,108 Q36,114 30,135 L30,238 Q30,258 44,264 L52,265 Z" fill="#A30000" />
      <path d="M88,108 Q104,114 110,135 L110,238 Q110,258 96,264 L88,265 Z" fill="#A30000" />
      {/* Air intakes */}
      <ellipse cx="39"  cy="147" rx="11" ry="21" fill="#060606" opacity="0.9" />
      <ellipse cx="101" cy="147" rx="11" ry="21" fill="#060606" opacity="0.9" />
      {/* Intake rim */}
      <path d="M32,136 Q32,126 39,126 Q46,126 46,136" stroke="#2a2a2a" strokeWidth="0.8" fill="none" />
      <path d="M94,136 Q94,126 101,126 Q108,126 108,136" stroke="#2a2a2a" strokeWidth="0.8" fill="none" />

      {/* ── ENGINE INTAKE ────────────────────────────────── */}
      <ellipse cx="70" cy="64" rx="15" ry="11" fill="#040404" opacity="0.92" />
      <path d="M58,57 Q70,52 82,57" stroke="#2a2a2a" strokeWidth="0.8" fill="none" />

      {/* ── COCKPIT ──────────────────────────────────────── */}
      <path
        d="M60,188 Q60,167 70,164 Q80,167 80,188 L80,256 Q80,268 70,270 Q60,268 60,256 Z"
        fill="#050505"
      />
      {/* Halo */}
      <path
        d="M59,202 L60,192 Q70,184 80,192 L81,202 L79,209 Q70,200 61,209 Z"
        fill="#555" stroke="#777" strokeWidth="0.7"
      />
      {/* Visor */}
      <ellipse cx="70" cy="228" rx="10" ry="13" fill="#0c1e42" />
      <path d="M63,224 Q70,216 77,224 L76,233 Q70,227 64,233 Z" fill="#1e4080" opacity="0.75" />

      {/* ── LIVERY TEXT ──────────────────────────────────── */}
      <text x="70"  y="157" textAnchor="middle" fill="white" fontSize="7.5" fontWeight="700"
            fontFamily="'JetBrains Mono',monospace">termf1</text>
      <text x="42"  y="208" textAnchor="middle" fill="white" fontSize="6.5" fontWeight="700"
            fontFamily="monospace" opacity="0.75" transform="rotate(-90,42,208)">v2</text>
      <text x="98"  y="198" textAnchor="middle" fill="white" fontSize="6.5" fontWeight="700"
            fontFamily="monospace" opacity="0.75" transform="rotate(90,98,198)">v2</text>

      {/* ── FRONT TYRES ──────────────────────────────────── */}
      <rect x="2"   y="273" width="32" height="64" rx="7.5" fill="#0e0e0e" />
      <rect x="106" y="273" width="32" height="64" rx="7.5" fill="#0e0e0e" />
      {/* Rims */}
      <rect x="8"   y="281" width="20" height="48" rx="4.5" fill="#1c1c1c" />
      <rect x="112" y="281" width="20" height="48" rx="4.5" fill="#1c1c1c" />
      {/* Front calipers */}
      <rect x="13"  y="291" width="5"  height="20" rx="1.5" fill="#E8B000" />
      <rect x="122" y="291" width="5"  height="20" rx="1.5" fill="#E8B000" />
      {/* Front push-rods */}
      <line x1="34"  y1="283" x2="52"  y2="283" stroke="#3a3a3a" strokeWidth="3"  strokeLinecap="round" />
      <line x1="34"  y1="325" x2="52"  y2="325" stroke="#3a3a3a" strokeWidth="3"  strokeLinecap="round" />
      <line x1="106" y1="283" x2="88"  y2="283" stroke="#3a3a3a" strokeWidth="3"  strokeLinecap="round" />
      <line x1="106" y1="325" x2="88"  y2="325" stroke="#3a3a3a" strokeWidth="3"  strokeLinecap="round" />

      {/* ── NOSE CONE ────────────────────────────────────── */}
      <path
        d="M57,342 Q52,354 54,363 Q59,371 70,373 Q81,371 86,363 Q88,354 83,342 Z"
        fill="#960000"
      />
      <ellipse cx="70" cy="372" rx="10" ry="4" fill="#740000" />

      {/* ── FRONT WING ───────────────────────────────────── */}
      <rect x="11" y="354" width="118" height="13" rx="3.5" fill="#111" stroke="#222" strokeWidth="0.5" />
      <rect x="14" y="347" width="112" height="8"  rx="2"   fill="#191919" />
      {/* End plates */}
      <rect x="4"   y="342" width="12" height="29" rx="2.5" fill="#141414" stroke="#2a2a2a" strokeWidth="0.5" />
      <rect x="124" y="342" width="12" height="29" rx="2.5" fill="#141414" stroke="#2a2a2a" strokeWidth="0.5" />
      {/* Wing element lines */}
      <line x1="15"  y1="357" x2="52"  y2="357" stroke="#252525" strokeWidth="0.8" />
      <line x1="88"  y1="357" x2="125" y2="357" stroke="#252525" strokeWidth="0.8" />
    </svg>
  )
}
