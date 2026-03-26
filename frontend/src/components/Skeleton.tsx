import styles from './Skeleton.module.css'

interface Props {
  width?: string | number
  height?: string | number
  borderRadius?: string
}

export default function Skeleton({ width = '100%', height = 16, borderRadius }: Props) {
  return (
    <div
      className={styles.skeleton}
      style={{
        width,
        height,
        borderRadius: borderRadius ?? 'var(--radius-sm)',
      }}
    />
  )
}
