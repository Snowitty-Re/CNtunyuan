'use client'

import Image, { type ImageProps } from 'next/image'
import { useEffect, useState } from 'react'

type SafeImageProps = Omit<ImageProps, 'src' | 'alt'> & {
  src?: string | null
  alt: string
  fallbackSrc?: string
}

export function SafeImage({
  src,
  alt,
  fallbackSrc = '/default-avatar.svg',
  unoptimized = true,
  onError,
  ...rest
}: SafeImageProps) {
  const [currentSrc, setCurrentSrc] = useState(normalizeSource(src, fallbackSrc))

  useEffect(() => {
    setCurrentSrc(normalizeSource(src, fallbackSrc))
  }, [src, fallbackSrc])

  return (
    <Image
      {...rest}
      alt={alt}
      src={currentSrc}
      unoptimized={unoptimized}
      onError={(event) => {
        if (currentSrc !== fallbackSrc) {
          setCurrentSrc(fallbackSrc)
        }
        onError?.(event)
      }}
    />
  )
}

function normalizeSource(src?: string | null, fallbackSrc?: string) {
  const normalized = String(src || '').trim()
  if (normalized) return normalized
  return fallbackSrc || '/default-avatar.svg'
}
