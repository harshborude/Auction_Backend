import { Link } from 'react-router-dom'
import styles from './Home.module.css'

export default function Home() {
  return (
    <>
      <section className={styles.hero}>
        <p className={styles.eyebrow}>Live Auction Platform</p>
        <h1 className={styles.heroTitle}>
          Bid. Win.<br />Own it.
        </h1>
        <p className={styles.heroSub}>
          Real-time bidding on unique items. Every auction ends with a winner — it could be you.
        </p>
        <Link to="/auctions" className={styles.heroCta}>
          Browse auctions →
        </Link>
      </section>

      <hr className={styles.divider} />

      <section className={styles.how}>
        <p className={styles.sectionTitle}>How it works</p>
        <div className={styles.steps}>
          <div className={styles.step}>
            <div className={styles.stepNum}>1</div>
            <div className={styles.stepTitle}>Create an account</div>
            <p className={styles.stepDesc}>
              Register and get credits assigned by an admin to start bidding.
            </p>
          </div>
          <div className={styles.step}>
            <div className={styles.stepNum}>2</div>
            <div className={styles.stepTitle}>Browse auctions</div>
            <p className={styles.stepDesc}>
              Explore active auctions with live countdowns and current bids.
            </p>
          </div>
          <div className={styles.step}>
            <div className={styles.stepNum}>3</div>
            <div className={styles.stepTitle}>Place your bid</div>
            <p className={styles.stepDesc}>
              Outbid others in real time. Credits are reserved — not deducted — until the auction ends.
            </p>
          </div>
          <div className={styles.step}>
            <div className={styles.stepNum}>4</div>
            <div className={styles.stepTitle}>Win &amp; collect</div>
            <p className={styles.stepDesc}>
              Highest bidder when the timer hits zero wins. Anti-sniping ensures fair endings.
            </p>
          </div>
        </div>
      </section>
    </>
  )
}
