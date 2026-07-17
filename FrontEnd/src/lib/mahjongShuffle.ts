export const mahjongShuffleAlgorithm = 'FY-XorShift32-v1'

export function shuffleMahjongWall<T>(input: readonly T[], seed: number): T[] {
  const values = [...input]
  let state = seed >>> 0
  if (state === 0) state = 0x6d2b79f5
  const next = () => {
    state ^= state << 13
    state ^= state >>> 17
    state ^= state << 5
    state >>>= 0
    return state
  }
  for (let i = values.length - 1; i > 0; i -= 1) {
    const j = next() % (i + 1)
    ;[values[i], values[j]] = [values[j], values[i]]
  }
  return values
}
