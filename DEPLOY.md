# Deployment Guide

## 1. Database Setup (Neon)
1.  Go to [Neon.tech](https://neon.tech) and sign up.
2.  Create a new project.
3.  Copy the **Connection String** (it looks like `postgres://user:pass@ep-xyz.aws.neon.tech/neondb?sslmode=require`).

## 2. Image Storage Setup (Cloudflare R2) - Optional
1.  Go to [Cloudflare Dashboard](https://dash.cloudflare.com) and navigate to **R2 Object Storage**.
2.  Create a new bucket (e.g., `winetrackr-images`).
3.  Go to **R2** -> **Overview** -> **Manage R2 API Tokens** -> **Create API Token**.
4.  Create a token with **Object Read & Write** permissions for your bucket.
5.  Note down:
    *   **Account ID** (found in the URL or R2 overview)
    *   **Access Key ID**
    *   **Secret Access Key**
6.  (Optional) Set up a custom domain for your bucket under **Settings** -> **Public Access** -> **Custom Domains**.

## 3. Hosting Setup (Render)
1.  Go to [Render.com](https://render.com) and sign up.
2.  Click **New +** -> **Web Service**.
3.  Connect your GitHub repository.
4.  **Name**: `wine-cellar` (or whatever you like).
5.  **Runtime**: `Docker`.
6.  **Region**: Choose one close to you (e.g., Frankfurt or Oregon).
7.  **Instance Type**:
    *   Select **"Free"** for hobby projects (spins down after inactivity).
    *   Select **"Starter"** or higher for production (stays online 24/7).
8.  **Environment Variables**:
    *   Key: `DATABASE_URL`
    *   Value: *(Paste your Neon connection string here)*
    *   Key: `CSRF_AUTH_KEY`
    *   Value: *(A 32-byte random string, e.g., `01234567890123456789012345678901`)*
    *   Key: `SESSION_SECRET`
    *   Value: *(A random string for session encryption)*
    *   Key: `R2_ACCOUNT_ID`
    *   Value: *(Your Cloudflare Account ID)*
    *   Key: `R2_ACCESS_KEY_ID`
    *   Value: *(Your R2 API Access Key ID)*
    *   Key: `R2_SECRET_ACCESS_KEY`
    *   Value: *(Your R2 API Secret Access Key)*
    *   Key: `R2_BUCKET_NAME`
    *   Value: *(Your bucket name, e.g., `winetrackr-images`)*
    *   Key: `R2_PUBLIC_URL`
    *   Value: *(Your custom domain URL, e.g., `https://images.yourdomain.com`)*
9.  Click **Create Web Service**.

> **Note**: R2 configuration is optional. If not configured, images will be stored as base64 in the database (works but uses more database storage).

## 4. Continuous Deployment
*   Render automatically watches your `main` branch.
*   Whenever you push code to GitHub, Render will:
    1.  Pull the new code.
    2.  Build the Docker image.
    3.  Deploy the new version.
*   The GitHub Action in `.github/workflows/ci.yml` will run tests on every push to ensure you don't break anything before deployment.
