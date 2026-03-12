import { useState } from 'react'
import { motion } from 'framer-motion'

const METHODS = [
  {
    id: 'binary',
    label: 'Pre-built binary',
    tag: 'Recommended',
    steps: [
      {
        comment: '# Download from GitHub Releases',
        cmd: 'curl -fsSL https://raw.githubusercontent.com/dk-a-dev/termf1/master/install.sh | bash     ',
      },
      {
        comment: '# Set your Groq API key (free tier works)',
        cmd: 'export GROQ_API_KEY=your_key_here',
      },
      {
        comment: '# Launch',
        cmd: './termf1',
      },
    ],
  },
  {
    id: 'go',
    label: 'go install',
    tag: 'Requires Go 1.22+',
    steps: [
      {
        comment: '# One-liner install',
        cmd: 'go installgithub.com/dk-a-dev/termf1/v2@latest',
      },
      {
        comment: '# Set API key',
        cmd: 'export GROQ_API_KEY=your_key_here',
      },
      {
        comment: '# Run anywhere',
        cmd: 'termf1',
      },
    ],
  },
  {
    id: 'source',
    label: 'Build from source',
    tag: 'Dev setup',
    steps: [
      {
        comment: '# Clone the repo',
        cmd: 'git clone https://github.com/dk-a-dev/termf1 && cd termf1',
      },
      {
        comment: '# Configure',
        cmd: 'cp .env.example .env  # add your GROQ_API_KEY',
      },
      {
        comment: '# Run',
        cmd: 'make run',
      },
    ],
  },
]

function CopyButton({ text }) {
  const [copied, setCopied] = useState(false)
  const copy = () => {
    navigator.clipboard.writeText(text).catch(() => {})
    setCopied(true)
    setTimeout(() => setCopied(false), 1800)
  }
  return (
    <button
      onClick={copy}
      className="text-[10px] font-mono px-2.5 py-1 rounded border border-white/10 text-gray-500
                 hover:border-[#E8002D]/60 hover:text-[#E8002D] transition-all duration-150"
    >
      {copied ? '✓ copied' : 'copy'}
    </button>
  )
}

export default function Install() {
  const [active, setActive] = useState('binary')
  const method = METHODS.find((m) => m.id === active)

  return (
    <section id="install" className="relative py-28 px-6 bg-black">
      {/* Red accent line */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-px h-24 bg-gradient-to-b from-transparent via-[#E8002D] to-transparent" />

      <div className="max-w-3xl mx-auto">
        {/* Heading */}
        <motion.div
          className="mb-12 text-center"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
        >
          <p className="section-heading mb-4">To The Grid</p>
          <h2 className="text-4xl md:text-5xl font-black text-white tracking-tight">
            Get started in <span style={{ color: '#E8002D' }}>seconds</span>
          </h2>
          <p className="mt-4 text-gray-500">
            Requires a free&nbsp;
            <a
              href="https://console.groq.com"
              target="_blank"
              rel="noopener noreferrer"
              className="text-[#E8002D] hover:underline"
            >
              Groq API key
            </a>
            &nbsp;for the AI chat tab.
          </p>
        </motion.div>

        {/* Method tabs */}
        <motion.div
          className="flex gap-2 mb-4"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.4, delay: 0.2 }}
        >
          {METHODS.map((m) => (
            <button
              key={m.id}
              onClick={() => setActive(m.id)}
              className={`px-4 py-2 rounded text-xs font-medium transition-all duration-150
                ${active === m.id
                  ? 'bg-[#E8002D] text-white'
                  : 'text-gray-500 border border-white/10 hover:text-white hover:border-white/25'
                }`}
            >
              {m.label}
              {m.tag && (
                <span className={`ml-2 text-[9px] ${active === m.id ? 'text-red-200' : 'text-gray-600'}`}>
                  {m.tag}
                </span>
              )}
            </button>
          ))}
        </motion.div>

        {/* Terminal window */}
        <motion.div
          className="term-chrome"
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.3 }}
        >
          {/* Chrome bar */}
          <div className="term-titlebar">
            <span className="term-dot" style={{ background: '#FF5F57' }} />
            <span className="term-dot" style={{ background: '#FEBC2E' }} />
            <span className="term-dot" style={{ background: '#28C840' }} />
            <span className="ml-3 text-[10px] font-mono text-gray-500">zsh — ~ — 120×30</span>
          </div>

          {/* Body */}
          <div className="p-6 font-mono text-sm space-y-5">
            {method.steps.map((step, i) => (
              <div key={i} className="space-y-1">
                <div className="text-gray-600 text-xs">{step.comment}</div>
                <div className="flex items-center justify-between group">
                  <div className="flex items-center gap-2 overflow-hidden">
                    <span className="text-[#22C55E] flex-shrink-0">❯</span>
                    <span className="text-gray-100 break-all">{step.cmd}</span>
                    <span className="animate-blink text-gray-600 flex-shrink-0">▌</span>
                  </div>
                  <div className="ml-4 flex-shrink-0 opacity-0 group-hover:opacity-100 transition-opacity">
                    <CopyButton text={step.cmd} />
                  </div>
                </div>
              </div>
            ))}
          </div>
        </motion.div>

        {/* Requirements */}
        <motion.div
          className="mt-8 grid grid-cols-1 sm:grid-cols-3 gap-4"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.4, delay: 0.5 }}
        >
          {[
            { icon: '🖥', title: 'macOS / Linux / Windows', sub: 'All platforms supported' },
            { icon: '🎨', title: 'True-colour terminal', sub: 'iTerm2, Ghostty, kitty, WezTerm' },
            { icon: '🔑', title: 'Groq API key', sub: 'Free tier · groq.com' },
          ].map((r) => (
            <div key={r.title} className="glass rounded-lg p-4 text-center">
              <div className="text-2xl mb-2">{r.icon}</div>
              <div className="text-xs font-medium text-white">{r.title}</div>
              <div className="text-[10px] text-gray-600 mt-0.5">{r.sub}</div>
            </div>
          ))}
        </motion.div>
      </div>
    </section>
  )
}
