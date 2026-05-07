/**
 * URL builder — concatenates base + API path + path, then appends a query
 * string. Pure function; no I/O.
 *
 * @internal
 */
export function buildUrl(args: {
  readonly baseUrl: string
  readonly apiPath: string
  readonly path: string
  readonly query: Record<string, string | number | boolean | undefined> | undefined
}): string {
  const base = args.baseUrl.replace(/\/+$/u, '')
  const normalizedPath = args.path.startsWith('/') ? args.path : `/${args.path}`
  let url = `${base}${args.apiPath}${normalizedPath}`

  if (args.query !== undefined) {
    const params = new URLSearchParams()
    for (const [key, value] of Object.entries(args.query)) {
      if (value === undefined) continue
      params.append(key, String(value))
    }
    const qs = params.toString()
    if (qs.length > 0) url += `?${qs}`
  }

  return url
}
