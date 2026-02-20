const stampModules = import.meta.glob('/assets/stamp/*.{png,jpg,jpeg,svg,webp}', {
  eager: true,
  import: 'default',
})

export const STAMP_IMAGES: string[] = Object.values(stampModules) as string[]

export const getRandomStamp = (): string => {
  if (STAMP_IMAGES.length === 0) {
    return ''
  }

  const randomIndex = Math.floor(Math.random() * STAMP_IMAGES.length)
  return STAMP_IMAGES[randomIndex] ?? ''
}
