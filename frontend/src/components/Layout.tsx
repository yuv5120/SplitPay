import React, { FC, ReactNode } from 'react';
import { useNavigate } from 'react-router-dom';
import { Wallet } from 'lucide-react';
import { auth } from '../lib/firebase';
import type { User } from 'firebase/auth';
import { useUser } from '@clerk/react';

interface LayoutProps {
  children: ReactNode;
}

export const Layout: FC<LayoutProps> = ({ children }: LayoutProps) => {
  if (process.env.REACT_APP_AUTH_PROVIDER === 'clerk') return <ClerkLayout>{children}</ClerkLayout>;
  return <FirebaseLayout>{children}</FirebaseLayout>;
};

const ClerkLayout: FC<LayoutProps> = ({ children }) => {
  const { user } = useUser();
  return <LayoutFrame name={user?.fullName || user?.firstName || user?.primaryEmailAddress?.emailAddress || ''} email={user?.primaryEmailAddress?.emailAddress} >{children}</LayoutFrame>;
};

const FirebaseLayout: FC<LayoutProps> = ({ children }) => {
  const [user, setUser] = React.useState<User | null>(null);

  React.useEffect(() => {
    const unsubscribe = auth.onAuthStateChanged((firebaseUser) => {
      setUser(firebaseUser);
    });
    return () => unsubscribe();
  }, []);

  return <LayoutFrame name={user?.displayName || user?.email || ''} email={user?.email || undefined}>{children}</LayoutFrame>;
};

const LayoutFrame: FC<LayoutProps & {name: string; email?: string}> = ({ children, name, email }) => {
  const navigate = useNavigate();
  return (
    <div className="min-h-screen bg-gradient-hero">
      {/* Header */}
      <header className="glass border-b sticky top-0 z-40">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div
              className="flex items-center gap-2 cursor-pointer"
              onClick={() => navigate('/dashboard')}
            >
              <div className="w-10 h-10 rounded-xl gradient-primary flex items-center justify-center shadow-glow">
                <Wallet className="w-6 h-6 text-primary-foreground" />
              </div>
              <div>
                <h1 className="text-xl font-bold gradient-text">SplitPay</h1>
                <p className="text-xs text-muted-foreground">Smart expense splitting</p>
              </div>
            </div>

            {name && (
              <div
                className="flex items-center gap-2 px-4 py-2 rounded-lg bg-card/50 border cursor-pointer hover:bg-card/70 transition-all"
                onClick={() => navigate('/profile')}
              >
                <div className="w-8 h-8 rounded-full gradient-primary flex items-center justify-center text-primary-foreground font-semibold text-sm">
                  {name.charAt(0).toUpperCase() || email?.charAt(0).toUpperCase()}
                </div>
                <div className="hidden md:block">
                  <p className="text-sm font-medium">{name || 'User'}</p>
                </div>
              </div>
            )}
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-8">
        {children}
      </main>

      {/* Footer */}
      <footer className="mt-auto border-t glass">
        <div className="container mx-auto px-4 py-6">
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            <p className="text-sm text-muted-foreground">
              © {new Date().getFullYear()} SplitPay. Split expenses with friends, simplified.
            </p>
            <div className="flex gap-4 text-sm text-muted-foreground">
              <a href="#" className="hover:text-primary transition-colors">Privacy</a>
              <a href="#" className="hover:text-primary transition-colors">Terms</a>
              <a href="#" className="hover:text-primary transition-colors">Support</a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
};
