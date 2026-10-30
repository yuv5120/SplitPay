import { initializeApp } from "firebase/app";
import { getAuth, Auth } from "firebase/auth";

// Firebase is retained only for rollback. Clerk builds do not need Firebase keys.
const clerkMode = process.env.REACT_APP_AUTH_PROVIDER === 'clerk';
const firebaseConfig = {
  apiKey: process.env.REACT_APP_FIREBASE_API_KEY || (clerkMode ? 'unused-in-clerk-mode' : ''),
  authDomain: process.env.REACT_APP_FIREBASE_AUTH_DOMAIN || (clerkMode ? 'unused.invalid' : ''),
  projectId: process.env.REACT_APP_FIREBASE_PROJECT_ID || (clerkMode ? 'unused' : ''),
  storageBucket: process.env.REACT_APP_FIREBASE_STORAGE_BUCKET,
  messagingSenderId: process.env.REACT_APP_FIREBASE_MESSAGING_SENDER_ID,
  appId: process.env.REACT_APP_FIREBASE_APP_ID || (clerkMode ? 'unused' : ''),
  measurementId: process.env.REACT_APP_FIREBASE_MEASUREMENT_ID,
};

// Optional: basic sanity check to help during local setup
if (!clerkMode && (!firebaseConfig.apiKey || !firebaseConfig.authDomain || !firebaseConfig.projectId || !firebaseConfig.appId)) {
  // eslint-disable-next-line no-console
  console.error(
    "Missing Firebase env config. Please create frontend/.env with REACT_APP_FIREBASE_* values (see frontend/.env.example)."
  );
}

const app = initializeApp(firebaseConfig);
export const auth: Auth = getAuth(app);
