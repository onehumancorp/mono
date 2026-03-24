type LogLevel = 'debug' | 'info' | 'warn' | 'error'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const isDev = (import.meta as any).env?.DEV ?? false

export const logger = {
  debug: (msg: string, ...args: unknown[]) => {
    if (isDev) console.debug(`[debug] ${msg}`, ...args)
  },
  info: (msg: string, ...args: unknown[]) => {
    console.info(`[info] ${msg}`, ...args)
  },
  warn: (msg: string, ...args: unknown[]) => {
    console.warn(`[warn] ${msg}`, ...args)
  },
  error: (msg: string, ...args: unknown[]) => {
    console.error(`[error] ${msg}`, ...args)
  },
  log: (level: LogLevel, msg: string, ...args: unknown[]) => {
    logger[level](msg, ...args)
  },
}
