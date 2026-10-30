import { useEffect, useState, ReactNode } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { onAuthStateChanged, User } from 'firebase/auth';
import { auth } from '../lib/firebase';
import { useAuth } from '@clerk/react';

export function ProtectedRoute({ children }: { children: ReactNode }) {
  if (process.env.REACT_APP_AUTH_PROVIDER === 'clerk') return <ClerkProtectedRoute>{children}</ClerkProtectedRoute>;
  return <FirebaseProtectedRoute>{children}</FirebaseProtectedRoute>;
}

function ClerkProtectedRoute({ children }: { children: ReactNode }) {
  const { isLoaded, isSignedIn } = useAuth();
  const location = useLocation();
  if (!isLoaded) return <div className="min-h-screen flex items-center justify-center">Loading your account…</div>;
  if (!isSignedIn) return <Navigate to="/login" state={{ from: location.pathname }} replace />;
  return <>{children}</>;
}

function FirebaseProtectedRoute({ children }: { children: ReactNode }) {
  const location = useLocation();
  const [user, setUser] = useState<User | null>(auth.currentUser);
  const [ready, setReady] = useState(false);

  useEffect(() => onAuthStateChanged(auth, currentUser => {
    setUser(currentUser);
    setReady(true);
  }), []);

  if (!ready) return <div className="min-h-screen flex items-center justify-center">Loading your account…</div>;
  if (!user) return <Navigate to="/login" state={{ from: location.pathname }} replace />;
  return <>{children}</>;
}
