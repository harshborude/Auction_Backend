import styles from './Pagination.module.css'

interface Props {
  page: number
  hasMore: boolean
  onPrev: () => void
  onNext: () => void
}

export default function Pagination({ page, hasMore, onPrev, onNext }: Props) {
  return (
    <div className={styles.pagination}>
      <button className="btn btn-ghost btn-sm" onClick={onPrev} disabled={page <= 1}>
        ← Prev
      </button>
      <span className={styles.page}>Page {page}</span>
      <button className="btn btn-ghost btn-sm" onClick={onNext} disabled={!hasMore}>
        Next →
      </button>
    </div>
  )
}
