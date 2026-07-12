# GCP Secret Manager Configuration

All production secrets (database credentials, JWT keys, SMTP credentials, Google Safe Browsing API keys) must be stored in **GCP Secret Manager** and injected into the Cloud Run container at runtime. Do not commit `.env` files to production.

## 1. Required Secrets list

The following secrets must be registered in Secret Manager:

| Secret Name | Purpose | Example / Format |
|---|---|---|
| `DB_PASSWORD` | PostgreSQL password | High-entropy random string |
| `JWT_SECRET` | Signing key for Access/Refresh tokens | Random 32+ character string |
| `SMTP_PASSWORD` | Mail sender authentication key | Gmail App Password or SendGrid key |
| `GOOGLE_SAFE_BROWSING_API_KEY` | Phishing check integration key | Google API credential key |

---

## 2. Secrets Mount Configuration for Cloud Run

Instead of injecting secrets as plaintext environment variables, mount them to the environment via Cloud Run settings to prevent log leaks:

```bash
# Deploy to Cloud Run mounting secrets directly to environment variables
gcloud run deploy url-shortener-service \
    --image gcr.io/[PROJECT_ID]/url-shortener:latest \
    --update-secrets=DB_PASSWORD=DB_PASSWORD:latest,\
JWT_SECRET=JWT_SECRET:latest,\
SMTP_PASSWORD=SMTP_PASSWORD:latest,\
GOOGLE_SAFE_BROWSING_API_KEY=GOOGLE_SAFE_BROWSING_API_KEY:latest \
    --update-env-vars=DB_HOST="[POSTGRES_INSTANCE_IP]",\
DB_PORT="5432",\
DB_USER="postgres",\
DB_NAME="shortener",\
SMTP_HOST="smtp.gmail.com",\
SMTP_PORT="587",\
SMTP_FROM="no-reply@example.com"
```

## 3. Secret Rotation Strategy

- **Frequency**: Every 90 days.
- **Procedure**:
  1. Add a new version of the secret in GCP Secret Manager.
  2. Deploy a new revision of the Cloud Run service referencing the `latest` or specific version tag.
  3. Ensure no downtime by keeping the previous version active until the transition is complete.
