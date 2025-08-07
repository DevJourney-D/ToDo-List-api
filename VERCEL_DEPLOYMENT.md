# ToDo List API - Vercel Deployment

This guide will help you deploy your Go-based ToDo List API to Vercel as a serverless function.

## 🏗️ Project Structure for Vercel

```
backend/
├── api/
│   └── index.go          # Serverless function entry point
├── config/
├── controllers/
├── handlers/
├── middleware/
├── models/
├── repositories/
├── services/
├── utils/
├── vercel.json           # Vercel configuration
├── .vercelignore         # Files to ignore during deployment
├── deploy.sh             # Deployment script
└── main.go               # Original main file (not used in Vercel)
```

## 🚀 Deployment Steps

### 1. Install Vercel CLI

```bash
npm install -g vercel
```

### 2. Login to Vercel

```bash
vercel login
```

### 3. Set Environment Variables

Before deploying, set up these environment variables in your Vercel dashboard:

- `SUPABASE_URL` - Your Supabase project URL
- `SUPABASE_ANON_KEY` - Your Supabase anon key
- `DATABASE_URL` - Your PostgreSQL database connection string
- `JWT_SECRET` - Secret key for JWT token generation

### 4. Deploy

Option A: Use the deployment script
```bash
./deploy.sh
```

Option B: Deploy manually
```bash
vercel --prod
```

## 🔧 Configuration

### vercel.json
The `vercel.json` file configures:
- Build settings for Go runtime
- Routing to the serverless function
- Environment variable references

### Environment Variables in Vercel

1. Go to your Vercel dashboard
2. Select your project
3. Go to Settings → Environment Variables
4. Add each required variable

## 📡 API Endpoints

Once deployed, your API will be available at:
- Base URL: `https://your-project.vercel.app`
- Health check: `https://your-project.vercel.app/health`
- Auth endpoints: `https://your-project.vercel.app/api/v1/register`, `/login`
- Protected endpoints: `https://your-project.vercel.app/api/v1/tasks`, `/habits`, etc.

## 🔒 CORS Configuration

The serverless function is configured to allow requests from any origin (`*`). For production, update the CORS configuration in `api/index.go` to include only your frontend domains.

## 🐛 Troubleshooting

### Common Issues:

1. **Database Connection Timeout**
   - Ensure your DATABASE_URL is correct
   - Check if your database allows connections from Vercel's IP ranges

2. **Environment Variables Not Found**
   - Verify all environment variables are set in Vercel dashboard
   - Check variable names match exactly

3. **Function Timeout**
   - Vercel has a 10-second timeout for serverless functions
   - Optimize database queries for better performance

### Logs

View deployment and runtime logs:
```bash
vercel logs [deployment-url]
```

## 📊 Performance Considerations

- Serverless functions have cold start delays
- Database connections are created on each request
- Consider implementing connection pooling for high traffic
- Use Vercel's Edge Functions for better performance if needed

## 🔄 Updates

To update your deployment:
1. Make changes to your code
2. Commit to Git
3. Run `vercel --prod` again

Vercel will automatically deploy from your Git repository if connected.

## 📝 Notes

- The original `main.go` file is not used in Vercel deployment
- All routing is handled through the `api/index.go` serverless function
- Static file serving is not needed for this API-only deployment
