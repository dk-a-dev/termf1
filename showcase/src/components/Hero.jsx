/**
 * Hero — full-screen dramatic intro sequence
 *
 * Z-layer stack:
 *   z-0   Stars (WarpCanvas) — density driven by car phase
 *   z-10  Radial vignette
 *   z-20  Title + CTAs  ← visible behind car through transparent WebGL bg
 *   z-30  3D car canvas (alpha:true — transparent where no geometry)
 *   z-40  White flash
 *
 * Car animation phases (F1Car3D):
 *   "approach" (0–2.8s)   → dense stars, car rushes nose-first
 *   "turn"     (2.8–4.1s) → car pivots left, title scramble starts
 *   "exit"     (4.1s+)    → car blasts off-screen; title fully resolves
 */
import { useState, useRef, useCallback, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import F1Car3D   from './F1Car3D'
import AsciiText from './AsciiText'
import PixelLoadingBar from './PixelLoadingBar'

/* ── Star-warp canvas ──────────────────────────────────────────────────────── */
function WarpCanvas({ phaseRef }) {
  const ref = useRef(null)

  useEffect(() => {
    const c = ref.current
    if (!c) return
    const ctx = c.getContext('2d')

    const resize = () => {
      c.width  = window.innerWidth
      c.height = window.innerHeight
    }
    resize()
    window.addEventListener('resize', resize)

    const MAX = 750
    const streaks = Array.from({ length: MAX }, () => ({
      angle : Math.random() * Math.PI * 2,
      pos   : Math.random(),
      speed : 0.003 + Math.random() * 0.013,
      len   : 0.03  + Math.random() * 0.09,
      alpha : 0.06  + Math.random() * 0.38,
      red   : Math.random() > 0.84,
      thick : Math.random() > 0.5 ? 1.6 : 0.7,
    }))

    let raf
    const draw = () => {
      const ph = phaseRef.current
      let density = 0.18, speedMult = 0.6, glowA = 0.03
      if      (ph === 'approach')  { density = 0.95; speedMult = 2.8;  glowA = 0.16 }
      else if (ph === 'slowdown')  { density = 0.85; speedMult = 1.9;  glowA = 0.20 }
      else if (ph === 'curvefast') { density = 0.72; speedMult = 2.6;  glowA = 0.28 }
      else if (ph === 'exit')      { density = 0.28; speedMult = 1.0;  glowA = 0.05 }

      const N  = Math.floor(MAX * density)
      const w  = c.width, h = c.height
      let cx = w / 2, cy = h / 2

      // Shift the vanishing point to simulate the camera panning into the corner
      if (ph === 'curvefast') {
        cx += w * 0.26
        cy += h * 0.05
      } else if (ph === 'slowdown') {
        cx += w * 0.13
      }

      const maxR = Math.hypot(cx, cy)

      ctx.fillStyle = 'rgba(0,0,0,0.20)'
      ctx.fillRect(0, 0, w, h)

      for (let i = 0; i < N; i++) {
        const s = streaks[i]
        s.pos += s.speed * speedMult
        if (s.pos > 1) s.pos = 0

        const r1 = maxR * s.pos
        const r2 = maxR * Math.min(s.pos + s.len, 1)
        const x1 = cx + Math.cos(s.angle) * r1, y1 = cy + Math.sin(s.angle) * r1
        const x2 = cx + Math.cos(s.angle) * r2, y2 = cy + Math.sin(s.angle) * r2

        const g = ctx.createLinearGradient(x1, y1, x2, y2)
        g.addColorStop(0, 'rgba(0,0,0,0)')
        if (s.red) {
          g.addColorStop(0.5,  `rgba(232,0,45,${s.alpha * 0.7})`)
          g.addColorStop(0.85, `rgba(232,0,45,${s.alpha})`)
        } else {
          g.addColorStop(0.5,  `rgba(200,210,255,${s.alpha * 0.45})`)
          g.addColorStop(0.85, `rgba(255,255,255,${s.alpha})`)
        }
        g.addColorStop(1, 'rgba(0,0,0,0)')

        ctx.beginPath()
        ctx.strokeStyle = g
        ctx.lineWidth   = s.thick
        ctx.moveTo(x1, y1)
        ctx.lineTo(x2, y2)
        ctx.stroke()
      }

      // Central red glow — pulses with car proximity
      const rg = ctx.createRadialGradient(cx, cy, 0, cx, cy, Math.min(w, h) * 0.38)
      rg.addColorStop(0, `rgba(232,0,45,${glowA})`)
      rg.addColorStop(1, 'rgba(0,0,0,0)')
      ctx.fillStyle = rg
      ctx.fillRect(0, 0, w, h)

      raf = requestAnimationFrame(draw)
    }
    draw()

    return () => {
      cancelAnimationFrame(raf)
      window.removeEventListener('resize', resize)
    }
  }, [phaseRef])

  return <canvas ref={ref} className="absolute inset-0 w-full h-full" />
}

/* ── F1 start lights overlay ───────────────────────────────────────────────── */
function StartLights({ onLightsOut }) {
  const [step, setStep] = useState(0) // 0 idle, 1–5 lights on, 6 all off
  const [tagline, setTagline] = useState('')
  const audioRef = useRef(null)

  useEffect(() => {
    const taglines = [
      "IT'S LIGHTS OUT AND AWAY WE GO!",
      'WOAAAH! TURN UP THE VOLUME!',
      'FIVE LIGHTS...AND THEY’RE OFF!',
    ]
    setTagline(taglines[Math.floor(Math.random() * taglines.length)])
  }, [])

  useEffect(() => {
    // Play the full lights-out.mp3 which has 1s pings built in.
    const a = new Audio('/audios/lights-out.mp3')
    a.volume = 0.8
    audioRef.current = a
    
    // Attempt play immediately, relies on parent having unlocked audio with click
    a.play().catch(e => console.warn('Lights audio blocked:', e))

    let tick = 0
    // To match 1s pings, we assume the pings are exactly 1s apart, starting around 1s in.
    const id = window.setInterval(() => {
      tick += 1
      // Lights 1 to 5 turn on exactly every 1000ms
      if (tick <= 5) {
        setStep(tick)
      }
      // At the 6th second, all lights go off, car blasts away.
      if (tick === 6) {
        setStep(6)
        setTimeout(() => onLightsOut?.(), 200) // gentle hide
        window.clearInterval(id)

        // Fade out lights out audio
        if (audioRef.current) {
          const fadeAudio = setInterval(() => {
            if (audioRef.current && audioRef.current.volume > 0.05) {
              audioRef.current.volume -= 0.05
            } else {
              if (audioRef.current) {
                audioRef.current.pause()
                audioRef.current.src = ''
                audioRef.current = null
              }
              clearInterval(fadeAudio)
            }
          }, 40)
        }
      }
    }, 1000)

    return () => {
      window.clearInterval(id)
      if (audioRef.current && tick < 6) {
        audioRef.current.pause()
        audioRef.current.src = ''
        audioRef.current = null
      }
    }
  }, [onLightsOut])

  // Once lights are out, fade the whole UI away
  const isDone = step >= 6

  return (
    <motion.div
      className="absolute inset-0 z-25 flex flex-col items-center justify-center pointer-events-none"
      initial={{ opacity: 0 }}
      animate={{ opacity: isDone ? 0 : 1 }}
      transition={{ duration: 0.4, ease: [0.16, 1, 0.3, 1] }}
    >
      <div className="flex gap-4 mb-6">
        {[0, 1, 2, 3, 4].map((idx) => {
          const lit = step > idx && step < 6
          return (
            <motion.div
              key={idx}
              className="w-10 h-10 md:w-12 md:h-12 rounded-full border border-red-900 shadow-lg"
              style={{
                background: lit ? '#E8002D' : '#25040a',
                boxShadow: lit
                  ? '0 0 35px rgba(232,0,45,0.9), 0 0 70px rgba(232,0,45,0.5)'
                  : '0 0 10px rgba(0,0,0,0.9)',
              }}
              animate={lit ? { scale: [1, 1.1, 1] } : { scale: 1 }}
              transition={{ duration: 0.4, repeat: lit ? Infinity : 0, repeatType: 'mirror' }}
            />
          )
        })}
      </div>
      {!isDone && (
        <div
          className="text-[11px] md:text-xs font-mono tracking-[0.25em] uppercase"
          style={{
            color: isDone ? '#E8002D' : '#6B7280',
            transition: 'color 0.4s ease',
          }}
        >
          {tagline}
        </div>
      )}
    </motion.div>
  )
}

/* ── Stat pill ─────────────────────────────────────────────────────────────── */
function Stat({ num, label }) {
  return (
    <div className="text-center">
      <div className="text-2xl font-black text-white tracking-tight">{num}</div>
      <div className="text-[10px] text-gray-500 mt-0.5 font-mono tracking-widest uppercase">{label}</div>
    </div>
  )
}

function GithubIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
      <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57
               0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695
               -.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99
               .105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225
               -.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405
               c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225
               0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3
               0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z" />
    </svg>
  )
}

/* ── Hero ──────────────────────────────────────────────────────────────────── */
export default function Hero() {
  const [loaded, setLoaded] = useState(false)
  const [started, setStarted] = useState(false)
  const [loadingText, setLoadingText] = useState('LOADING ASSETS...')

  // UI-phase drives text visibility: 'idle' → 'turning' → 'revealed'
  const [uiPhase, setUiPhase] = useState('idle')
  const [lightsDone, setLightsDone] = useState(false)
  // phaseRef is read by WarpCanvas every frame (no re-render)
  const phaseRef     = useRef('approach')
  const engineRef    = useRef(null)
  const engineFaded  = useRef(false)

  // Preload assets
  useEffect(() => {
    let unmounted = false
    const loadAssets = async () => {
      try {
        const p1 = new Promise(r => {
          const audio1 = new Audio()
          audio1.oncanplaythrough = r
          audio1.onerror = r
          audio1.src = '/audios/lights-out.mp3'
        })
        const p2 = new Promise(r => {
          const audio2 = new Audio()
          audio2.oncanplaythrough = r
          audio2.onerror = r
          audio2.src = '/audios/racing-car.mp3'
        })
        // Simple fetch preload for GLB
        const p3 = fetch('/models/formula_1_car_version_1.glb').then(res => res.blob()).catch(() => {})

        await Promise.all([p1, p2, p3])
        if (!unmounted) setLoaded(true)
      } catch (e) {
        if (!unmounted) setLoaded(true) // failsafe
      }
    }
    loadAssets()
    return () => { unmounted = true }
  }, [])

  // Setup engine audio
  useEffect(() => {
    const audio = new Audio('/audios/racing-car.mp3')
    audio.volume = 0.0 // silent until car appears
    audio.loop   = true
    engineRef.current = audio

    return () => {
      audio.pause()
      audio.src = ''
    }
  }, [])

  const handleStart = () => {
    if (!loaded) return
    setStarted(true)
    if (engineRef.current) {
      engineRef.current.play().catch(() => {})
    }
  }

  // Play the loud engine scream precisely when the car flies out of the dark
  useEffect(() => {
    if (lightsDone && engineRef.current) {
      engineRef.current.currentTime = 0
      engineRef.current.volume = 0.38
      // already playing, but just in case
      engineRef.current.play().catch(() => {})
    }
  }, [lightsDone])

  const handleCarPhase = useCallback((phase) => {
    phaseRef.current = phase
    if (phase === 'curvefast') {
      setUiPhase('turning')
    } else if (phase === 'exit') {
      setTimeout(() => setUiPhase('revealed'), 400)
      // Fade engine audio out over 800 ms
      if (!engineFaded.current && engineRef.current) {
        engineFaded.current = true
        const audio     = engineRef.current
        const startVol  = audio.volume
        const startTime = performance.now()
        const fade = (now) => {
          const t = Math.min(1, (now - startTime) / 800)
          audio.volume = startVol * (1 - t)
          if (t < 1) {
            requestAnimationFrame(fade)
          } else {
            audio.pause()
            audio.removeAttribute('src')
            audio.load()
            engineRef.current = null
          }
        }
        requestAnimationFrame(fade)
      }
    }
  }, [])

  const titleVisible = uiPhase === 'turning' || uiPhase === 'revealed'
  const ctasVisible  = uiPhase === 'revealed'

  return (
    <section className="relative w-full h-screen overflow-hidden bg-black flex items-center justify-center selection:bg-red-500/30">
      
      {!started && (
        <div className="absolute inset-0 z-50 flex flex-col items-center justify-center bg-black">
          <PixelLoadingBar loaded={loaded} onStart={handleStart} />
        </div>
      )}

      {started && (
        <>
          {/* ── 0: Stars — only once the car is moving ───────── */}
          {lightsDone && <WarpCanvas phaseRef={phaseRef} />}

          {/* ── 10: Vignette ───────────────────────────────── */}
          <div
            className="absolute inset-0 z-10 pointer-events-none"
            style={{
              background:
                'radial-gradient(ellipse 62% 62% at 50% 50%, transparent 14%, rgba(0,0,0,0.91) 100%)',
            }}
          />

          {!lightsDone && <StartLights onLightsOut={() => setLightsDone(true)} />}

          {/* 3D WebGL Overlay — transparent, sits between background and UI */}
          {lightsDone && (
            <div className="absolute inset-0 z-30 pointer-events-none">
              <F1Car3D onPhaseChange={handleCarPhase} />
            </div>
          )}
        </>
      )}

      {/* ── 20: Title ────────────────────────────────────── */}
      <div className="absolute inset-0 z-20 flex items-center justify-center" style={{ display: started ? 'flex' : 'none' }}>
        <div
          className="text-center px-6 select-none"
          style={{
            opacity: titleVisible ? 1 : 0,
            transition: 'opacity 0.6s ease',
          }}
        >
          {/* Eye-brow tag */}
          <p className="text-[11px] font-mono tracking-[0.55em] text-[#E8002D] uppercase mb-6">
            <AsciiText
              text="Formula 1 Terminal UI  ·  v2.0"
              active={titleVisible}
              staggerMs={28}
              scrambleTicks={7}
            />
          </p>

          {/* Giant title */}
          <h1
          className="font-black leading-none tracking-tighter"
          style={{ fontSize: 'clamp(5.5rem, 18vw, 13rem)', fontFamily: "'Alphacorsa', 'Speeday', 'Inter', sans-serif" }}
          >
            <AsciiText
              text="term"
              active={titleVisible}
              staggerMs={70}
              scrambleTicks={14}
              style={{ color: '#ffffff' }}
            />
            <AsciiText
              text="F1"
              active={titleVisible}
              staggerMs={70}
              scrambleTicks={14}
              style={{
                color: '#E8002D',
                textShadow: '0 0 55px rgba(232,0,45,0.60), 0 0 110px rgba(232,0,45,0.28)',
              }}
            />
          </h1>

          {/* Tagline */}
          <p className="mt-6 text-base md:text-lg text-gray-400 max-w-md mx-auto leading-relaxed">
            <AsciiText
              text="Every F1 insight you need."
              active={titleVisible}
              staggerMs={28}
              scrambleTicks={6}
            />
            <br />
            <AsciiText
              text="Without leaving your terminal."
              active={titleVisible}
              staggerMs={28}
              scrambleTicks={6}
              style={{ color: '#ffffff', fontWeight: 600 }}
            />
          </p>

          {/* CTAs */}
          <AnimatePresence>
            {ctasVisible && (
              <motion.div
                className="mt-10 flex gap-3 justify-center flex-wrap"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.15, duration: 0.55, ease: [0.16, 1, 0.3, 1] }}
              >
                <a
                  href="https://github.com/dk-a-dev/termf1"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-2 px-7 py-3.5 bg-[#E8002D] text-white font-bold rounded text-sm
                             hover:bg-[#ff1a3d] transition-all duration-150 hover:scale-105 active:scale-95"
                  style={{ boxShadow: '0 0 28px rgba(232,0,45,0.45)' }}
                >
                  <GithubIcon />
                  Star on GitHub
                </a>
                <a
                  href="#install"
                  className="px-7 py-3.5 border border-white/20 text-white font-medium rounded text-sm
                             hover:border-white/50 hover:bg-white/5 transition-all duration-150"
                >
                  Install now ↓
                </a>
                <a
                  href="#screenshots"
                  className="px-7 py-3.5 border border-white/10 text-gray-500 font-medium rounded text-sm
                             hover:border-white/25 hover:text-white transition-all duration-150"
                >
                  See it in action
                </a>
              </motion.div>
            )}
          </AnimatePresence>

          {/* Stats */}
          <AnimatePresence>
            {ctasVisible && (
              <motion.div
                className="mt-12 flex gap-10 justify-center"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ delay: 0.45, duration: 0.55 }}
              >
                {/* <Stat num="8"  label="Views"           />
                <Stat num="9"  label="Analysis Charts" />
                <Stat num="4"  label="Data Sources"    />
                <Stat num="∞"  label="Laps Analysed"   /> */}
              </motion.div>
            )}
          </AnimatePresence>

          {/* Scroll cue */}
          <AnimatePresence>
            {ctasVisible && (
              <motion.div
                className="mt-14 flex flex-col items-center gap-1 text-gray-700"
                animate={{ opacity: [0, 0.55, 0] }}
                transition={{ delay: 1.0, duration: 2.6, repeat: Infinity }}
              >
                <span className="text-[10px] font-mono tracking-[0.42em] uppercase">scroll</span>
                <span className="text-xl">↓</span>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </div>

    </section>
  )
}