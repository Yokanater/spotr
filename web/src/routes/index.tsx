import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/')({ component: Home })

function Home() {
  return (
    <main className="landing" id="top">
      <header className="site-header">
        <a className="wordmark" href="#top" aria-label="spotr home">
          spotr<span className="header-cursor" aria-hidden="true" />
        </a>
        <nav className="nav-links" aria-label="Primary navigation">
          <a href="https://github.com/spotr" target="_blank" rel="noreferrer">github ↗</a>
          <a href="https://github.com/Yokanater" target="_blank" rel="noreferrer">made by yokanater</a>
        </nav>
      </header>

      <section className="hero" aria-labelledby="site-title">
        <div className="hero-copy">
          <h1 id="site-title">workout logging<br />for nerds.</h1>
          <p className="intro">
            Track programs, run workouts, and keep your training history without leaving the terminal.
          </p>

          <div className="install" aria-label="Homebrew install command">
            <span className="prompt">$</span>
            <code>brew install --cask spotr</code>
          </div>

          <div className="actions">
            <a className="primary-action" href="https://github.com/spotr" target="_blank" rel="noreferrer">
              get spotr <span>↗</span>
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
        <span>open source · built in go</span>
      </footer>
    </main>
  )
}
