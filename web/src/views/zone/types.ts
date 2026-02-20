export interface Coordinates {
  x: number
  y: number
}

export interface PaperCardData {
  id: string
  text: string
  x: number
  y: number
  rotation: number
  timestamp: number
  isTyping: boolean
  width: number
  height: number
  stampImage?: string
  stampRotation?: number
  stampPosition?: Coordinates
}
