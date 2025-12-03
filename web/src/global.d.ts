// Google Model Viewer type declarations
declare namespace JSX {
    interface IntrinsicElements {
        'model-viewer': ModelViewerJSX & React.DetailedHTMLProps<React.HTMLAttributes<HTMLElement>, HTMLElement>
    }
}

interface ModelViewerJSX {
    src?: string
    alt?: string
    poster?: string
    loading?: 'auto' | 'lazy' | 'eager'
    reveal?: 'auto' | 'interaction' | 'manual'
    'auto-rotate'?: boolean | string
    'camera-controls'?: boolean | string
    'shadow-intensity'?: string
    exposure?: string
    'environment-image'?: string
    'skybox-image'?: string
    ar?: boolean | string
    'ar-modes'?: string
    'ar-scale'?: string
    'camera-orbit'?: string
    'field-of-view'?: string
    'max-camera-orbit'?: string
    'min-camera-orbit'?: string
    'max-field-of-view'?: string
    'min-field-of-view'?: string
    'interaction-prompt'?: string
    'interaction-prompt-style'?: string
    'interaction-prompt-threshold'?: string
    'touch-action'?: string
    'disable-zoom'?: boolean | string
    'disable-pan'?: boolean | string
    'disable-tap'?: boolean | string
    'interpolation-decay'?: string
    'orbit-sensitivity'?: string
    'zoom-sensitivity'?: string
    'pan-sensitivity'?: string
}

// Vue component declaration for model-viewer
declare module 'vue' {
    export interface GlobalComponents {
        'model-viewer': ModelViewerJSX
    }
}
