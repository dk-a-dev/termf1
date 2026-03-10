import { motion } from 'framer-motion'

const ITEMS = [
  {
    id: '01',
    icon: '📡',
    title: 'Live Timing Server',
    sub: 'termf1-server',
    color: '#E8002D',
    desc: 'Custom Go server that taps the official F1 live-timing SignalR feed — sector splits, telemetry (speed, throttle, brake, gear, DRS), race control messages, team radio, live gap tree, battle detection.',
    tags: ['SignalR', 'WebSocket', 'Go'],
  },
  {
    id: '02',
    icon: '🏎',
    title: 'Live Dashboard',
    sub: 'v3 major focus',
    color: '#3671C6',
    desc: 'Unified live race command centre: animated driver dots on the circuit, real-time sector splits, gap sparklines, tyre age counters, weather overlay — all in one terminal screen.',
    tags: ['Bubbletea', 'Real-time', 'TUI'],
  },
  {
    id: '03',
    icon: '📈',
    title: 'Analysis Enhancements',
    sub: 'Race Analysis v2',
    color: '#A855F7',
    desc: 'Tyre degradation model with per-stint slope, undercut/overcut detection, interactive session picker, head-to-head driver comparison, animated position replay with lap scrubber.',
    tags: ['OpenF1', 'Stats', 'Charts'],
  },
]

export default function Roadmap() {
  return (
    <section
      className="relative py-28 px-6 overflow-hidden scanlines"
      style={{ background: 'linear-gradient(180deg, #040404 0%, #000 100%)' }}
    >
      {/* Background diagonal stripe */}
      <div
        className="absolute inset-0 pointer-events-none opacity-[0.03]"
        style={{
          backgroundImage: 'repeating-linear-gradient(45deg, white 0, white 1px, transparent 0, transparent 50%)',
          backgroundSize: '20px 20px',
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
          <p className="section-heading mb-4">V3 Incoming</p>
          <h2 className="text-4xl md:text-5xl font-black text-white tracking-tight">
            What&apos;s next
          </h2>
          <p className="mt-4 text-gray-500 max-w-lg mx-auto">
            v2 is race-ready. v3 brings the live feed — real telemetry,
            animated track position, and a web dashboard for everyone.
          </p>
        </motion.div>

        {/* Items */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {ITEMS.map((item, i) => (
            <motion.div
              key={item.id}
              className="glass rounded-2xl p-8 relative overflow-hidden group"
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.1, duration: 0.5, ease: [0.16, 1, 0.3, 1] }}
            >
              {/* Faint number watermark */}
              <div
                className="absolute -right-2 -top-4 text-[7rem] font-black pointer-events-none select-none leading-none"
                style={{ color: item.color, opacity: 0.06 }}
              >
                {item.id}
              </div>

              {/* Accent top-left bar */}
              <div
                className="absolute top-0 left-0 w-1 h-full rounded-l-2xl opacity-40 group-hover:opacity-80 transition-opacity duration-300"
                style={{ background: item.color }}
              />

              <div className="flex items-start gap-4 mb-4">
                <div
                  className="w-12 h-12 rounded-xl flex items-center justify-center text-2xl flex-shrink-0"
                  style={{ background: `${item.color}20`, border: `1px solid ${item.color}35` }}
                >
                  {item.icon}
                </div>
                <div>
                  <div className="text-[9px] font-mono tracking-[0.4em] uppercase mb-1" style={{ color: item.color }}>
                    {item.sub}
                  </div>
                  <h3 className="text-lg font-bold text-white">{item.title}</h3>
                </div>
              </div>

              <p className="text-sm text-gray-500 leading-relaxed mb-5">{item.desc}</p>

              <div className="flex gap-2 flex-wrap">
                {item.tags.map((tag) => (
                  <span
                    key={tag}
                    className="text-[9px] font-mono px-2 py-1 rounded"
                    style={{ background: `${item.color}15`, color: item.color, border: `1px solid ${item.color}25` }}
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </motion.div>
          ))}
        </div>

        {/* CTA */}
        <motion.div
          className="mt-16 text-center"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.3 }}
        >
          <p className="text-gray-600 text-sm mb-4 font-mono">
            Want to contribute or follow development?
          </p>
          <a
            href="https://github.com/dk-a-dev/termf1"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-2 px-6 py-3 border border-[#E8002D]/40 text-[#E8002D]
                       rounded-lg text-sm font-medium hover:bg-[#E8002D]/10 transition-all duration-200"
          >
            ⭐ Star on GitHub
          </a>
        </motion.div>
      </div>
    </section>
  )
}
