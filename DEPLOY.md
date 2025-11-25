# Deployment Guide

## 1. Database Setup (Neon)
1.  Go to [Neon.tech](https://neon.tech) and sign up.
2.  Create a new project.
3.  Copy the **Connection String** (it looks like `postgres://user:pass@ep-xyz.aws.neon.tech/neondb?sslmode=require`).

## 2. Hosting Setup (Render)
1.  Go to [Render.com](https://render.com) and sign up.
2.  Click **New +** -> **Web Service**.
3.  Connect your GitHub repository.
4.  **Name**: `wine-cellar` (or whatever you like).
5.  **Runtime**: `Docker`.
6.  **Region**: Choose one close to you (e.g., Frankfurt or Oregon).
7.  **Free Tier**: Select "Free".
8.  **Environment Variables**:
    *   Key: `DATABASE_URL`
    *   Value: *(Paste your Neon connection string here)*
    *   Key: `CSRF_AUTH_KEY`
    *   Value: *(A 32-byte random string, e.g., `01234567890123456789012345678901`)*
9.  Click **Create Web Service**.

## 3. Continuous Deployment
*   Render automatically watches your `main` branch.
*   Whenever you push code to GitHub, Render will:
    1.  Pull the new code.
    2.  Build the Docker image.
    3.  Deploy the new version.
*   The GitHub Action in `.github/workflows/ci.yml` will run tests on every push to ensure you don't break anything before deployment.
