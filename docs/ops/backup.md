# Postgres & Redis Backup Strategy

To ensure zero data loss and enable disaster recovery in production, the following backup strategy must be configured.

## 1. PostgreSQL (GCP Cloud SQL) Backups

We use managed GCP Cloud SQL for PostgreSQL. The backup configuration is fully managed by GCP and should be configured as follows:

### Automated Backups
- **Frequency**: Daily automated backups.
- **Retention**: Retain backups for at least **30 days**.
- **Backup Window**: Scheduled during low-traffic periods (e.g., 02:00 - 06:00 UTC).
- **Location**: Multi-regional storage to protect against regional outages.

### Point-in-Time Recovery (PITR)
- **Status**: Enabled.
- **Write-Ahead Logging (WAL)**: Enabled with a minimum retention of **7 days**.
- **Usage**: Allows restoring the database to its exact state at any given microsecond within the WAL retention window. Useful for recovering from accidental data deletion or corruption.

### Recovery Runbook
To restore a backup or perform PITR via the Google Cloud Console or CLI:

```bash
# Restore database to a specific point in time
gcloud sql instances restore [INSTANCE_NAME] \
    --restore-instance=[SOURCE_INSTANCE_NAME] \
    --backup-id=[BACKUP_ID] \
    --point-in-time="2026-07-12T10:00:00Z"
```

---

## 2. Redis Cache & Quotas Persistence

Redis stores transient cache data and rate limiter states, but also critical **monthly usage limits and API quotas**. Therefore, standard persistence must be enabled.

### Persistence Configuration (`redis.conf`)
- **AOF (Append Only File)**: Enabled with `appendfsync everysec` for maximum durability with minimal performance overhead.
- **RDB (Snapshotting)**: Enabled as a fallback.
  ```conf
  save 900 1      # Save if 1 key changed in 15 minutes
  save 300 10     # Save if 10 keys changed in 5 minutes
  save 60 10000   # Save if 10,000 keys changed in 1 minute
  ```
