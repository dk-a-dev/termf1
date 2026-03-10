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
const APPROACH_END   = 1.3   // car rushes to center of screen
const SPIN_END       = 2.4   // 360° spin showcase at center (1.1 s)
const EXIT_START     = 2.4   // blast off right

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

    // Smoother quintic easings for fluid motion
    const easeInQuint     = (t) => t * t * t * t * t
    const easeInOutQuint  = (t) =>
      t < 0.5 ? 16 * t * t * t * t * t : 1 - Math.pow(-2 * t + 2, 5) / 2
    const easeOutQuint    = (t) => 1 - Math.pow(1 - t, 5)
    const easeInOutSine   = (t) => -(Math.cos(Math.PI * t) - 1) / 2

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
        // ── APPROACH: rush straight to center of screen ──
        setPhase('approach')
        const raw = (t - LIGHTS_OFFSET) / (APPROACH_END - LIGHTS_OFFSET)
        const p   = easeInOutQuint(raw)
        
        // Slight S-curve but ends dead-center (x≈0, z≈2)
        const outCurve = Math.sin(p * Math.PI)
        const yBob = Math.sin(p * Math.PI * 2) * 0.06
        carGroup.position.set(-6 * outCurve, yBob, -55 + 57 * p)
        
        const yaw  = -Math.sin(p * Math.PI) * (Math.PI / 8)
        const roll =  Math.sin(p * Math.PI) * 0.04
        carGroup.rotation.set(0, yaw, roll)

        const glow = Math.max(0, (p - 0.3) / 0.7)
        bodyMaterials.forEach(m => {
          m.emissive.setHex(0xe8002d)
          m.emissiveIntensity = glow * 0.45
        })
        under.intensity  = 3  + glow * 8
        rimRed.intensity = 12 + glow * 18

      } else if (t < SPIN_END) {
        // ── SPIN 360°: car holds at center, quick full rotation showcase ──
        setPhase('slowdown')
        const sp = (t - APPROACH_END) / (SPIN_END - APPROACH_END)
        const se = easeInOutQuint(sp)

        // Stay dead center (x=0), float up slightly mid-spin
        carGroup.position.x = 0
        carGroup.position.y = Math.sin(sp * Math.PI) * 0.35
        carGroup.position.z = 2.0

        // Full 360° spin — starts nose-on (yaw≈0), completes full revolution
        // Spin from nose-on (0) to facing right (-Math.PI/2)
        const spinAngle = se * (Math.PI * 2 + Math.PI / 2)
        carGroup.rotation.x = Math.sin(sp * Math.PI) * 0.04
        carGroup.rotation.y = spinAngle
        carGroup.rotation.z = Math.sin(sp * Math.PI * 2) * 0.05

        // Pulsing glow during spin
        const pulse = 0.18 + Math.sin(sp * Math.PI * 4) * 0.12
        bodyMaterials.forEach(m => {
          m.emissive.setHex(0xe8002d)
          m.emissiveIntensity = pulse
        })
        rimRed.intensity = 14 + Math.sin(sp * Math.PI * 2) * 10
        under.intensity  = 3  + Math.sin(sp * Math.PI) * 6

      } else {
        // ── EXIT: blast off-screen right after the spin ──
        setPhase('exit')
        const ep = t - EXIT_START
        const exitEase = easeInQuint(Math.min(ep / 0.6, 1))

        // Rockets right from center (x=0)
        carGroup.position.x =  exitEase * 25 + ep * ep * 80
        carGroup.position.y =  Math.max(0, 0.35 * (1 - ep * 2)) + ep * 0.8
        carGroup.position.z =  2.0 + ep * 2.5
        // Nose points right toward exit
        // Start exit facing right (-Math.PI/2), then tilt slightly as it blasts off
        carGroup.rotation.y = Math.PI / 2 + exitEase * Math.PI * 0.55
        carGroup.rotation.z = -exitEase * 0.16
        carGroup.rotation.x =  0
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