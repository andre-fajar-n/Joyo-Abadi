# 🚂 Railway Deployment Guide

This guide will help you deploy your Joyo Abadi Go Fiber application to Railway.

## Prerequisites

1. [Railway CLI](https://docs.railway.app/develop/cli) installed
2. Railway account created
3. Git repository pushed to GitHub/GitLab

## Quick Deployment Steps

### 1. Install Railway CLI
```bash
npm install -g @railway/cli
```

### 2. Login to Railway
```bash
railway login
```

### 3. Initialize Railway Project
```bash
railway init
```

### 4. Add PostgreSQL Database
```bash
railway add postgresql
```

### 5. Set Environment Variables
```bash
# Railway will automatically set DATABASE_URL for PostgreSQL
# You may want to set additional variables:
railway variables set LOG_LEVEL=info
railway variables set ENV=production
```

### 6. Deploy
```bash
railway up
```

## Environment Variables

Railway will automatically provide:
- `DATABASE_URL` - PostgreSQL connection string
- `PORT` - Application port
- `RAILWAY_ENVIRONMENT` - Set to "production"

Optional variables you can set:
- `LOG_LEVEL` - Set to "debug", "info", "warn", or "error"
- `ENV` - Set to "production" for production environment

## Database Configuration

The application automatically detects Railway's `DATABASE_URL` environment variable and uses it for database connection. No additional configuration needed!

## Health Check

Railway will automatically health check your application at the root path `/`.

## Custom Domain

After deployment, you can add a custom domain in the Railway dashboard.

## Monitoring

Check your application logs with:
```bash
railway logs
```

## Troubleshooting

1. **Build fails**: Check that all dependencies are in `go.mod`
2. **Database connection fails**: Ensure PostgreSQL service is added
3. **Port issues**: Railway automatically sets the PORT environment variable
4. **Template loading issues**: Templates are copied to the container in the Dockerfile

## Files Added for Railway

- `railway.toml` - Railway configuration
- `.railwayignore` - Files to exclude from deployment
- Updated `Dockerfile` - Optimized for Railway
- Updated `main.go` - DATABASE_URL support
- Updated `utils/session.go` - Production-ready session config
