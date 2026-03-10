/**
 * ModelShowcase — 4-card 3D model grid using vanilla Three.js.
 * Each card mounts its own WebGL canvas via useEffect (same pattern as F1Car3D).
 * Models: F1 wheel, steering wheel, Schumacher helmet, starting lights.
 */
import { useEffect, useRef } from 'react'
import * as THREE from 'three'
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js'
import { motion } from 'framer-motion'

const MODELS = [
  {
    path: '/models/formula_1_2012_wheel.glb',
    title: 'F1 Wheel',
    subtitle: '2012 Era',
    desc: 'Single-nut centerlock, 13-inch rim. The iconic multi-spoke design used through the thundering V8 era.',
    color: '#E8002D',
    lightHex: 0xff3333,
  },
  {
    path: '/models/mclaren_formula_1_steering_wheel.glb',
    title: 'Steering Wheel',
    subtitle: 'McLaren Carbon',
    desc: "The pilot's command centre: DRS, ERS mode, brake bias, radio — all within thumb reach at 300km/h.",
    color: '#FF8000',
    lightHex: 0xff8800,
  },
  {
    path: '/models/michael_schumacher_2002_helmet.glb',
    title: 'Schumacher Helmet',
    subtitle: 'Ferrari 2002',
    desc: "Michael Schumacher's 2002 lid — the season he scored 11 wins and clinched his 5th world title.",
    color: '#3671C6',
    lightHex: 0x4488ff,
  },
  {
    path: '/models/race_track_props_starting_lights.glb',
    title: 'Starting Lights',
    subtitle: 'Race Start',
    desc: 'Five red lights, one by one — then darkness. The most electrifying two seconds in motorsport.',
    color: '#22C55E',
    lightHex: 0x00ff55,
  },
]

function ModelCanvas({ model }) {
  const mountRef = useRef(null)

  useEffect(() => {
    const el = mountRef.current
    if (!el) return

    const renderer = new THREE.WebGLRenderer({ alpha: true, antialias: true })
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
    renderer.setSize(el.clientWidth, el.clientHeight)
    renderer.outputColorSpace = THREE.SRGBColorSpace
    renderer.toneMapping = THREE.ACESFilmicToneMapping
    renderer.toneMappingExposure = 1.3
    el.appendChild(renderer.domElement)

    const scene  = new THREE.Scene()
    const camera = new THREE.PerspectiveCamera(46, el.clientWidth / el.clientHeight, 0.1, 100)
    camera.position.set(0, 1.2, 5)
    camera.lookAt(0, 0.3, 0)

    scene.add(new THREE.AmbientLight(0xffffff, 0.5))
    const key = new THREE.DirectionalLight(0xffffff, 2.0)
    key.position.set(4, 8, 5)
    scene.add(key)
    const accent = new THREE.PointLight(model.lightHex, 18, 20)
    accent.position.set(-3, 2, 3)
    scene.add(accent)
    const fill = new THREE.PointLight(0xaabbff, 3, 15)
    fill.position.set(3, -2, 2)
    scene.add(fill)

    const group = new THREE.Group()
    scene.add(group)

    // Gentle float
    let floatT = Math.random() * Math.PI * 2

    const loader = new GLTFLoader()
    loader.load(
      model.path,
      (gltf) => {
        const m = gltf.scene
        const box  = new THREE.Box3().setFromObject(m)
        const size = new THREE.Vector3()
        box.getSize(size)
        m.scale.setScalar(2.6 / Math.max(size.x, size.y, size.z))
        m.updateMatrixWorld(true)
        const box2   = new THREE.Box3().setFromObject(m)
        const center = new THREE.Vector3()
        box2.getCenter(center)
        m.position.sub(center)
        group.add(m)
      },
      undefined,
      (err) => console.warn('ModelShowcase GLB error:', model.path, err),
    )

    let raf, alive = true
    const clock = new THREE.Clock()
    const animate = () => {
      if (!alive) return
      raf = requestAnimationFrame(animate)
      const dt = clock.getDelta()
      floatT += dt * 1.1
      group.rotation.y += dt * 0.44
      group.position.y = Math.sin(floatT) * 0.12
      renderer.render(scene, camera)
    }
    animate()

    const onResize = () => {
      camera.aspect = el.clientWidth / el.clientHeight
      camera.updateProjectionMatrix()
      renderer.setSize(el.clientWidth, el.clientHeight)
    }
    window.addEventListener('resize', onResize)

    return () => {
      alive = false
      cancelAnimationFrame(raf)
      window.removeEventListener('resize', onResize)
      renderer.dispose()
      if (el.contains(renderer.domElement)) el.removeChild(renderer.domElement)
    }
  }, [model])

  return <div ref={mountRef} style={{ width: '100%', height: '100%' }} />
}

function ModelCard({ model, index }) {
  return (
    <motion.div
      className="rounded-xl overflow-hidden flex flex-col"
      style={{ border: `1px solid ${model.color}22`, background: '#080808' }}
      initial={{ opacity: 0, y: 48 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, margin: '-60px' }}
      transition={{ duration: 0.65, delay: index * 0.11, ease: [0.16, 1, 0.3, 1] }}
    >
      {/* 3D canvas */}
      <div
        style={{
          height: 240,
          background: `radial-gradient(ellipse at 50% 75%, ${model.color}18 0%, #000 68%)`,
        }}
      >
        <ModelCanvas model={model} />
      </div>

      {/* Info */}
      <div className="p-5 flex-1 flex flex-col">
        <p
          className="text-[9px] font-mono tracking-[0.42em] uppercase mb-1.5"
          style={{ color: model.color }}
        >
          {model.subtitle}
        </p>
        <h3 className="text-base font-bold text-white mb-2">{model.title}</h3>
        <p className="text-xs text-gray-500 leading-relaxed flex-1">{model.desc}</p>
        <div
          className="mt-4 h-px rounded-full"
          style={{ background: `linear-gradient(90deg, ${model.color}70, transparent)` }}
        />
      </div>
    </motion.div>
  )
}

export default function ModelShowcase() {
  return (
    <section className="relative py-28 px-6 bg-black overflow-hidden">
      {/* Background radial glow */}
      <div
        className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[600px] rounded-full pointer-events-none"
        style={{
          background: 'radial-gradient(ellipse, rgba(232,0,45,0.05) 0%, transparent 70%)',
        }}
      />

      <div className="relative max-w-6xl mx-auto">
        {/* Section heading */}
        <motion.div
          className="mb-16 text-center"
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
        >
          <p className="section-heading mb-4">In The Garage</p>
          <h2 className="text-4xl md:text-5xl font-black text-white tracking-tight">
            Anatomy of a machine.{' '}
            <span style={{ color: '#E8002D' }}>Rendered in 3D.</span>
          </h2>
          <p className="mt-4 text-gray-500 max-w-lg mx-auto">
            Explore the iconic hardware of Formula 1 — from the centerlock wheel that grips the
            tarmac to the five red lights that ignite every race.
          </p>
        </motion.div>

        {/* 4-column card grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
          {MODELS.map((model, i) => (
            <ModelCard key={model.title} model={model} index={i} />
          ))}
        </div>
      </div>
    </section>
  )
}
