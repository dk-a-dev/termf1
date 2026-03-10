import { motion } from 'framer-motion'

// Clean F1-specific icon components
const IconTimingTower = () => (
  <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
    <rect x="2"  y="15" width="3" height="7" opacity="0.45" />
    <rect x="7"  y="10" width="3" height="12" opacity="0.65" />
    <rect x="12" y="5"  width="3" height="17" />
    <rect x="17" y="12" width="3" height="10" opacity="0.55" />
    <rect x="2"  y="21" width="18" height="1.5" />
  </svg>
)
const IconCheckeredFlag = () => (
  <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
    <rect x="3"  y="2" width="3.5" height="3.5" />
    <rect x="10" y="2" width="3.5" height="3.5" />
    <rect x="6.5" y="5.5" width="3.5" height="3.5" />
    <rect x="13.5" y="5.5" width="3.5" height="3.5" />
    <rect x="3"  y="9" width="3.5" height="3.5" />
    <rect x="10" y="9" width="3.5" height="3.5" />
    <rect x="2.5" y="2" width="1.2" height="18" />
    <rect x="2.5" y="18.8" width="5" height="1.2" transform="rotate(-15 2.5 18.8)" />
  </svg>
)
const IconHairpin = () => (
  <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round">
    <path d="M4 21 L4 9 C4 4.5 8 4 9.5 4 C13 4 14 7 14 9 L14 15 C14 18 16 20 20 20" />
    <path d="M17 17 L20 20 L17 23" strokeWidth="1.8" />
  </svg>
)
const IconRainGauge = () => (
  <svg viewBox="0 0 24 24" width="22" height="22" fill="none">
    <path d="M4 10 C4 6 7.5 3 12 3 C16.5 3 20 6 20 10" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
    <line x1="7"  y1="13" x2="5.5" y2="17" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
    <line x1="12" y1="12" x2="10.5" y2="16" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
    <line x1="17" y1="13" x2="15.5" y2="17" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" />
    <line x1="9.5" y1="19" x2="8" y2="23" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" opacity="0.5" />
    <line x1="14.5" y1="19" x2="13" y2="23" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" opacity="0.5" />
  </svg>
)
const IconStartingGrid = () => (
  <svg viewBox="0 0 24 24" width="22" height="22" fill="currentColor">
    {/* pole position (center top) */}
    <rect x="9.5" y="2" width="5" height="3.5" rx="0.8" />
    {/* row 2 */}
    <rect x="3"   y="7" width="5" height="3.5" rx="0.8" opacity="0.8" />
    <rect x="16"  y="7" width="5" height="3.5" rx="0.8" opacity="0.8" />
    {/* row 3 */}
    <rect x="9.5" y="13" width="5" height="3.5" rx="0.8" opacity="0.6" />
    {/* row 4 */}
    <rect x="3"   y="18" width="5" height="3.5" rx="0.8" opacity="0.4" />
    <rect x="16"  y="18" width="5" height="3.5" rx="0.8" opacity="0.4" />
  </svg>
)
const IconHelmet = () => (
  <svg viewBox="0 0 24 24" width="22" height="22" fill="none">
    <path d="M5 15 C5 8.5 7.5 4 12 4 C16.5 4 19 8.5 19 15 L18 19 C17.5 20.5 6.5 20.5 6 19 Z"
      stroke="currentColor" strokeWidth="1.7" strokeLinejoin="round" />
    <path d="M7.5 15.5 L16.5 15.5 C17 15.5 17.5 15 17.5 14 L17.5 11.5 C17.5 10.5 17 10 16.5 10 L7.5 10 C7 10 6.5 10.5 6.5 11.5 L6.5 14 C6.5 15 7 15.5 7.5 15.5 Z"
      fill="currentColor" fillOpacity="0.25" stroke="currentColor" strokeWidth="1.5" />
    <path d="M8 9.5 C8 7.5 9.5 6 11 5.8" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" opacity="0.6" />
  </svg>
)
const IconTeamRadio = () => (
  <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round">
    <rect x="8" y="3" width="8" height="11" rx="4" />
    <line x1="12" y1="14" x2="12" y2="18" />
    <line x1="8.5" y1="18" x2="15.5" y2="18" />
    <path d="M4.5 8.5 C4 10.5 4.5 13 6.5 14.5" strokeDasharray="1.8 2" />
    <path d="M19.5 8.5 C20 10.5 19.5 13 17.5 14.5" strokeDasharray="1.8 2" />
    <path d="M2 6.5 C1 9.5 1.5 14 5 16" strokeDasharray="1.8 2" opacity="0.5" strokeWidth="1.4" />
    <path d="M22 6.5 C23 9.5 22.5 14 19 16" strokeDasharray="1.8 2" opacity="0.5" strokeWidth="1.4" />
  </svg>
)
const IconTyre = () => (
  <svg viewBox="0 0 24 24" width="22" height="22" fill="none">
    <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="1.8" />
    <circle cx="12" cy="12" r="5" stroke="currentColor" strokeWidth="1.5" />
    <circle cx="12" cy="12" r="1.8" fill="currentColor" />
    {/* lug bolts */}
    <circle cx="12" cy="7.5" r="1" fill="currentColor" opacity="0.7" />
    <circle cx="12" cy="16.5" r="1" fill="currentColor" opacity="0.7" />
    <circle cx="7.5" cy="12" r="1" fill="currentColor" opacity="0.7" />
    <circle cx="16.5" cy="12" r="1" fill="currentColor" opacity="0.7" />
  </svg>
)

const FEATURES = [
  {
    icon: <IconTimingTower />,
    title: 'Race Analysis',
    tag: '9 Charts',
    desc: 'Deep-dive into any completed session — strategy, pace, sectors, pit stops, positions, speed trap, and team comparison.',
    color: '#E8002D',
  },
  {
    icon: <IconCheckeredFlag />,
    title: 'Live Standings',
    tag: 'Driver & Constructor',
    desc: 'Championship standings with team-coloured proportional bar charts. Always up to date.',
    color: '#3671C6',
  },
  {
    icon: <IconHairpin />,
    title: 'Circuit Maps',
    tag: 'Real GPS Data',
    desc: 'Actual circuit outlines from GPS coordinates. Corner numbers overlaid. Speed heatmap overlay.',
    color: '#22C55E',
  },
  {
    icon: <IconRainGauge />,
    title: 'Race Weather',
    tag: 'Live Conditions',
    desc: 'Track & air temperature, humidity, wind speed/direction, rainfall — with sparkline trend charts.',
    color: '#60A5FA',
  },
  {
    icon: <IconStartingGrid />,
    title: 'Season Calendar',
    tag: 'Full Schedule',
    desc: 'Complete race calendar grouped by month. Next race auto-highlighted. Expand for session detail.',
    color: '#A855F7',
  },
  {
    icon: <IconHelmet />,
    title: 'Driver Stats',
    tag: 'Per-Driver Deep Dive',
    desc: 'Lap sparklines, best/avg/worst times, sector trends, lap histogram, pit stop history — three sub-tabs.',
    color: '#F97316',
  },
  {
    icon: <IconTeamRadio />,
    title: 'AI Chat',
    tag: 'Groq compound-beta',
    desc: 'Ask anything F1 — race strategy, regulations, history — powered by Groq with live web search.',
    color: '#14B8A6',
  },
  {
    icon: <IconTyre />,
    title: 'Pit Stop Analysis',
    tag: 'Duration Rankings',
    desc: 'Ranked pit stop chart with speed-tier colouring, per-driver stop numbers, avg marker, and retirement detection.',
    color: '#EAB308',
  },
]

const containerVariants = {
  hidden: {},
  visible: {
    transition: { staggerChildren: 0.07 },
  },
}

const cardVariants = {
  hidden:  { opacity: 0, y: 30 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.5, ease: [0.16, 1, 0.3, 1] } },
}

export default function Features() {
  return (
    <section className="relative py-28 px-6 bg-black overflow-hidden">
      {/* Subtle background grid */}
      <div
        className="absolute inset-0 pointer-events-none opacity-[0.025]"
        style={{
          backgroundImage:
            'linear-gradient(rgba(255,255,255,0.1) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.1) 1px, transparent 1px)',
          backgroundSize: '48px 48px',
        }}
      />

      <div className="relative max-w-6xl mx-auto">
        {/* Heading */}
        <motion.div
          className="mb-16 text-center"
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
        >
          <p className="section-heading mb-4">The Pit Wall</p>
          <h2 className="text-4xl md:text-5xl font-black text-white tracking-tight">
            Everything F1.{' '}
            <span style={{ color: '#E8002D' }}>One terminal.</span>
          </h2>
          <p className="mt-4 text-gray-500 max-w-lg mx-auto">
            Eight views, nine analysis charts, and an AI that knows the difference
            between a DRS train and a tyre cliff.
          </p>
        </motion.div>

        {/* Grid */}
        <motion.div
          className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4"
          variants={containerVariants}
          initial="hidden"
          whileInView="visible"
          viewport={{ once: true, margin: '-80px' }}
        >
          {FEATURES.map((f) => (
            <motion.div
              key={f.title}
              variants={cardVariants}
              className="group glass rounded-xl p-6 hover:border-white/15 transition-colors duration-300 cursor-default"
              style={{
                '--accent': f.color,
              }}
            >
              {/* icon */}
              <div
                className="w-11 h-11 rounded-lg flex items-center justify-center text-2xl mb-4"
                style={{ background: `${f.color}18`, border: `1px solid ${f.color}30` }}
              >
                {f.icon}
              </div>

              {/* tag */}
              <p
                className="text-[9px] font-mono tracking-[0.4em] uppercase mb-1.5"
                style={{ color: f.color }}
              >
                {f.tag}
              </p>

              {/* title */}
              <h3 className="text-base font-bold text-white mb-2">{f.title}</h3>

              {/* desc */}
              <p className="text-xs text-gray-500 leading-relaxed">{f.desc}</p>

              {/* Bottom accent line */}
              <div
                className="mt-5 h-px w-0 group-hover:w-full transition-all duration-500 rounded-full"
                style={{ background: f.color }}
              />
            </motion.div>
          ))}
        </motion.div>
      </div>
    </section>
  )
}
