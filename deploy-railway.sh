#!/bin/bash

# Railway Deployment Script for Joyo Abadi

echo "🚂 Deploying Joyo Abadi to Railway..."

# Check if Railway CLI is installed
if ! command -v railway &> /dev/null; then
    echo "❌ Railway CLI is not installed. Please install it first:"
    echo "npm install -g @railway/cli"
    exit 1
fi

# Check if user is logged in
if ! railway whoami &> /dev/null; then
    echo "🔐 Please login to Railway first:"
    echo "railway login"
    exit 1
fi

# Build the application locally to check for errors
echo "🔨 Building application locally..."
if ! go build -o main .; then
    echo "❌ Build failed. Please fix the errors before deploying."
    exit 1
fi

echo "✅ Local build successful!"

# Clean up local build
rm -f main

# Check if this is a new project
if [ ! -f "railway.toml" ]; then
    echo "❌ railway.toml not found. This script should be run from the project root."
    exit 1
fi

# Deploy to Railway
echo "🚀 Deploying to Railway..."
railway up

echo "✅ Deployment initiated!"
echo ""
echo "📋 Next steps:"
echo "1. Check deployment status: railway status"
echo "2. View logs: railway logs"
echo "3. Open in browser: railway open"
echo "4. Add custom domain in Railway dashboard if needed"
echo ""
echo "🔗 Don't forget to add a PostgreSQL database if you haven't already:"
echo "railway add postgresql"
