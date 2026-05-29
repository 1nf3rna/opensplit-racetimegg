const DEBUG = import.meta.env.DEV;

export type Logger = {
  debug: (message: string, ...args: any[]) => void;
  info: (message: string, ...args: any[]) => void;
  warn: (message: string, ...args: any[]) => void;
  error: (message: string, error?: unknown, ...args: any[]) => void;
};

export function moduleLogger(component: string): Logger {
  return {
    debug(message, ...args) {
      if (!DEBUG) return;

      console.debug(`[DEBUG] ${component}: ${message}`, ...args);
    },

    info(message, ...args) {
      console.log(`[INFO] ${component}: ${message}`, ...args);
    },

    warn(message, ...args) {
      console.warn(`[WARN] ${component}: ${message}`, ...args);
    },

    error(message, error?, ...args) {
      console.error(`[ERROR] ${component}: ${message}`, error, ...args);
    },
  };
}