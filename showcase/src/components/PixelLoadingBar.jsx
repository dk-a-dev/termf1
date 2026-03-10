import { motion } from 'framer-motion'
import { useState, useEffect, useRef } from 'react'

/* ── Pure-CSS pixel art: bikini girl carrying a chequered flag ─────────────
   Each "pixel" is a 3×3 CSS box-shadow on a tiny 1×1 element.
   Walking is done by toggling between two leg frames via CSS animation.    */

const PX = 3 // pixel size
const flagPixels = [
  // Chequered flag (8×6 grid, offset right from her hand)
  // row 0
  [10,0,'#fff'],[11,0,'#222'],[12,0,'#fff'],[13,0,'#222'],[14,0,'#fff'],[15,0,'#222'],
  // row 1
  [10,1,'#222'],[11,1,'#fff'],[12,1,'#222'],[13,1,'#fff'],[14,1,'#222'],[15,1,'#fff'],
  // row 2
  [10,2,'#fff'],[11,2,'#222'],[12,2,'#fff'],[13,2,'#222'],[14,2,'#fff'],[15,2,'#222'],
  // row 3
  [10,3,'#222'],[11,3,'#fff'],[12,3,'#222'],[13,3,'#fff'],[14,3,'#222'],[15,3,'#fff'],
  // pole
  [9,0,'#888'],[9,1,'#888'],[9,2,'#888'],[9,3,'#888'],[9,4,'#888'],[9,5,'#888'],[9,6,'#888'],[9,7,'#888'],
]

const bodyPixels = [
  // Hair (dark brown)
  [4,0,'#3B1F0B'],[5,0,'#3B1F0B'],[6,0,'#3B1F0B'],
  [3,1,'#3B1F0B'],[4,1,'#3B1F0B'],[5,1,'#3B1F0B'],[6,1,'#3B1F0B'],[7,1,'#3B1F0B'],
  // Face (skin)
  [4,2,'#F5C5A3'],[5,2,'#F5C5A3'],[6,2,'#F5C5A3'],
  [4,3,'#F5C5A3'],[5,3,'#F5C5A3'],[6,3,'#F5C5A3'],
  // Eyes
  [4,2,'#222'],[6,2,'#222'],
  // Neck
  [5,4,'#F5C5A3'],
  // Bikini top (red)
  [4,5,'#E8002D'],[5,5,'#E8002D'],[6,5,'#E8002D'],
  // Torso (skin)
  [4,6,'#F5C5A3'],[5,6,'#F5C5A3'],[6,6,'#F5C5A3'],
  [4,7,'#F5C5A3'],[5,7,'#F5C5A3'],[6,7,'#F5C5A3'],
  // Bikini bottom (red)
  [4,8,'#E8002D'],[5,8,'#E8002D'],[6,8,'#E8002D'],
  // Arm holding flag (skin, reaching right toward pole at x=9)
  [7,5,'#F5C5A3'],[8,5,'#F5C5A3'],[9,5,'#F5C5A3'],
  // Other arm (skin, left side)
  [3,5,'#F5C5A3'],[3,6,'#F5C5A3'],
]

// Two frames for walking legs
const legsFrame1 = [
  // Left leg forward, right leg back
  [4,9,'#F5C5A3'],[4,10,'#F5C5A3'],[4,11,'#F5C5A3'],
  [6,9,'#F5C5A3'],[7,10,'#F5C5A3'],[7,11,'#F5C5A3'],
  // Shoes
  [4,12,'#E8002D'],[7,12,'#E8002D'],
]
const legsFrame2 = [
  // Swap legs
  [4,9,'#F5C5A3'],[3,10,'#F5C5A3'],[3,11,'#F5C5A3'],
  [6,9,'#F5C5A3'],[6,10,'#F5C5A3'],[6,11,'#F5C5A3'],
  // Shoes
  [3,12,'#E8002D'],[6,12,'#E8002D'],
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
    <div className="relative" style={{ width: 16 * PX, height: 13 * PX }}>
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
    // Slow fake progress while loading
    const interval = setInterval(() => {
      setProgress(p => Math.min(p + Math.random() * 8, 88))
    }, 300)
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
