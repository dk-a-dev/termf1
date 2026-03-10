import { motion } from 'framer-motion'
import { useState, useEffect, useRef } from 'react'

/* ── Pure-CSS pixel art: bikini grid girl with chequered flag ──────────────
   Each "pixel" is a 3×3 CSS box-shadow on a tiny 1×1 element.
   Walking is done by toggling between two leg frames via CSS animation.    */

const PX = 3 // pixel size
const flagPixels = [
  // Chequered flag (6×5 grid, offset right from her hand)
  [11,0,'#fff'],[12,0,'#222'],[13,0,'#fff'],[14,0,'#222'],[15,0,'#fff'],[16,0,'#222'],
  [11,1,'#222'],[12,1,'#fff'],[13,1,'#222'],[14,1,'#fff'],[15,1,'#222'],[16,1,'#fff'],
  [11,2,'#fff'],[12,2,'#222'],[13,2,'#fff'],[14,2,'#222'],[15,2,'#fff'],[16,2,'#222'],
  [11,3,'#222'],[12,3,'#fff'],[13,3,'#222'],[14,3,'#fff'],[15,3,'#222'],[16,3,'#fff'],
  [11,4,'#fff'],[12,4,'#222'],[13,4,'#fff'],[14,4,'#222'],[15,4,'#fff'],[16,4,'#222'],
  // Flag pole
  [10,0,'#aaa'],[10,1,'#999'],[10,2,'#888'],[10,3,'#888'],[10,4,'#777'],
  [10,5,'#777'],[10,6,'#666'],[10,7,'#666'],[10,8,'#555'],
]

const bodyPixels = [
  // Hair (flowing dark brown, slightly longer for a feminine look)
  [4,0,'#3B1F0B'],[5,0,'#3B1F0B'],[6,0,'#3B1F0B'],[7,0,'#3B1F0B'],
  [3,1,'#3B1F0B'],[4,1,'#3B1F0B'],[5,1,'#3B1F0B'],[6,1,'#3B1F0B'],[7,1,'#3B1F0B'],[8,1,'#3B1F0B'],
  [3,2,'#3B1F0B'],[8,2,'#3B1F0B'], // hair sides framing face
  // Face (skin tone)
  [4,2,'#F5C5A3'],[5,2,'#F5C5A3'],[6,2,'#F5C5A3'],[7,2,'#F5C5A3'],
  [4,3,'#F5C5A3'],[5,3,'#F5C5A3'],[6,3,'#F5C5A3'],[7,3,'#F5C5A3'],
  // Eyes (dark) + smile
  [5,2,'#222'],[7,2,'#222'],
  [5,3,'#E06060'],[6,3,'#E06060'], // lips/smile
  // Neck
  [5,4,'#F5C5A3'],[6,4,'#F5C5A3'],
  // Bikini top (red)
  [4,5,'#E8002D'],[5,5,'#C70025'],[6,5,'#C70025'],[7,5,'#E8002D'],
  [3,5,'#F5C5A3'], // shoulder
  // Torso (skin)
  [4,6,'#F5C5A3'],[5,6,'#F5C5A3'],[6,6,'#F5C5A3'],[7,6,'#F5C5A3'],
  [4,7,'#F5C5A3'],[5,7,'#F5C5A3'],[6,7,'#F5C5A3'],[7,7,'#F5C5A3'],
  // Bikini bottom (red)
  [4,8,'#E8002D'],[5,8,'#C70025'],[6,8,'#C70025'],[7,8,'#E8002D'],
  // Arm holding flag (reaching right toward pole at x=10)
  [8,5,'#F5C5A3'],[9,5,'#F5C5A3'],[10,5,'#F5C5A3'],
  [8,6,'#F5C5A3'],
  // Other arm resting (left side)
  [3,6,'#F5C5A3'],[2,6,'#F5C5A3'],
]

// Two frames for walking legs
const legsFrame1 = [
  // Left leg forward, right leg back
  [4,9,'#F5C5A3'],[4,10,'#F5C5A3'],[3,11,'#F5C5A3'],
  [7,9,'#F5C5A3'],[7,10,'#F5C5A3'],[8,11,'#F5C5A3'],
  // Red heels
  [3,12,'#E8002D'],[8,12,'#E8002D'],
]
const legsFrame2 = [
  // Swap legs
  [5,9,'#F5C5A3'],[5,10,'#F5C5A3'],[5,11,'#F5C5A3'],
  [6,9,'#F5C5A3'],[6,10,'#F5C5A3'],[6,11,'#F5C5A3'],
  // Red heels closer together
  [5,12,'#E8002D'],[6,12,'#E8002D'],
]

function pixelsToShadow(pixels) {
  return pixels
    .map(([x, y, c]) => `${x * PX}px ${y * PX}px 0 ${c}`)
    .join(',')
}

const shadowBody = pixelsToShadow([...flagPixels, ...bodyPixels])
const shadowLegs1 = pixelsToShadow(legsFrame1)
const shadowLegs2 = pixelsToShadow(legsFrame2)

function PixelGirl({ walking }) {
  return (
    <div className="relative" style={{ width: 17 * PX, height: 13 * PX }}>
      {/* Body + flag (static) */}
      <div
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          width: PX,
          height: PX,
          background: 'transparent',
          boxShadow: shadowBody,
          imageRendering: 'pixelated',
        }}
      />
      {/* Legs (animated between two frames) */}
      <div
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          width: PX,
          height: PX,
          background: 'transparent',
          boxShadow: walking ? undefined : shadowLegs1,
          animation: walking ? 'pixelWalk 0.35s steps(1) infinite' : 'none',
          imageRendering: 'pixelated',
        }}
      />
      <style>{`
        @keyframes pixelWalk {
          0%, 49.9%  { box-shadow: ${shadowLegs1}; }
          50%, 100%  { box-shadow: ${shadowLegs2}; }
        }
      `}</style>
    </div>
  )
}

/* ── Loading bar with walking sprite ───────────────────────────────────────── */
export default function PixelLoadingBar({ loaded, onStart }) {
  const [progress, setProgress] = useState(0)
  const prevLoaded = useRef(false)

  useEffect(() => {
    if (loaded) {
      // Smoothly fill to 100 over ~600ms
      const fill = setInterval(() => {
        setProgress(p => {
          if (p >= 100) { clearInterval(fill); return 100 }
          return Math.min(p + 2, 100)
        })
      }, 12)
      prevLoaded.current = true
      return () => clearInterval(fill)
    }
    // Slow fake progress while loading — realistic stalls and bursts
    const interval = setInterval(() => {
      setProgress(p => {
        if (p >= 88) return 88
        // Simulate realistic loading: occasional stalls, varying speeds
        const zone = p < 20 ? 3 : p < 50 ? 2 : p < 75 ? 1.5 : 0.5
        const jitter = Math.random() * zone
        // 20% chance of a "stall" where nothing moves
        if (Math.random() < 0.2) return p
        return Math.min(p + jitter, 88)
      })
    }, 400)
    return () => clearInterval(interval)
  }, [loaded])

  const done = loaded && progress >= 100

  return (
    <div className="flex flex-col items-center select-none">
      {/* termF1 title above loader */}
      <div className="mb-6 text-center">
        <span className="text-2xl font-black tracking-tight" style={{ fontFamily: "'Alphacorsa','Inter',sans-serif" }}>
          <span className="text-white">term</span>
          <span style={{ color: '#E8002D' }}>F1</span>
        </span>
      </div>

      {/* Loading bar track */}
      <div className="relative w-72 sm:w-80">
        {/* Track background */}
        <div className="w-full h-[6px] rounded-full overflow-hidden" style={{ background: '#1A1C23' }}>
          <motion.div
            className="h-full rounded-full"
            style={{ background: 'linear-gradient(90deg, #E8002D, #ff3355)' }}
            animate={{ width: `${progress}%` }}
            transition={{ ease: 'linear', duration: 0.15 }}
          />
        </div>

        {/* Pixel girl walking on top of the bar */}
        <motion.div
          className="absolute z-10"
          style={{ bottom: 6, marginLeft: -24 }}
          animate={{ left: `${progress}%` }}
          transition={{ ease: 'linear', duration: 0.15 }}
        >
          <PixelGirl walking={!done} />
        </motion.div>
      </div>

      {/* Status text */}
      <div className="mt-10">
        {!done ? (
          <div className="flex items-center gap-3">
            <div className="text-gray-600 font-mono text-[10px] tracking-[0.3em] uppercase">
              Loading assets
            </div>
            <div className="text-gray-600 font-mono text-[10px]">
              {Math.round(progress)}%
            </div>
          </div>
        ) : (
          <motion.button
            onClick={onStart}
            initial={{ opacity: 0, y: 6 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4 }}
            className="group flex flex-col items-center gap-1 cursor-pointer"
          >
            <span className="text-[#E8002D] font-black text-sm tracking-[0.2em] uppercase group-hover:text-white transition-colors">
              START RACE
            </span>
            <span className="text-gray-600 text-[9px] font-mono tracking-widest group-hover:text-gray-400 transition-colors">
              CLICK TO BEGIN
            </span>
          </motion.button>
        )}
      </div>
    </div>
  )
}
