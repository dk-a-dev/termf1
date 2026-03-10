import Hero from './components/Hero'
import Features from './components/Features'
import Screenshots from './components/Screenshots'
import Install from './components/Install'
import Roadmap from './components/Roadmap'
import Footer from './components/Footer'
import FloatingModel from './components/FloatingModel'

export default function App() {
  return (
    <main className="bg-black min-h-screen">
      <Hero />
      <FloatingModel path="/models/formula_1_2012_wheel.glb"             side="right" lightHex={0xe8002d} />
      <Features />
      <FloatingModel path="/models/michael_schumacher_2002_helmet.glb"   side="left"  lightHex={0x3671c6} />
      <Screenshots />
      <FloatingModel path="/models/mclaren_formula_1_steering_wheel.glb" side="right" lightHex={0xff8800} />
      <Install />
      <FloatingModel path="/models/race_track_props_starting_lights.glb" side="left"  lightHex={0x22c55e} />
      <Footer />
    </main>
  )
}
