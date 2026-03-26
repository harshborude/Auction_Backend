import styles from './StatusBadge.module.css'

type Status = 'ACTIVE' | 'SCHEDULED' | 'ENDED' | 'CANCELLED'

export default function StatusBadge({ status }: { status: Status }) {
  return <span className={`${styles.badge} ${styles[status.toLowerCase()]}`}>{status}</span>
}
