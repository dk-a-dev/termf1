import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'

const RAW = 'https://raw.githubusercontent.com/dk-a-dev/termf1/master/screenshots/'

const CORE = [
  { src: 'dashboard.png',   title: 'Dashboard',     desc: 'Live timing command centre (v3 rebuild in progress)' },
  { src: 'standings.png',   title: 'Standings',     desc: 'Driver & Constructor standings with team bar charts' },
  { src: 'schedule.png',    title: 'Schedule',      desc: 'Full season calendar with session detail' },
  { src: 'weather.png',     title: 'Weather',       desc: 'Real-time track & air conditions + sparklines' },
  { src: 'driver-stats.png',title: 'Driver Stats',  desc: 'Lap analysis, sector trends, pit stop counts' },
  { src: 'trackmap-1.png',  title: 'Track Map',     desc: 'Real GPS circuit + speed heatmap overlay' },
]

const ANALYSIS = [
  { src: 'graph-tyre-strategy.png',           title: '① Strategy',    desc: 'Tyre stint timeline with compound colours' },
  { src: 'graph-sparklines.png',              title: '② Sparklines',  desc: 'Per-driver lap-time sparklines' },
  { src: 'graph-race-pace.png',               title: '③ Pace',        desc: 'Pace distribution per driver' },
  { src: 'graph-sector-pace.png',             title: '④ Sectors',     desc: 'Sector breakdown across all laps' },
  { src: 'graph-speed-traps.png',             title: '⑤ Speed Trap',  desc: 'Top speed by driver' },
  { src: 'graph-track-positions-by-laps.png', title: '⑥ Positions',   desc: 'Race trajectory lap-by-lap' },
  { src: 'graph-team-pace.png',               title: '⑦ Team Pace',   desc: 'Constructor-level pace comparison' },
  { src: 'graph-pitstops.png',                title: '⑧ Pit Stops',   desc: 'Ranked stop durations with tier colouring' },
  { src: 'all-heatmap.png',                   title: 'Speed Heatmap', desc: 'Full-field speed trace overlaid on circuit' },
]

function TermCard({ src, title, desc }) {
  const [hover, setHover] = useState(false)
  const [loaded, setLoaded] = useState(false)

  return (
    <motion.div
      className="term-chrome group"
      onHoverStart={() => setHover(true)}
      onHoverEnd={() => setHover(false)}
      animate={hover ? { scale: 1.025, y: -4 } : { scale: 1, y: 0 }}
      transition={{ type: 'spring', stiffness: 280, damping: 22 }}
      style={hover ? { boxShadow: '0 24px 80px rgba(0,0,0,0.9), 0 0 0 1px rgba(232,0,45,0.25), 0 0 40px rgba(232,0,45,0.12)' } : {}}
    >
      {/* Title bar */}
      <div className="term-titlebar">
        <span className="term-dot" style={{ background: '#FF5F57' }} />
        <span className="term-dot" style={{ background: '#FEBC2E' }} />
        <span className="term-dot" style={{ background: '#28C840' }} />
        <span className="ml-3 text-[10px] font-mono text-gray-500 tracking-wide truncate">
          termf1 — {title}
        </span>
      </div>

      {/* Screenshot */}
      <div className="relative overflow-hidden" style={{ aspectRatio: '16/9', background: '#0a0a0a' }}>
        {!loaded && (
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="w-6 h-6 border-2 border-[#E8002D] border-t-transparent rounded-full animate-spin" />
          </div>
        )}
        <img
          src={`${RAW}${src}`}
          alt={title}
          className={`w-full h-full object-cover transition-transform duration-500 ${hover ? 'scale-105' : 'scale-100'} ${loaded ? 'opacity-100' : 'opacity-0'}`}
          onLoad={() => setLoaded(true)}
        />
        {/* Hover overlay */}
        <div
          className="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-300"
          style={{ background: 'linear-gradient(to top, rgba(0,0,0,0.7) 0%, transparent 50%)' }}
        />
      </div>

      {/* Caption */}
      <div className="px-4 py-3 border-t border-white/5">
        <p className="text-xs font-mono text-[#E8002D] tracking-wide">{title}</p>
        <p className="text-[11px] text-gray-500 mt-0.5 leading-snug">{desc}</p>
      </div>
    </motion.div>
  )
}

const tabs = [
  { id: 'core',     label: 'Core Views',    count: CORE.length },
  { id: 'analysis', label: 'Race Analysis', count: ANALYSIS.length },
]

export default function Screenshots() {
  const [active, setActive] = useState('core')
  const items = active === 'core' ? CORE : ANALYSIS

  return (
    <section id="screenshots" className="relative py-28 px-6" style={{ background: '#040404' }}>
      <div className="max-w-7xl mx-auto">
        {/* Heading */}
        <motion.div
          className="mb-12 text-center"
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
        >
          <p className="section-heading mb-4">Live Feed</p>
          <h2 className="text-4xl md:text-5xl font-black text-white tracking-tight">
            See it in action
          </h2>
          <p className="mt-4 text-gray-500 max-w-md mx-auto">
            Every chart, every view — captured from the 2026 Australian GP.
          </p>
        </motion.div>

        {/* Tabs */}
        <motion.div
          className="flex justify-center mb-10 gap-2"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.4, delay: 0.2 }}
        >
          {tabs.map((t) => (
            <button
              key={t.id}
              onClick={() => setActive(t.id)}
              className={`relative px-6 py-2.5 rounded text-sm font-medium transition-all duration-200
                ${active === t.id ? 'text-white' : 'text-gray-500 hover:text-gray-300'}`}
            >
              {active === t.id && (
                <motion.div
                  layoutId="tab-bg"
                  className="absolute inset-0 rounded"
                  style={{ background: '#E8002D' }}
                  transition={{ type: 'spring', stiffness: 350, damping: 30 }}
                />
              )}
              <span className="relative z-10">
                {t.label}
                <span className={`ml-2 text-[10px] ${active === t.id ? 'text-red-200' : 'text-gray-600'}`}>
                  {t.count}
                </span>
              </span>
            </button>
          ))}
        </motion.div>

        {/* Grid */}
        <AnimatePresence mode="wait">
          <motion.div
            key={active}
            className={`grid gap-4 ${active === 'core' ? 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3' : 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3'}`}
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -12 }}
            transition={{ duration: 0.3 }}
          >
            {items.map((item, i) => (
              <motion.div
                key={item.src}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.06, duration: 0.4 }}
              >
                <TermCard {...item} />
              </motion.div>
            ))}
          </motion.div>
        </AnimatePresence>

        {/* Bottom strip marquee */}
        <div className="mt-16 overflow-hidden border-t border-b border-white/5 py-4">
          <div
            className="whitespace-nowrap animate-marquee inline-block text-[11px] font-mono text-gray-700 tracking-widest"
          >
            {Array(4).fill(
              'STRATEGY · SPARKLINES · PACE · SECTORS · SPEED TRAP · POSITIONS · TEAM PACE · PIT STOPS · STANDINGS · TRACK MAP · WEATHER · DRIVER STATS · AI CHAT · SCHEDULE · '
            ).join('')}
          </div>
        </div>
      </div>
    </section>
  )
}
