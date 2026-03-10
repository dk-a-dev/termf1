/**
 * AsciiText — terminal character-scramble reveal effect.
 *
 * While characters are resolving they flicker through random ASCII symbols
 * in #E8002D red (like a terminal decoding). Once locked, they adopt the
 * parent className / style colours.
 *
 * Props:
 *   text          — string to display
 *   active        — trigger the reveal (boolean)
 *   staggerMs     — ms between each character starting its scramble (default 45)
 *   scrambleTicks — how many random-char frames before locking (default 10)
 *   className     — passed to outer <span>
 *   style         — passed to outer <span>
 */
import { useState, useEffect, useRef } from 'react'

const POOL = '!@#$%^&*<>[]{}|~+-=?/\\▓░▒█▄▀◆●■▲ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'

const rnd = () => POOL[Math.floor(Math.random() * POOL.length)]

export default function AsciiText({
  text = '',
  active = false,
  staggerMs = 45,
  scrambleTicks = 10,
  className = '',
  style = {},
}) {
  // null = not yet shown, string = current char (may be scrambled)
  const [chars, setChars] = useState(() => Array(text.length).fill(null))
  const timerRef = useRef(null)
  const intRef   = useRef([])

  useEffect(() => {
    if (!active) return

    const letters = text.split('')

    // Clear any ongoing intervals
    intRef.current.forEach(clearInterval)
    intRef.current = []

    const revealChar = (idx) => {
      if (letters[idx] === ' ') {
        setChars((c) => { const n = [...c]; n[idx] = ' '; return n })
        return
      }
      let ticks = 0
      const id = setInterval(() => {
        ticks++
        if (ticks >= scrambleTicks) {
          clearInterval(id)
          setChars((c) => { const n = [...c]; n[idx] = letters[idx]; return n })
        } else {
          setChars((c) => { const n = [...c]; n[idx] = rnd(); return n })
        }
      }, 38)
      intRef.current.push(id)
    }

    // Stagger each character
    letters.forEach((_, i) => {
      const t = setTimeout(() => revealChar(i), i * staggerMs)
      intRef.current.push(t)
    })

    return () => {
      intRef.current.forEach((id) => { clearInterval(id); clearTimeout(id) })
    }
  }, [active, text, staggerMs, scrambleTicks])

  return (
    <span className={className} style={style}>
      {chars.map((ch, i) => {
        const isScrambled = ch !== null && ch !== text[i] && text[i] !== ' '
        // Non-breaking space keeps width for inline-block space chars
        const displayChar = (c) => (c === ' ' ? '\u00A0' : c)
        return (
          <span
            key={i}
            style={
              isScrambled
                ? {
                    color: '#E8002D',
                    textShadow: '0 0 10px rgba(232,0,45,0.8)',
                    display: 'inline-block',
                  }
                : { display: 'inline-block', opacity: ch === null ? 0 : 1, transition: 'opacity 0.08s' }
            }
          >
            {ch === null ? displayChar(text[i]) : displayChar(ch)}
          </span>
        )
      })}
    </span>
  )
}
