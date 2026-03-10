import { useEffect, useRef, useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import F1Car from './F1Car'

const RADIO_CARDS = [
  {
    id: 'raikkonen-leave-me-alone',
    number: '7',
    team: 'LOTUS F1 TEAM',
    driver: 'Kimi Räikkönen',
    shortLabel: 'RAIKKONEN RADIO',
    dialogue: [
      { role: 'ENGINEER', text: 'KIMI YOU NEED TO PUSH NOW.' },
      { role: 'DRIVER', text: 'LEAVE ME ALONE, I KNOW WHAT I AM DOING.' },
    ],
    colour: '#FFD800',
  },
  {
    id: 'raikkonen-drink',
    number: '7',
    team: 'LOTUS F1 TEAM',
    driver: 'Kimi Räikkönen',
    shortLabel: 'RAIKKONEN RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'GIVE ME THE DRINK.' },
      { role: 'ENGINEER', text: 'NO KIMI, YOU WILL NOT HAVE THE DRINK.' },
    ],
    colour: '#FFD800',
  },
  {
    id: 'raikkonen-steering-wheel',
    number: '7',
    team: 'FERRARI',
    driver: 'Kimi Räikkönen',
    shortLabel: 'RAIKKONEN RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'STEERING WHEEL.' },
      { role: 'DRIVER', text: 'GLOVES AND STEERING WHEEL.' },
      { role: 'DRIVER', text: 'HEY! STEERING WHEEL SOMEBODY TELL HIM TO GIVE IT TO ME.' },
    ],
    colour: '#DC0000',
  },
  {
    id: 'raikkonen-for-what',
    number: '7',
    team: 'LOTUS F1 TEAM',
    driver: 'Kimi Räikkönen',
    shortLabel: 'RAIKKONEN RADIO',
    dialogue: [
      { role: 'ENGINEER', text: 'KIMI YOU WILL HAVE A PENALTY.' },
      { role: 'DRIVER', text: 'FOR WHAT?' },
    ],
    colour: '#FFD800',
  },
  {
    id: 'alonso-gp2',
    number: '14',
    team: 'MCLAREN HONDA',
    driver: 'Fernando Alonso',
    shortLabel: 'ALONSO RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'GP2 ENGINE! GP2!' },
      { role: 'DRIVER', text: 'AARGH!' },
    ],
    colour: '#FF8700',
  },
  {
    id: 'alonso-engine-slow',
    number: '14',
    team: 'MCLAREN HONDA',
    driver: 'Fernando Alonso',
    shortLabel: 'ALONSO RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'THE ENGINE FEELS GOOD.' },
      { role: 'DRIVER', text: 'MUCH SLOWER THAN BEFORE.' },
      { role: 'DRIVER', text: 'AMAZING.' },
    ],
    colour: '#FF8700',
  },
  {
    id: 'alonso-bye-bye',
    number: '14',
    team: 'ALPINE F1 TEAM',
    driver: 'Fernando Alonso',
    shortLabel: 'ALONSO RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'YES! BYE BYE!' },
    ],
    colour: '#0090FF',
  },
  {
    id: 'massa-faster-than-you',
    number: '7',
    team: 'SCUDERIA FERRARI',
    driver: 'Felipe Massa',
    shortLabel: 'MASSA RADIO',
    dialogue: [
      { role: 'ENGINEER', text: 'FERNANDO IS FASTER THAN YOU.' },
      { role: 'ENGINEER', text: 'PLEASE CONFIRM YOU UNDERSTAND THIS MESSAGE.' },
    ],
    colour: '#DC0000',
  },
  {
    id: 'vettel-space',
    number: '5',
    team: 'FERRARI',
    driver: 'Sebastian Vettel',
    shortLabel: 'VETTEL RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'ALL THE TIME YOU HAVE TO LEAVE A SPACE!' },
    ],
    colour: '#DC0000',
  },
  {
    id: 'vettel-loose',
    number: '5',
    team: 'RED BULL RACING',
    driver: 'Sebastian Vettel',
    shortLabel: 'VETTEL RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'THERE IS SOMETHING LOOSE BETWEEN MY LEGS.' },
      { role: 'DRIVER', text: 'APART FROM THE OBVIOUS.' },
    ],
    colour: '#1E5BC6',
  },
  {
    id: 'vettel-karma',
    number: '5',
    team: 'FERRARI',
    driver: 'Sebastian Vettel',
    shortLabel: 'VETTEL RADIO',
    dialogue: [
      { role: 'ENGINEER', text: 'PALMER HAS RETIRED.' },
      { role: 'DRIVER', text: 'KARMA.' },
    ],
    colour: '#DC0000',
  },
  {
    id: 'vettel-world-champion',
    number: '1',
    team: 'RED BULL RACING',
    driver: 'Sebastian Vettel',
    shortLabel: 'VETTEL RADIO',
    dialogue: [
      { role: 'ENGINEER', text: 'SEBASTIAN VETTEL YOU ARE THE WORLD CHAMPION.' },
      { role: 'DRIVER', text: 'YES! YES! YES!' },
    ],
    colour: '#1E5BC6',
  },
  {
    id: 'vettel-love-team',
    number: '5',
    team: 'RED BULL RACING',
    driver: 'Sebastian Vettel',
    shortLabel: 'VETTEL RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'WE HAVE TO REMEMBER THESE DAYS.' },
      { role: 'DRIVER', text: 'THERE IS NO GUARANTEE THEY WILL LAST FOREVER.' },
      { role: 'DRIVER', text: 'I LOVE YOU GUYS.' },
    ],
    colour: '#1E5BC6',
  },
  {
    id: 'leclerc-im-stupid',
    number: '16',
    team: 'FERRARI',
    driver: 'Charles Leclerc',
    shortLabel: 'LECLERC RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'NOOOOOOO!' },
      { role: 'DRIVER', text: 'I AM STUPID.' },
    ],
    colour: '#DC0000',
  },
  {
    id: 'leclerc-cat',
    number: '16',
    team: 'FERRARI',
    driver: 'Charles Leclerc',
    shortLabel: 'LECLERC RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'THAT IS A CUT.' },
      { role: 'ENGINEER', text: 'WHAT?' },
      { role: 'DRIVER', text: 'A CAT... I MEANT A CAT.' },
    ],
    colour: '#DC0000',
  },
  {
    id: 'hamilton-tyres-gone',
    number: '44',
    team: 'MERCEDES AMG PETRONAS',
    driver: 'Lewis Hamilton',
    shortLabel: 'HAMILTON RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'BONO MY TYRES ARE GONE.' },
    ],
    colour: '#00A19B',
  },
  {
    id: 'hamilton-great-driver',
    number: '44',
    team: 'MERCEDES AMG PETRONAS',
    driver: 'Lewis Hamilton',
    shortLabel: 'HAMILTON RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'SUCH A GREAT DRIVER LANDO.' },
    ],
    colour: '#00A19B',
  },
  {
    id: 'hamilton-dangerous-driving',
    number: '44',
    team: 'MERCEDES AMG PETRONAS',
    driver: 'Lewis Hamilton',
    shortLabel: 'HAMILTON RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'THAT IS SOME DANGEROUS DRIVING MAN.' },
    ],
    colour: '#00A19B',
  },
  {
    id: 'verstappen-fastest-lap',
    number: '1',
    team: 'RED BULL RACING',
    driver: 'Max Verstappen',
    shortLabel: 'VERSTAPPEN RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'WHAT IS THE FASTEST LAP?' },
      { role: 'ENGINEER', text: 'WE ARE NOT CONCERNED ABOUT THAT.' },
      { role: 'DRIVER', text: 'YEAH BUT I AM.' },
    ],
    colour: '#1E5BC6',
  },
  {
    id: 'verstappen-simply-lovely',
    number: '1',
    team: 'RED BULL RACING',
    driver: 'Max Verstappen',
    shortLabel: 'VERSTAPPEN RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'SIMPLY LOVELY.' },
    ],
    colour: '#1E5BC6',
  },
  {
    id: 'verstappen-lizard',
    number: '33',
    team: 'RED BULL RACING',
    driver: 'Max Verstappen',
    shortLabel: 'VERSTAPPEN RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'THERE IS A GIANT LIZARD ON THE TRACK.' },
    ],
    colour: '#1E5BC6',
  },
  {
    id: 'gasly-monza-win',
    number: '10',
    team: 'ALPHATAURI',
    driver: 'Pierre Gasly',
    shortLabel: 'GASLY RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'WHAT DID WE JUST DO?' },
      { role: 'ENGINEER', text: 'YOU WON THE RACE.' },
      { role: 'DRIVER', text: 'OH MY GOD!' },
    ],
    colour: '#2B4562',
  },
  {
    id: 'stroll-ok-button',
    number: '18',
    team: 'ASTON MARTIN',
    driver: 'Lance Stroll',
    shortLabel: 'STROLL RADIO',
    dialogue: [
      { role: 'ENGINEER', text: 'PRESS THE OK BUTTON LANCE.' },
      { role: 'DRIVER', text: 'I PRESSED IT.' },
      { role: 'ENGINEER', text: 'YOU ARE PRESSING PIT CONFIRM.' },
      { role: 'DRIVER', text: 'PIT CONFIRM IS THE OK BUTTON BRAD.' },
    ],
    colour: '#006F62',
  },
  {
    id: 'norris-talent',
    number: '4',
    team: 'MCLAREN',
    driver: 'Lando Norris',
    shortLabel: 'NORRIS RADIO',
    dialogue: [
      { role: 'ENGINEER', text: 'LANDO WHAT DAMAGE DO YOU HAVE?' },
      { role: 'DRIVER', text: 'TALENT.' },
    ],
    colour: '#FF8700',
  },
  {
    id: 'grosjean-ericsson',
    number: '8',
    team: 'HAAS F1 TEAM',
    driver: 'Romain Grosjean',
    shortLabel: 'GROSJEAN RADIO',
    dialogue: [
      { role: 'ENGINEER', text: 'WHAT HAPPENED?' },
      { role: 'DRIVER', text: 'I THINK ERICSSON HIT US.' },
    ],
    colour: '#B6BABD',
  },
  {
    id: 'webber-vomiting',
    number: '2',
    team: 'RED BULL RACING',
    driver: 'Mark Webber',
    shortLabel: 'WEBBER RADIO',
    dialogue: [
      { role: 'DRIVER', text: 'MATE I AM VOMITING.' },
      { role: 'DRIVER', text: 'THERE IS SOME VOMITING GOING ON.' },
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

                    {/* Dialogue — F1 poster style, left/right differentiation */}
                    <div className="flex-1 overflow-y-auto px-5 py-4 flex flex-col justify-center">
                      {card.dialogue.map((line, idx) => {
                        const isDriver = line.role === 'DRIVER';
                        return (
                          <div
                            key={idx}
                            className={`w-full flex flex-col mb-3 ${isDriver ? 'items-start' : 'items-end'}`}
                          >
                            <span
                              className={`text-[13px] font-black tracking-[0.18em] uppercase ${isDriver ? 'text-left' : 'text-right'}`}
                              style={{
                                color: isDriver ? card.colour : '#fff',
                                background: 'none',
                                textShadow: isDriver ? `0 1px 8px ${card.colour}55` : '0 1px 8px #0008',
                                letterSpacing: '0.12em',
                                padding: '0.1em 0',
                                maxWidth: '90%',
                              }}
                            >
                              {line.text}
                            </span>
                            {line.role && (
                              <span
                                className={`text-[10px] font-mono tracking-[0.18em] uppercase mt-1 opacity-70 ${isDriver ? 'text-left' : 'text-right'}`}
                                style={{ color: isDriver ? card.colour : '#aaa', maxWidth: '90%' }}
                              >
                                {line.role}
                              </span>
                            )}
                          </div>
                        );
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
