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

  // Simple loop through the cards, like a fast radio carousel
  useEffect(() => {
    const timer = window.setInterval(() => {
      setActiveIdx((prev) => (prev + 1) % RADIO_CARDS.length)
    }, 3000)
    return () => window.clearInterval(timer)
  }, [])

  // Basic audio hook – we only have one mp3, so reuse it for all cards.
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
    <footer className="relative border-t border-white/5 bg-black py-16 px-6 overflow-hidden">
      {/* Very faint car in background */}
      <div className="absolute right-10 bottom-0 opacity-[0.04] pointer-events-none select-none">
        <F1Car width={200} glow={false} />
      </div>

      <div className="max-w-6xl mx-auto relative z-10">
        {/* ── Top row: Brand left + Links right ──────────────────── */}
        <div className="flex flex-col md:flex-row md:items-start md:justify-between gap-10">
          {/* Brand */}
          <div className="shrink-0">
            <div className="text-3xl font-black tracking-tight mb-2" style={{ fontFamily: "'Alphacorsa', 'Inter', sans-serif" }}>
              <span className="text-white">TERM</span>
              <span style={{ color: '#E8002D' }}>F1</span>
            </div>
            <p className="text-xs text-gray-600 max-w-xs leading-relaxed">
              A full-featured Formula 1 terminal UI — open source,
              built with Go + Charm, powered by OpenF1 &amp; Groq.
            </p>
          </div>

          {/* Links */}
          <div className="flex gap-12 text-sm text-gray-500">
            <div className="space-y-2">
              <p className="text-[9px] font-mono tracking-widest text-gray-700 uppercase">Project</p>
              <div className="space-y-1.5">
                <a href="https://github.com/dk-a-dev/termf1" target="_blank" rel="noopener noreferrer"
                   className="block hover:text-white transition-colors">GitHub</a>
                <a href="https://github.com/dk-a-dev/termf1/releases" target="_blank" rel="noopener noreferrer"
                   className="block hover:text-white transition-colors">Releases</a>
                <a href="https://github.com/dk-a-dev/termf1/issues" target="_blank" rel="noopener noreferrer"
                   className="block hover:text-white transition-colors">Issues</a>
              </div>
            </div>
            <div className="space-y-2">
              <p className="text-[9px] font-mono tracking-widest text-gray-700 uppercase">Data</p>
              <div className="space-y-1.5">
                <a href="https://openf1.org" target="_blank" rel="noopener noreferrer"
                   className="block hover:text-white transition-colors">OpenF1 API</a>
                <a href="https://groq.com" target="_blank" rel="noopener noreferrer"
                   className="block hover:text-white transition-colors">Groq AI</a>
                <a href="https://api.multiviewer.app" target="_blank" rel="noopener noreferrer"
                   className="block hover:text-white transition-colors">Multiviewer</a>
              </div>
            </div>
          </div>
        </div>

        {/* ── Divider ────────────────────────────────────────────── */}
        <div className="mt-12 pt-8 border-t border-white/5">
          {/* Radio section header */}
          <div className="flex flex-col md:flex-row md:items-end md:justify-between gap-4 mb-8">
            <div className="space-y-1">
              <p className="text-[9px] font-mono tracking-[0.3em] text-gray-700 uppercase">Team Radio Wall</p>
              <p className="text-[11px] text-gray-600 font-mono leading-relaxed max-w-sm">
                End the lap with a wall of iconic F1 team radios. Tap the card for a surprise.
              </p>
            </div>
            {/* Dot indicators */}
            <div className="flex gap-1.5">
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

          {/* Radio card — centred, fixed height */}
          <div className="flex justify-center">
            <div className="relative w-full max-w-md h-[460px]" onClick={toggleAudio}>
              <AnimatePresence mode="popLayout">
                <motion.div
                  key={card.id}
                  initial={{ opacity: 0, x: 60, scale: 0.96 }}
                  animate={{ opacity: 1, x: 0, scale: 1 }}
                  exit={{ opacity: 0, x: -60, scale: 0.96 }}
                  transition={{ duration: 0.45, ease: [0.16, 1, 0.3, 1] }}
                  className="absolute inset-0 rounded-2xl overflow-hidden cursor-pointer"
                  style={{
                    background: '#090A0C',
                    border: '1px solid rgba(255,255,255,0.06)',
                    boxShadow: `0 30px 60px -10px rgba(0,0,0,0.95), 0 0 80px -20px ${card.colour}15`,
                  }}
                >
                  <div className="m-[6px] h-[calc(100%-12px)] rounded-xl overflow-hidden flex flex-col" style={{ background: '#0D0E12' }}>
                    
                    {/* Header */}
                    <div className="px-6 pt-5 pb-3 flex items-start justify-between">
                      <div className="flex flex-col">
                        <span className="text-[11px] font-black tracking-[0.22em] uppercase leading-none mb-1" style={{ color: card.colour }}>
                          {card.driver.split(' ')[1] || card.driver}
                        </span>
                        <span className="text-[22px] font-black tracking-[0.18em] text-white uppercase leading-none">
                          RADIO
                        </span>
                      </div>
                      <div
                        className="w-9 h-9 rounded-lg border-2 flex items-center justify-center text-[11px] font-bold tracking-widest"
                        style={{ borderColor: `${card.colour}44`, color: card.colour }}
                      >
                        {card.number}
                      </div>
                    </div>

                    {/* Giant Number + Progress */}
                    <div className="px-6 pb-4 flex items-center gap-4">
                      <span className="text-[48px] font-black leading-none tracking-tighter" style={{ color: card.colour }}>
                        {card.number}
                      </span>
                      <div className="flex-1 h-[6px] rounded-full overflow-hidden" style={{ background: '#1A1C23' }}>
                        <div
                          className="h-full rounded-full"
                          style={{
                            width: '45%',
                            background: `linear-gradient(90deg, ${card.colour} 0%, transparent 100%)`,
                            opacity: isPlaying ? 1 : 0.35,
                            transition: 'opacity 0.3s',
                          }}
                        />
                      </div>
                    </div>

                    {/* Divider */}
                    <div className="h-px mx-4" style={{ background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.06) 50%, transparent)' }} />

                    {/* Dialogue — scrollable area */}
                    <div className="flex-1 overflow-y-auto px-6 py-5 space-y-3">
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
                              className={`px-3 py-2 rounded-lg max-w-[85%] ${isDriver ? 'rounded-tl-sm' : 'rounded-tr-sm'}`}
                              style={{
                                backgroundColor: isDriver ? `${card.colour}18` : '#1A1C23',
                                borderLeft: isDriver ? `2px solid ${card.colour}` : 'none',
                                borderRight: !isDriver ? '2px solid #4B5563' : 'none',
                              }}
                            >
                              <p
                                className="text-[12px] font-medium tracking-[0.01em] leading-snug"
                                style={{ color: isDriver ? '#fff' : '#D1D5DB', textAlign: isDriver ? 'left' : 'right' }}
                              >
                                {line.text}
                              </p>
                            </div>
                          </div>
                        )
                      })}
                    </div>

                    {/* Team name */}
                    <div className="px-6 pb-4 pt-2 border-t border-white/[0.03]">
                      <span className="text-[9px] font-mono text-gray-600 uppercase tracking-[0.25em]">
                        {card.team}
                      </span>
                    </div>

                  </div>
                </motion.div>
              </AnimatePresence>
            </div>
          </div>
        </div>
      </div>
    </footer>
  )
}
