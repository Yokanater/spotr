import { createFileRoute } from '@tanstack/react-router'
import { useState } from 'react'

export const Route = createFileRoute('/')({ component: Home })

function Home() {
  const installCommand = 'brew install --cask Yokanater/tap/spotr'
  const [copied, setCopied] = useState(false)

  async function copyInstallCommand() {
    try {
      await navigator.clipboard.writeText(installCommand)
      setCopied(true)
      window.setTimeout(() => setCopied(false), 1600)
    } catch {
      // The command remains selectable when clipboard access is unavailable.
    }
  }

  return (
    <main className="landing" id="top">
      <header className="site-header">
        <a className="wordmark" href="#top" aria-label="spotr home">
          spotr<span className="header-cursor" aria-hidden="true" />
        </a>
        <nav className="nav-links" aria-label="Primary navigation">
          <a href="https://github.com/Yokanater/spotr" target="_blank" rel="noreferrer">github ↗</a>
          <a href="https://github.com/Yokanater/spotr" target="_blank" rel="noreferrer">made by yokanater</a>
        </nav>
      </header>

      <section className="hero" aria-labelledby="site-title">
        <div className="hero-copy">
          <h1 id="site-title">workout logging<br />for nerds.</h1>
          <p className="intro">
            Track programs, run workouts, and keep your training history without leaving the terminal.
          </p>

          <div className="install">
            <div className="install-command" aria-label="Homebrew install command">
              <span className="prompt">$</span>
              <code>{installCommand}</code>
            </div>
            <button type="button" onClick={copyInstallCommand} aria-live="polite">
              {copied ? 'copied' : 'copy'}
            </button>
          </div>

          <div className="actions">
            <a className="primary-action" href="https://github.com/Yokanater/spotr/releases/latest" target="_blank" rel="noreferrer">
              download spotr <span>↗</span>
            </a>
            <span className="platforms">macOS · Linux · Windows</span>
          </div>
        </div>

        <figure className="app-shot">
          <img
            src="/spotr-tui.png"
            alt="The spotr terminal app home screen, showing navigation, the spotr wordmark, and keyboard shortcuts"
          />
        </figure>
      </section>

      <footer>
        <span>your first workout starts here</span>
        <span>open source · local data · built in go</span>
      </footer>
    </main>
  )
}
