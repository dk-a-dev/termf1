/**
 * FloatingModel — a naked, transparent Three.js canvas that floats between
 * page sections. No card, no heading, no UI chrome. Just a rotating 3D model
 * drifting in space, revealed as the user scrolls.
 */
import { useEffect, useRef } from 'react'
import * as THREE from 'three'
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js'
import { motion, useScroll, useTransform } from 'framer-motion'

export default function FloatingModel({
  path,
  side = 'right',       // 'left' | 'right'
  size = 320,            // canvas px
  lightHex = 0xe8002d,
  rotSpeed = 0.38,
}) {
  const mountRef = useRef(null)
  const scrollRef = useRef(null)

  useEffect(() => {
    const el = mountRef.current
    if (!el) return

    const renderer = new THREE.WebGLRenderer({ alpha: true, antialias: true })
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
    renderer.setSize(el.clientWidth, el.clientHeight)
    renderer.outputColorSpace = THREE.SRGBColorSpace
    renderer.toneMapping = THREE.ACESFilmicToneMapping
    renderer.toneMappingExposure = 1.2
    el.appendChild(renderer.domElement)

    const scene  = new THREE.Scene()
    const camera = new THREE.PerspectiveCamera(44, 1, 0.1, 100)
    camera.position.set(0, 1.0, 5)
    camera.lookAt(0, 0.2, 0)

    scene.add(new THREE.AmbientLight(0xffffff, 0.4))
    const key = new THREE.DirectionalLight(0xffffff, 1.8)
    key.position.set(5, 8, 4)
    scene.add(key)
    const accent = new THREE.PointLight(lightHex, 16, 18)
    accent.position.set(side === 'right' ? -3 : 3, 2, 3)
    scene.add(accent)
    const rim = new THREE.PointLight(0xaaccff, 2.5, 12)
    rim.position.set(3, -1, 2)
    scene.add(rim)

    const group = new THREE.Group()
    scene.add(group)
    let floatT = Math.random() * Math.PI * 2

    const loader = new GLTFLoader()
    loader.load(
      path,
      (gltf) => {
        const m = gltf.scene
        const box  = new THREE.Box3().setFromObject(m)
        const size = new THREE.Vector3()
        box.getSize(size)
        m.scale.setScalar(2.8 / Math.max(size.x, size.y, size.z))
        m.updateMatrixWorld(true)
        const box2   = new THREE.Box3().setFromObject(m)
        const center = new THREE.Vector3()
        box2.getCenter(center)
        m.position.sub(center)
        group.add(m)
      },
      undefined,
      (err) => console.warn('FloatingModel error:', path, err),
    )

    let raf, alive = true
    const clock = new THREE.Clock()
    const animate = () => {
      if (!alive) return
      raf = requestAnimationFrame(animate)
      const dt = clock.getDelta()
      floatT += dt * 0.9
      group.rotation.y += dt * rotSpeed
      group.position.y = Math.sin(floatT) * 0.14
      renderer.render(scene, camera)
    }
    animate()

    return () => {
      alive = false
      cancelAnimationFrame(raf)
      renderer.dispose()
      if (el.contains(renderer.domElement)) el.removeChild(renderer.domElement)
    }
  }, [path, side, lightHex, rotSpeed])

  // Scroll-driven parallax rotation & drift — gives the feeling that
  // the model rolls / turns as the user scrolls past each section.
  const { scrollYProgress } = useScroll({
    target: scrollRef,
    offset: ['start 80%', 'end 20%'],
  })
  const rotateZ = useTransform(
    scrollYProgress,
    [0, 1],
    side === 'right' ? [0, Math.PI * 3] : [0, -Math.PI * 3],
  )
  const translateY = useTransform(scrollYProgress, [0, 1], [80, -80])

  return (
    <div
      style={{
        display: 'flex',
        justifyContent: side === 'right' ? 'flex-end' : 'flex-start',
        padding: '0 3vw',
        marginTop: '-70px',
        marginBottom: '-70px',
        position: 'relative',
        zIndex: 10,
        pointerEvents: 'none',
      }}
    >
      <motion.div
        ref={scrollRef}
        style={{ width: size, height: size, rotateZ, y: translateY }}
        initial={{ opacity: 0, x: side === 'right' ? 90 : -90, scale: 0.75 }}
        whileInView={{ opacity: 1, x: 0, scale: 1 }}
        viewport={{ once: true, margin: '-80px' }}
        transition={{ duration: 1.1, ease: [0.16, 1, 0.3, 1] }}
      >
        <div ref={mountRef} style={{ width: '100%', height: '100%' }} />
      </motion.div>
    </div>
  )
}
