#!/bin/bash

# Vercel Deployment Script for ToDo List API

echo "🚀 Deploying ToDo List API to Vercel..."

# Check if Vercel CLI is installed
if ! command -v vercel &> /dev/null; then
    echo "❌ Vercel CLI is not installed. Installing..."
    npm install -g vercel
fi

# Login to Vercel (if not already logged in)
echo "🔑 Checking Vercel authentication..."
vercel whoami || vercel login

# Set environment variables
echo "🔧 Setting up environment variables..."
echo "Please make sure to set these environment variables in your Vercel dashboard:"
echo "- SUPABASE_URL"
echo "- SUPABASE_ANON_KEY" 
echo "- DATABASE_URL"
echo "- JWT_SECRET"

# Deploy to Vercel
echo "📦 Deploying to Vercel..."
vercel --prod

echo "✅ Deployment completed!"
echo "🌐 Your API should be available at the URL provided by Vercel"
echo ""
echo "📝 Next steps:"
echo "1. Set environment variables in Vercel dashboard"
echo "2. Test your API endpoints"
echo "3. Update CORS origins if needed"
