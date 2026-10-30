import React from "react";
import ReactDOM from "react-dom/client";
import "./index.css";
import App from "./App";
import { ClerkProvider, useAuth } from '@clerk/react';
import { useEffect } from 'react';
import { setClerkTokenGetter } from './lib/api';

const clerkMode = process.env.REACT_APP_AUTH_PROVIDER === 'clerk';
const publishableKey = process.env.REACT_APP_CLERK_PUBLISHABLE_KEY;

function ClerkApp() {
  const { getToken } = useAuth();
  useEffect(() => {
    setClerkTokenGetter(() => getToken());
    return () => setClerkTokenGetter(null);
  }, [getToken]);
  return <App />;
}

const root = ReactDOM.createRoot(document.getElementById("root")!);
root.render(
  <React.StrictMode>
    {clerkMode ? (publishableKey ? <ClerkProvider publishableKey={publishableKey}><ClerkApp /></ClerkProvider> : <div>Set REACT_APP_CLERK_PUBLISHABLE_KEY to start SplitPay.</div>) : <App />}
  </React.StrictMode>
);
