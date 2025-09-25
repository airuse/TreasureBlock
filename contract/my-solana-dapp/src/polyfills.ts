// Ensure Node-like globals exist before any library code runs
import { Buffer } from 'buffer'

if (!(globalThis as any).global) {
  ;(globalThis as any).global = globalThis
}

if (!(globalThis as any).Buffer) {
  ;(globalThis as any).Buffer = Buffer
}


