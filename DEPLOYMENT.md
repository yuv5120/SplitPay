# Vercel Deployment Guide for Split-It

## 🚀 Quick Fix Summary

The deployment was failing because Vercel didn't know how to handle the monorepo structure (Go backend + React frontend). I've added the necessary configuration files.

## ✅ Files Added

1. **`vercel.json`** - Tells Vercel how to build and deploy the frontend
2. **`.vercelignore`** - Excludes the Go backend from deployment

## 📋 Deployment Steps

### Option 1: Deploy via Vercel Dashboard (Recommended)

1. **Go to [Vercel Dashboard](https://vercel.com/dashboard)**

2. **Import your GitHub repository:**
   - Click "Add New" → "Project"
   - Select your `yuv5120/SplitPay` repository
   - Click "Import"

3. **Configure the project:**
   - **Framework Preset:** Create React App
   - **Root Directory:** Leave as `./` (root)
   - **Build Command:** `cd frontend && npm install && npm run build`
   - **Output Directory:** `frontend/build`
   - **Install Command:** `cd frontend && npm install`

4. **Add Environment Variables:**
   Click "Environment Variables" and add these:
   
   ```
   REACT_APP_API_URL=<your-backend-url>
   REACT_APP_FIREBASE_API_KEY=AIzaSyDjFbd6B0DwINx6bsFAkW2cBf7xBBIz450
   REACT_APP_FIREBASE_AUTH_DOMAIN=splitease-4b04c.firebaseapp.com
   REACT_APP_FIREBASE_PROJECT_ID=splitease-4b04c
   REACT_APP_FIREBASE_STORAGE_BUCKET=splitease-4b04c.firebasestorage.app
   REACT_APP_FIREBASE_MESSAGING_SENDER_ID=270289445841
   REACT_APP_FIREBASE_APP_ID=1:270289445841:web:746938b84e1a42b3c8f6f0
   REACT_APP_FIREBASE_MEASUREMENT_ID=G-CMP8LTM288
   ```

5. **Click "Deploy"**

### Option 2: Deploy via Vercel CLI

```bash
# Install Vercel CLI
npm i -g vercel

# Login to Vercel
vercel login

# Deploy from project root
vercel

# Follow the prompts:
# - Set up and deploy? Yes
# - Which scope? Your account
# - Link to existing project? No (first time) or Yes (subsequent)
# - What's your project's name? split-it
# - In which directory is your code located? ./

# For production deployment
vercel --prod
```

## 🔧 Backend Deployment

Your **Go backend** needs to be deployed separately. Here are your options:

### Option A: Railway.app (Easiest for Go)

1. Go to [Railway.app](https://railway.app/)
2. Click "New Project" → "Deploy from GitHub repo"
3. Select your repository
4. Railway will auto-detect the Go backend
5. Add environment variables in Railway dashboard
6. Get your backend URL and update `REACT_APP_API_URL` in Vercel

### Option B: Render.com

1. Go to [Render.com](https://render.com/)
2. Click "New" → "Web Service"
3. Connect your GitHub repository
4. Configure:
   - **Root Directory:** `backend`
   - **Build Command:** `go build -o main .`
   - **Start Command:** `./main`
5. Add environment variables
6. Deploy

### Option C: Google Cloud Run

```bash
# From backend directory
cd backend

# Build Docker image
docker build -t gcr.io/YOUR_PROJECT_ID/split-it-backend .

# Push to Google Container Registry
docker push gcr.io/YOUR_PROJECT_ID/split-it-backend

# Deploy to Cloud Run
gcloud run deploy split-it-backend \
  --image gcr.io/YOUR_PROJECT_ID/split-it-backend \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated
```

## 🐛 Common Deployment Issues

### Issue 1: Build fails with "command not found: craco"

**Solution:** The `vercel.json` now includes proper install commands.

### Issue 2: Environment variables not working

**Solution:** 
- Make sure all `REACT_APP_*` variables are set in Vercel dashboard
- Redeploy after adding variables

### Issue 3: 404 on page refresh

**Solution:** The `vercel.json` includes rewrites to handle client-side routing.

### Issue 4: CORS errors after deployment

**Solution:** Update your backend's `CLIENT_URL` environment variable to include your Vercel domain:
```
CLIENT_URL=https://your-app.vercel.app
```

## 📝 Post-Deployment Checklist

- [ ] Frontend deployed successfully on Vercel
- [ ] Backend deployed on Railway/Render/Cloud Run
- [ ] Updated `REACT_APP_API_URL` in Vercel to point to backend
- [ ] Updated `CLIENT_URL` in backend to point to Vercel frontend
- [ ] Tested authentication flow
- [ ] Tested API endpoints
- [ ] Verified Firebase configuration
- [ ] Checked browser console for errors

## 🔗 Useful Links

- [Vercel Documentation](https://vercel.com/docs)
- [Railway Documentation](https://docs.railway.app/)
- [Render Documentation](https://render.com/docs)

## 💡 Tips

1. **Use Vercel for Frontend only** - Vercel is optimized for static sites and serverless functions
2. **Use Railway/Render for Go Backend** - These platforms have better support for long-running Go services
3. **Set up automatic deployments** - Connect your GitHub repo so every push triggers a deployment
4. **Use environment-specific configs** - Different URLs for development vs production

---

Need help? Check the deployment logs in your Vercel dashboard for specific error messages.
