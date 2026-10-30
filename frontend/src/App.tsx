import { FC } from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { Toaster } from "./components/ui/sonner";
import { LandingPage } from "./pages/LandingPage";
import { LoginPage } from "./pages/LoginPage";
import { Dashboard } from "./pages/Dashboard";
import { GroupDetail } from "./pages/GroupDetail";
import { Profile } from "./pages/Profile";
import { ProtectedRoute } from "./components/ProtectedRoute";
import { JoinGroup } from "./pages/JoinGroup";
import { SignIn, SignUp } from '@clerk/react';

const clerkMode = process.env.REACT_APP_AUTH_PROVIDER === 'clerk';
const ClerkLogin = () => <div className="min-h-screen flex items-center justify-center bg-gradient-hero p-4"><SignIn routing="hash" signUpUrl="/sign-up" fallbackRedirectUrl="/dashboard" /></div>;
const ClerkSignup = () => <div className="min-h-screen flex items-center justify-center bg-gradient-hero p-4"><SignUp routing="hash" signInUrl="/login" fallbackRedirectUrl="/dashboard" /></div>;

const App: FC = () => {
  return (
    <BrowserRouter>
      <Toaster position="top-right" />
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/login" element={clerkMode ? <ClerkLogin /> : <LoginPage />} />
        {clerkMode && <Route path="/sign-up" element={<ClerkSignup />} />}
        <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
        <Route path="/profile" element={<ProtectedRoute><Profile /></ProtectedRoute>} />
        <Route path="/group/:groupId" element={<ProtectedRoute><GroupDetail /></ProtectedRoute>} />
        <Route path="/join/:token" element={<ProtectedRoute><JoinGroup /></ProtectedRoute>} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
};

export default App;
