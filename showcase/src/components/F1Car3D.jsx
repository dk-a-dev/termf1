/**
 * F1Car3D — Vanilla Three.js GLB loader + animation sequence
 *
 * Phase timeline:
 *   0   → 3.2s  (HTML lights overlay plays — car parked at z=-80)
 *   3.2 → 6.0s  "approach"   car rushes nose-first through warp
 *   6.0 → 7.8s  "slowdown"   car decelerates, pivots ~75° to show side profile
 *   7.8 → 9.1s  "curvefast"  banking hard + accelerating; background curves
 *   9.1s+        "exit"       car blasts off-screen right
 *
 * The parent positions this div as absolute inset-0 z-30 with a transparent
 * WebGL canvas – text at z-20 shows through where no car is rendered.
 */
import { useEffect, useRef } from 'react'
import * as THREE from 'three'
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js'

const LIGHTS_OFFSET  = 0.0   // wait time before car appears
const APPROACH_END   = 1.5   // end of nose-first rush (1.5 s approach)
const SLOWDOWN_END   = 2.4   // end of side-profile reveal (0.9 s)
const CURVEFAST_END  = 3.2   // end of hard banking (0.8 s)

export default function F1Car3D({ onPhaseChange }) {
  const mountRef  = useRef(null)
  const cbRef     = useRef(onPhaseChange)
  useEffect(() => { cbRef.current = onPhaseChange }, [onPhaseChange])

  useEffect(() => {
    const el = mountRef.current
    if (!el) return

    // Audio is managed by the parent (Hero.jsx) so it can start
    // during the lights sequence before this canvas begins animation.

    /* ── Renderer ──────────────────────────────────────────────── */
    const renderer = new THREE.WebGLRenderer({ alpha: true, antialias: true })
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
    renderer.setSize(el.clientWidth, el.clientHeight)
    renderer.outputColorSpace = THREE.SRGBColorSpace
    renderer.toneMapping      = THREE.ACESFilmicToneMapping
    renderer.toneMappingExposure = 1.4
    el.appendChild(renderer.domElement)

    /* ── Scene & Camera ────────────────────────────────────────── */
    const scene  = new THREE.Scene()
    const camera = new THREE.PerspectiveCamera(52, el.clientWidth / el.clientHeight, 0.1, 300)
    camera.position.set(0, 1.6, 7.5)
    camera.lookAt(0, 0.4, 0)

    /* ── Lights ────────────────────────────────────────────────── */
    scene.add(new THREE.AmbientLight(0xffffff, 0.6))

    const sun = new THREE.DirectionalLight(0xffffff, 2.5)
    sun.position.set(6, 12, 8)
    scene.add(sun)

    // Red rim light – front, gives dramatic F1 reflection on white car
    const rimRed = new THREE.PointLight(0xe8002d, 12, 35)
    rimRed.position.set(0, 2, 6)
    scene.add(rimRed)

    // Cool fill from above-left
    const fillBlue = new THREE.PointLight(0x6699ff, 4, 30)
    fillBlue.position.set(-6, 8, 4)
    scene.add(fillBlue)

    // Underbody glow
    const under = new THREE.PointLight(0xff2200, 3, 15)
    under.position.set(0, -3, 2)
    scene.add(under)

    /* ── Car group ─────────────────────────────────────────────── */
    const carGroup = new THREE.Group()
    carGroup.position.set(0, 0, -80)
    scene.add(carGroup)

    // Materials array for runtime emissive animation
    const bodyMaterials = []

    /* ── Load GLB ──────────────────────────────────────────────── */
    const loader = new GLTFLoader()
    loader.load(
      '/models/formula_1_car_version_1.glb',
      (gltf) => {
        const model = gltf.scene

        // Scale: fit longest dimension to 5.5 world units
        const box  = new THREE.Box3().setFromObject(model)
        const size = new THREE.Vector3()
        box.getSize(size)
        model.scale.setScalar(5.5 / Math.max(size.x, size.y, size.z))

        // Auto-orient: rotate so car's FRONT faces camera (+Z)
        if (size.x >= size.z) {
          model.rotation.y = Math.PI / 2    // body runs X → nose at -X → face +Z
        } else {
          model.rotation.y = 0              // body runs Z → nose at -Z → face +Z
        }

        // Re-compute after transform; centre on XZ, floor on Y
        model.updateMatrixWorld(true)
        const box2   = new THREE.Box3().setFromObject(model)
        const center = new THREE.Vector3()
        box2.getCenter(center)
        model.position.x = -center.x
        model.position.z = -center.z
        model.position.y = -box2.min.y   // seat car on ground

        // Repaint untextured white GLB to F1 Ferrari red
        model.traverse((child) => {
          if (!child.isMesh) return
          const mat = new THREE.MeshStandardMaterial({
            color:             new THREE.Color('#C30000'),
            metalness:         0.65,
            roughness:         0.30,
            emissive:          new THREE.Color('#000000'),
            emissiveIntensity: 0,
          })
          child.material = mat
          bodyMaterials.push(mat)
        })

        carGroup.add(model)
      },
      undefined,
      (err) => console.warn('GLB load error:', err)
    )

    /* ── Animation loop ────────────────────────────────────────── */
    const clock = new THREE.Clock()
    let   phase = ''
    let   alive = true

    const easeInCubic    = (t) => t * t * t
    const easeInOutCubic = (t) =>
      t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2

    // Easing that shoots up faster then slows down (great for rotation leading position)
    const easeOutQuart = (t) => 1 - Math.pow(1 - t, 4)

    const setPhase = (p) => {
      if (phase === p) return
      phase = p
      cbRef.current?.(p)
    }

    const animate = () => {
      if (!alive) return
      requestAnimationFrame(animate)

      const t = clock.getElapsedTime()

      if (t < LIGHTS_OFFSET) {
        // ── LIGHTS: car invisible, parked far back; HTML overlay plays ──
        carGroup.position.set(-60, 0, -40)
        carGroup.rotation.set(0, Math.PI / 4, 0)
        bodyMaterials.forEach(m => { m.emissiveIntensity = 0 })
        under.intensity  = 3
        rimRed.intensity = 4

      } else if (t < APPROACH_END) {
        // ── APPROACH: originate from center under lights, curl out left, sweep toward camera (1.5 s) ──
        setPhase('approach')
        const p  = easeInCubic((t - LIGHTS_OFFSET) / (APPROACH_END - LIGHTS_OFFSET))
        
        // Starts at x=0, wide to x=-12 mid-way, returns to x=0 at the end
        const outCurve = Math.sin(p * Math.PI)
        carGroup.position.set(-12 * outCurve, 0, -55 + 57 * p)
        
        // Starts straight, turns nose out left during curl, straightens exactly to 0 at the end
        const yaw = -Math.sin(p * Math.PI) * (Math.PI / 8)
        carGroup.rotation.set(0, yaw, 0)

        const glow = Math.max(0, (p - 0.4) / 0.6)
        bodyMaterials.forEach(m => {
          m.emissive.setHex(0xe8002d)
          m.emissiveIntensity = glow * 0.40
        })
        under.intensity  = 3  + glow * 6
        rimRed.intensity = 12 + glow * 16

      } else if (t < SLOWDOWN_END) {
        // ── SLOWDOWN: car brakes, pivots to reveal sleek left-flank (0.9 s) ──
        setPhase('slowdown')
        const sp = (t - APPROACH_END) / (SLOWDOWN_END - APPROACH_END)
        const se = easeInOutCubic(sp)
        const rotE = easeOutQuart(sp) // Rotation leads position

        // Flows right much faster
        carGroup.position.x =  se * 7.5
        carGroup.position.y =  se * 0.3
        carGroup.position.z =  2.0 - se * 1.5
        
        // Nose lifts slightly under aerodynamic load/turn
        const pitch = Math.sin(sp * Math.PI) * 0.08
        // Slight wobble on the roll
        const wobble = Math.sin(sp * Math.PI * 2) * 0.02

        carGroup.rotation.x = pitch
        carGroup.rotation.y =  rotE * Math.PI * 0.48   // Nose turns ahead of movement
        carGroup.rotation.z = -rotE * 0.15 + wobble    // Banks into turn with slight wobble

        bodyMaterials.forEach(m => {
          m.emissive.setHex(0xe8002d)
          m.emissiveIntensity = (1 - se) * 0.30
        })
        under.intensity  = 3
        rimRed.intensity = 12

      } else if (t < CURVEFAST_END) {
        // ── CURVEFAST: hard banking right, accelerates into the turn (0.8 s) ──
        setPhase('curvefast')
        const cp = (t - SLOWDOWN_END) / (CURVEFAST_END - SLOWDOWN_END)
        const ce = easeInCubic(cp)
        const rotE = easeOutQuart(cp)

        carGroup.position.x =  7.5 + ce * 8.5
        carGroup.position.y =  0.3 + ce * 0.6
        carGroup.position.z =  0.5 + ce * 2.5
        
        carGroup.rotation.x = 0
        carGroup.rotation.y =  Math.PI * 0.48 + rotE * Math.PI * 0.12  // continue rotating
        carGroup.rotation.z = -0.15 - rotE * 0.10                      // max banking

        bodyMaterials.forEach(m => {
          m.emissive.setHex(0xe8002d)
          m.emissiveIntensity = ce * 0.18
        })
        rimRed.intensity = 12 + ce * 18

      } else {
        // ── EXIT: quadratic blast off-screen right ──
        setPhase('exit')
        const ep = t - CURVEFAST_END
        carGroup.position.x =  16.0 + ep * ep * 125
        carGroup.position.y =   0.9 + ep * 2.2
        carGroup.position.z =   3.0 + ep * 3.0
        carGroup.rotation.y =  Math.PI * 0.60
        carGroup.rotation.z = -0.22 - ep * 0.08
      }

      renderer.render(scene, camera)
    }
    animate()

    /* ── Resize ────────────────────────────────────────────────── */
    const onResize = () => {
      const w = el.clientWidth, h = el.clientHeight
      camera.aspect = w / h
      camera.updateProjectionMatrix()
      renderer.setSize(w, h)
    }
    window.addEventListener('resize', onResize)

    return () => {
      alive = false
      window.removeEventListener('resize', onResize)
      renderer.dispose()
      if (el.contains(renderer.domElement)) el.removeChild(renderer.domElement)
    }
  }, [])

  return (
    <div
      ref={mountRef}
      style={{ position: 'absolute', inset: 0, pointerEvents: 'none' }}
    />
  )
}