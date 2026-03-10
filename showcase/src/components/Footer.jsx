import { useEffect, useRef, useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import F1Car from './F1Car'

const RADIO_CARDS = [
  {
    id: 'leclerc-water',
    number: '16',
    team: 'SCUDERIA FERRARI',
    driver: 'Charles Leclerc',
    shortLabel: 'LECLERC RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'IS THERE A LEAKAGE?' },
      { role: 'ENGINEER', text: 'A LEAKAGE OF WHAT?' },
      { role: 'DRIVER', text: 'I HAVE THE SEAT FULL OF WATER! LIKE, FULL OF WATER!' },
      { role: 'DRIVER', text: 'MUST BE THE WATER' },
      { role: 'ENGINEER', text: 'LET\'S ADD THAT TO THE WORDS OF WISDOM.' },
    ],
    colour: '#DC0000',
  },
  {
    id: 'sainz-tyres',
    number: '55',
    team: 'SCUDERIA FERRARI',
    driver: 'Carlos Sainz',
    shortLabel: 'SAINZ RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'TYRES ARE TOTALLY DEAD.' },
      { role: 'ENGINEER', text: 'YOU JUST SET PURPLE SECTOR 2.' },
      { role: 'DRIVER', text: 'I\'M TELLING YOU, I NEED TO PIT.' },
      { role: 'ENGINEER', text: 'THEN WHY ARE YOU FASTER THAN EVERYONE ELSE?' },
    ],
    colour: '#DC0000',
  },
  {
    id: 'hamilton-back',
    number: '44',
    team: 'MERCEDES-AMG PETRONAS',
    driver: 'Lewis Hamilton',
    shortLabel: 'HAMILTON RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'ARHH MY BACK IS KILLING ME, GUYS.' },
      { role: 'ENGINEER', text: 'YEAH COPY LEWIS, LET\'S BRING IT HOME.' },
    ],
    colour: '#00A19B',
  },
  {
    id: 'verstappen-boring',
    number: '1',
    team: 'ORACLE RED BULL RACING',
    driver: 'Max Verstappen',
    shortLabel: 'VERSTAPPEN RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'THIS IS SO BORING, SHOULD HAVE BROUGHT MY PILLOW.' },
    ],
    colour: '#1E5BC6',
  },
]

export default function Footer() {
  const audioRef = useRef(null)
  const [activeIdx, setActiveIdx] = useState(0)
  const [isPlaying, setIsPlaying] = useState(false)

  useEffect(() => {
    const timer = window.setInterval(() => {
      setActiveIdx((prev) => (prev + 1) % RADIO_CARDS.length)
    }, 3500)
    return () => window.clearInterval(timer)
  }, [])

  useEffect(() => {
    if (!audioRef.current) {
      audioRef.current = new Audio('/audios/driver-audio.mp3')
      audioRef.current.volume = 0.8
    }
  }, [])

  const toggleAudio = () => {
    const audio = audioRef.current
    if (!audio) return
    if (audio.paused) {
      audio.currentTime = 0
      audio.play().then(() => setIsPlaying(true)).catch(() => {})
    } else {
      audio.pause()
      setIsPlaying(false)
    }
  }

  const card = RADIO_CARDS[activeIdx]

  return (
    <footer className="relative bg-black overflow-hidden">
      {/* ═══════════════════════════════════════════════════════════════════
          STEERING-WHEEL COCKPIT FOOTER
          Left: brand + project info in a display panel
          Right: radio card embedded like a cockpit screen
       ═══════════════════════════════════════════════════════════════════ */}

      {/* Subtle ghost car background */}
      <div className="absolute right-0 bottom-0 opacity-[0.03] pointer-events-none select-none translate-x-1/4 translate-y-1/4">
        <F1Car width={300} glow={false} />
      </div>

      {/* Top edge — dash line of LEDs like a steering wheel */}
      <div className="w-full flex justify-center py-3 border-b border-white/[0.04]">
        <div className="flex gap-[6px]">
          {Array.from({ length: 15 }).map((_, i) => {
            const isRed    = i < 5
            const isYellow = i >= 5 && i < 10
            const isGreen  = i >= 10
            const color = isRed ? '#E8002D' : isYellow ? '#FFB800' : '#00C853'
            return (
              <div
                key={i}
                className="w-[6px] h-[6px] rounded-full"
                style={{
                  background: color,
                  opacity: 0.6,
                  boxShadow: `0 0 6px ${color}44`,
                }}
              />
            )
          })}
        </div>
      </div>

      <div className="max-w-7xl mx-auto relative z-10 px-6 py-10 md:py-14">
        {/* ── Main cockpit layout: dashboard left, radio screen right ── */}
        <div className="grid grid-cols-1 lg:grid-cols-[1fr_380px] gap-10 lg:gap-14 items-start">

          {/* ═══ LEFT: Dashboard panel ═══ */}
          <div className="space-y-8">
            {/* Brand + description */}
            <div className="flex items-start gap-6">
              <div className="shrink-0">
                <div
                  className="text-4xl font-black tracking-tight"
                  style={{ fontFamily: "'Alphacorsa', 'Inter', sans-serif" }}
                >
                  <span className="text-white">TERM</span>
                  <span style={{ color: '#E8002D' }}>F1</span>
                </div>
              </div>
              <div className="pt-1">
                <p className="text-[11px] text-gray-500 leading-relaxed max-w-xs">
                  A full-featured Formula 1 terminal UI — open source,
                  built with Go + Charm, powered by OpenF1 &amp; Groq.
                </p>
              </div>
            </div>

            {/* Link columns in a "display panel" */}
            <div
              className="rounded-xl p-5 grid grid-cols-2 sm:grid-cols-3 gap-6"
              style={{
                background: '#0A0B0D',
                border: '1px solid rgba(255,255,255,0.04)',
              }}
            >
              <div className="space-y-2.5">
                <p className="text-[9px] font-mono tracking-[0.3em] text-gray-600 uppercase">Project</p>
                <div className="space-y-1.5 text-[13px] text-gray-400">
                  <a href="https://github.com/dk-a-dev/termf1" target="_blank" rel="noopener noreferrer"
                     className="block hover:text-white transition-colors">GitHub</a>
                  <a href="https://github.com/dk-a-dev/termf1/releases" target="_blank" rel="noopener noreferrer"
                     className="block hover:text-white transition-colors">Releases</a>
                  <a href="https://github.com/dk-a-dev/termf1/issues" target="_blank" rel="noopener noreferrer"
                     className="block hover:text-white transition-colors">Issues</a>
                </div>
              </div>
              <div className="space-y-2.5">
                <p className="text-[9px] font-mono tracking-[0.3em] text-gray-600 uppercase">Data</p>
                <div className="space-y-1.5 text-[13px] text-gray-400">
                  <a href="https://openf1.org" target="_blank" rel="noopener noreferrer"
                     className="block hover:text-white transition-colors">OpenF1 API</a>
                  <a href="https://groq.com" target="_blank" rel="noopener noreferrer"
                     className="block hover:text-white transition-colors">Groq AI</a>
                  <a href="https://api.multiviewer.app" target="_blank" rel="noopener noreferrer"
                     className="block hover:text-white transition-colors">Multiviewer</a>
                </div>
              </div>
              <div className="space-y-2.5 col-span-2 sm:col-span-1">
                <p className="text-[9px] font-mono tracking-[0.3em] text-gray-600 uppercase">Built With</p>
                <div className="space-y-1.5 text-[13px] text-gray-400">
                  <span className="block">Go + Charm</span>
                  <span className="block">BubbleTea TUI</span>
                  <span className="block">Groq LLM</span>
                </div>
              </div>
            </div>

            {/* Radio intro text */}
            <div className="space-y-1">
              <p className="text-[9px] font-mono tracking-[0.3em] text-gray-600 uppercase">Team Radio Wall</p>
              <p className="text-[11px] text-gray-500 font-mono leading-relaxed">
                Iconic F1 team radios. Tap the card to play audio.
              </p>
            </div>
          </div>

          {/* ═══ RIGHT: Radio card — cockpit screen style ═══ */}
          <div className="flex flex-col items-center">
            {/* Screen bezel */}
            <div
              className="w-full rounded-2xl p-[3px]"
              style={{
                background: 'linear-gradient(135deg, #1A1A1A 0%, #0A0A0A 50%, #1A1A1A 100%)',
                boxShadow: '0 0 0 1px rgba(255,255,255,0.04), 0 20px 50px -12px rgba(0,0,0,0.8)',
              }}
            >
              <div className="relative h-[440px] rounded-[13px] overflow-hidden" onClick={toggleAudio}>
                <AnimatePresence mode="popLayout">
                  <motion.div
                    key={card.id}
                    initial={{ opacity: 0, y: 30, scale: 0.97 }}
                    animate={{ opacity: 1, y: 0, scale: 1 }}
                    exit={{ opacity: 0, y: -30, scale: 0.97 }}
                    transition={{ duration: 0.45, ease: [0.16, 1, 0.3, 1] }}
                    className="absolute inset-0 flex flex-col cursor-pointer"
                    style={{ background: '#0D0E12' }}
                  >
                    {/* Header */}
                    <div className="px-5 pt-5 pb-3 flex items-start justify-between shrink-0">
                      <div>
                        <span
                          className="text-[11px] font-black tracking-[0.22em] uppercase leading-none block mb-1"
                          style={{ color: card.colour }}
                        >
                          {card.driver.split(' ')[1] || card.driver}
                        </span>
                        <span className="text-[20px] font-black tracking-[0.18em] text-white uppercase leading-none block">
                          RADIO
                        </span>
                      </div>
                      <div
                        className="w-9 h-9 rounded-lg border-2 flex items-center justify-center text-[11px] font-bold"
                        style={{ borderColor: `${card.colour}44`, color: card.colour }}
                      >
                        {card.number}
                      </div>
                    </div>

                    {/* Number + audio bar */}
                    <div className="px-5 pb-3 flex items-center gap-3 shrink-0">
                      <span
                        className="text-[42px] font-black leading-none tracking-tighter"
                        style={{ color: card.colour }}
                      >
                        {card.number}
                      </span>
                      <div className="flex-1 h-[5px] rounded-full overflow-hidden" style={{ background: '#1A1C23' }}>
                        <motion.div
                          className="h-full rounded-full"
                          animate={{ width: isPlaying ? '75%' : '35%' }}
                          transition={{ duration: 0.6 }}
                          style={{
                            background: `linear-gradient(90deg, ${card.colour}, transparent)`,
                            opacity: isPlaying ? 1 : 0.3,
                          }}
                        />
                      </div>
                    </div>

                    {/* Separator */}
                    <div className="h-px mx-4 shrink-0" style={{ background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.06) 50%, transparent)' }} />

                    {/* Dialogue — scrollable */}
                    <div className="flex-1 overflow-y-auto px-5 py-4 space-y-3">
                      {card.dialogue.map((line, idx) => {
                        const isDriver = line.role === 'DRIVER'
                        return (
                          <div key={idx} className={`flex flex-col ${isDriver ? 'items-start' : 'items-end'}`}>
                            <span
                              className="text-[9px] font-mono tracking-[0.2em] uppercase mb-1"
                              style={{ color: isDriver ? card.colour : '#6B7280' }}
                            >
                              {line.role}
                            </span>
                            <div
                              className={`px-3 py-2 rounded-lg max-w-[88%] ${isDriver ? 'rounded-tl-sm' : 'rounded-tr-sm'}`}
                              style={{
                                backgroundColor: isDriver ? `${card.colour}18` : '#1A1C23',
                                borderLeft: isDriver ? `2px solid ${card.colour}` : 'none',
                                borderRight: !isDriver ? '2px solid #4B5563' : 'none',
                              }}
                            >
                              <p
                                className="text-[11px] font-medium leading-snug"
                                style={{ color: isDriver ? '#fff' : '#D1D5DB', textAlign: isDriver ? 'left' : 'right' }}
                              >
                                {line.text}
                              </p>
                            </div>
                          </div>
                        )
                      })}
                    </div>

                    {/* Footer: team name */}
                    <div className="px-5 pb-4 pt-2 border-t border-white/[0.03] shrink-0">
                      <span className="text-[9px] font-mono text-gray-600 uppercase tracking-[0.25em]">
                        {card.team}
                      </span>
                    </div>
                  </motion.div>
                </AnimatePresence>
              </div>
            </div>

            {/* Dot indicators */}
            <div className="flex gap-2 mt-4">
              {RADIO_CARDS.map((c, i) => (
                <button
                  key={c.id}
                  onClick={() => setActiveIdx(i)}
                  className="w-2 h-2 rounded-full transition-all duration-300"
                  style={{
                    background: i === activeIdx ? card.colour : '#2A2A2A',
                    boxShadow: i === activeIdx ? `0 0 8px ${card.colour}55` : 'none',
                  }}
                />
              ))}
            </div>
          </div>

        </div>
      </div>

      {/* Bottom copyright bar */}
      <div className="border-t border-white/[0.04] py-4 px-6">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <span className="text-[10px] font-mono text-gray-700 tracking-widest">
            &copy; {new Date().getFullYear()} TERMF1
          </span>
          <span className="text-[10px] font-mono text-gray-700 tracking-widest">
            OPEN SOURCE &middot; MIT LICENSE
          </span>
        </div>
      </div>
    </footer>
  )
}
