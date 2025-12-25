# 🚀 Complete Setup Guide - Expense Tracker with Backend

This guide will help you set up the complete application with Firebase Authentication and MongoDB.

## 📋 Prerequisites

- Node.js (v18 or higher)
- MongoDB installed locally
- Firebase account
- Terminal/Command Line access

---

## Part 1: Backend Setup (30 minutes)

### Step 1: Install MongoDB

**macOS:**
```bash
brew tap mongodb/brew
brew install mongodb-community@7.0
brew services start mongodb-community@7.0
```

**Verify installation:**
```bash
mongosh
# Should show: Connected to: mongodb://localhost:27017
# Type 'exit' to quit
```

### Step 2: Setup Firebase Project

1. Go to [Firebase Console](https://console.firebase.google.com/)
2. Click **"Add project"** or select existing project
3. Enter project name (e.g., "expense-tracker")
4. Disable Google Analytics (optional)
5. Click **"Create project"**

### Step 3: Enable Firebase Authentication

1. In Firebase Console, go to **Build** > **Authentication**
2. Click **"Get started"**
3. Enable **Email/Password** sign-in method
4. Click **"Save"**

### Step 4: Get Firebase Service Account Key

1. Go to **Project Settings** (gear icon) > **Service Accounts**
2. Click **"Generate new private key"**
3. Click **"Generate key"** (downloads JSON file)
4. Rename file to `firebase-service-account.json`
5. Move it to `backend/` directory

### Step 5: Install Backend Dependencies

```bash
cd backend
npm install
```

### Step 6: Start Backend Server

```bash
npm run dev
```

You should see:
```
✅ MongoDB Connected: localhost
✅ Firebase Admin initialized
🚀 Server running on port 5000
```

**Test it:**
```bash
curl http://localhost:5000/health
```

---

## Part 2: Frontend Setup (20 minutes)

### Step 1: Get Firebase Web Config

1. In Firebase Console, go to **Project Settings** > **General**
2. Scroll to **"Your apps"** section
3. Click **Web app** icon (`</>`)
4. Register app (name: "expense-tracker-web")
5. Copy the `firebaseConfig` object

### Step 2: Install Firebase in Frontend

```bash
cd frontend
npm install firebase
```

### Step 3: Create Firebase Config File

Create `frontend/src/config/firebase.js`:
```javascript
import { initializeApp } from 'firebase/app';
import { getAuth } from 'firebase/auth';

const firebaseConfig = {
  apiKey: "YOUR_API_KEY",
  authDomain: "YOUR_PROJECT_ID.firebaseapp.com",
  projectId: "YOUR_PROJECT_ID",
  storageBucket: "YOUR_PROJECT_ID.appspot.com",
  messagingSenderId: "YOUR_MESSAGING_SENDER_ID",
  appId: "YOUR_APP_ID"
};

const app = initializeApp(firebaseConfig);
export const auth = getAuth(app);
```

Replace the values with your Firebase config from Step 1.

### Step 4: Create API Service

Create `frontend/src/lib/api.js`:
```javascript
const API_URL = 'http://localhost:5000/api';

// Get auth token from Firebase
const getAuthToken = async () => {
  const user = auth.currentUser;
  if (user) {
    return await user.getIdToken();
  }
  return null;
};

// API methods
export const api = {
  // Users
  getUserProfile: async () => {
    const token = await getAuthToken();
    const response = await fetch(`${API_URL}/users/profile`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    });
    return response.json();
  },

  // Groups
  getGroups: async () => {
    const token = await getAuthToken();
    const response = await fetch(`${API_URL}/groups`, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    return response.json();
  },

  createGroup: async (groupData) => {
    const token = await getAuthToken();
    const response = await fetch(`${API_URL}/groups`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(groupData),
    });
    return response.json();
  },

  updateGroup: async (groupId, groupData) => {
    const token = await getAuthToken();
    const response = await fetch(`${API_URL}/groups/${groupId}`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(groupData),
    });
    return response.json();
  },

  deleteGroup: async (groupId) => {
    const token = await getAuthToken();
    const response = await fetch(`${API_URL}/groups/${groupId}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token}` },
    });
    return response.json();
  },

  // Expenses
  addExpense: async (groupId, expenseData) => {
    const token = await getAuthToken();
    const response = await fetch(`${API_URL}/groups/${groupId}/expenses`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(expenseData),
    });
    return response.json();
  },

  deleteExpense: async (groupId, expenseId) => {
    const token = await getAuthToken();
    const response = await fetch(`${API_URL}/groups/${groupId}/expenses/${expenseId}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${token}` },
    });
    return response.json();
  },
};
```

### Step 5: Update Login Page

Replace `frontend/src/pages/LoginPage.jsx` to use Firebase Auth:
```javascript
import { signInWithEmailAndPassword, createUserWithEmailAndPassword } from 'firebase/auth';
import { auth } from '../config/firebase';
import { api } from '../lib/api';

// In your login handler:
const handleLogin = async () => {
  try {
    const userCredential = await signInWithEmailAndPassword(auth, email, password);
    // Or for signup: await createUserWithEmailAndPassword(auth, email, password);
    
    // Create/get user profile in backend
    await api.getUserProfile();
    
    toast.success('Logged in successfully!');
    navigate('/dashboard');
  } catch (error) {
    toast.error(error.message);
  }
};
```

---

## Part 3: Testing the Full Stack

### Test Backend

1. Start MongoDB: `brew services start mongodb-community@7.0`
2. Start backend: `cd backend && npm run dev`
3. Check health: `curl http://localhost:5000/health`

### Test Frontend

1. Start frontend: `cd frontend && npm start`
2. Open browser: `http://localhost:3000`
3. Sign up with email/password
4. Create a group
5. Add expenses

### Verify Data in MongoDB

```bash
mongosh expense-tracker
db.users.find().pretty()
db.groups.find().pretty()
```

---

## 🐛 Common Issues

### "MongoDB connection failed"
- Check if MongoDB is running: `brew services list`
- Start it: `brew services start mongodb-community@7.0`

### "Firebase authentication error"
- Verify `firebase-service-account.json` is in `backend/` folder
- Check Firebase config in frontend is correct
- Ensure Email/Password auth is enabled in Firebase Console

### "CORS error"
- Backend should allow `http://localhost:3000` (check `.env`)
- Frontend should call `http://localhost:5000/api` (check `api.js`)

### "Port 5000 already in use"
- Change backend port in `.env`: `PORT=5001`
- Update frontend API URL to match

---

## 📚 Next Steps

1. **Replace all localStorage calls** with API calls in:
   - `Dashboard.jsx`
   - `GroupDetail.jsx`
   - `storage.js` (can delete or keep as fallback)

2. **Add loading states** for API calls

3. **Handle authentication state** with Firebase `onAuthStateChanged`

4. **Add error handling** for network failures

5. **Deploy**:
   - Backend: Railway, Render, or Heroku
   - Frontend: Vercel or Netlify
   - MongoDB: MongoDB Atlas (cloud)

---

## ✅ Checklist

Backend:
- [ ] MongoDB installed and running
- [ ] Backend dependencies installed
- [ ] Firebase service account key downloaded
- [ ] Backend server running on port 5000

Frontend:
- [ ] Firebase SDK installed
- [ ] Firebase config added
- [ ] API service created
- [ ] Login page updated
- [ ] Frontend running on port 3000

Testing:
- [ ] Can sign up new user
- [ ] Can create groups via API
- [ ] Can add expenses via API
- [ ] Data persists in MongoDB

---

## 🎉 You're Done!

Your expense tracker now has:
- ✅ Real authentication with Firebase
- ✅ Persistent storage in MongoDB
- ✅ RESTful API backend
- ✅ Secure token-based authorization

Need help? Check the README files in `backend/` and `frontend/` directories.
