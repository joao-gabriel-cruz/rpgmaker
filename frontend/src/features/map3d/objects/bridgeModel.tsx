import { Box } from './parts'
import { C } from './palette'

// Deck top sits at local y=0 (the placement lifts it to the deck height); planks run across x.
export function Bridge() {
  const planks = Array.from({ length: 6 }, (_, i) => -0.42 + i * 0.168)
  return (
    <>
      {planks.map((x, i) => (
        <Box key={x} size={[0.15, 0.07, 0.98]} pos={[x, -0.035, 0]} rot={[0, (i % 3 - 1) * 0.015, 0]} color={i % 2 ? C.wood : '#7b5236'} />
      ))}
      {[-0.4, 0.4].map((z) => (
        <Box key={z} size={[1.02, 0.28, 0.1]} pos={[0, -0.2, z]} color={C.darkWood} />
      ))}
      {[-0.46, 0.46].flatMap((z) => [-0.45, 0.45].map((x) => (
        <Box key={`${x}${z}`} size={[0.07, 0.42, 0.07]} pos={[x, 0.21, z]} color={C.darkWood} />
      )))}
      {[-0.46, 0.46].map((z) => (
        <Box key={z} size={[1, 0.05, 0.05]} pos={[0, 0.38, z]} color={C.wood} />
      ))}
    </>
  )
}
