import admin from 'firebase-admin';
import { readFileSync } from 'fs';

let firebaseApp;

export const initializeFirebase = () => {
  try {
    const serviceAccount = JSON.parse(
      readFileSync(process.env.FIREBASE_SERVICE_ACCOUNT_PATH, 'utf8')
    );

    firebaseApp = admin.initializeApp({
      credential: admin.credential.cert(serviceAccount),
    });

    console.log('✅ Firebase Admin initialized');
  } catch (error) {
    console.error('❌ Firebase initialization error:', error.message);
    console.log('⚠️  Continuing without Firebase (download service account key from Firebase Console)');
  }
};

export const getFirebaseAdmin = () => firebaseApp;
