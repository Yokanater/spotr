# spotr.info

The public website for [spotr](https://github.com/Yokanater/spotr), deployed at
[spotr.info](https://spotr.info).

## Local development

Requires Bun.

```bash
bun install
bun run dev
```

The development server runs on `http://localhost:3000`.

## Verify

```bash
bun run build
```

Vercel project state under `.vercel/` and installed dependencies under
`node_modules/` are local-only and must not be committed.
