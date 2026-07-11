import { HeadContent, Scripts, createRootRoute } from '@tanstack/react-router'

import appCss from '../styles.css?url'

export const Route = createRootRoute({
  head: () => ({
    meta: [
      {
        charSet: 'utf-8',
      },
      {
        name: 'viewport',
        content: 'width=device-width, initial-scale=1',
      },
      {
        title: 'spotr — workout logging for nerds',
      },
      {
        name: 'description',
        content: 'A keyboard-first, local workout tracker for the terminal. Available for macOS, Linux, and Windows.',
      },
      {
        property: 'og:title',
        content: 'spotr — workout logging for nerds',
      },
      {
        property: 'og:description',
        content: 'Track programs, run workouts, and own your training history without leaving the terminal.',
      },
      {
        property: 'og:image',
        content: 'https://spotr.info/spotr-tui.png',
      },
      {
        property: 'og:url',
        content: 'https://spotr.info/',
      },
      {
        name: 'twitter:card',
        content: 'summary_large_image',
      },
    ],
    links: [
      {
        rel: 'stylesheet',
        href: appCss,
      },
      {
        rel: 'icon',
        type: 'image/svg+xml',
        href: '/favicon.svg',
      },
      {
        rel: 'canonical',
        href: 'https://spotr.info/',
      },
    ],
  }),
  shellComponent: RootDocument,
})

function RootDocument({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <head>
        <HeadContent />
      </head>
      <body className="spotr-shell">
        {children}
        <Scripts />
      </body>
    </html>
  )
}
