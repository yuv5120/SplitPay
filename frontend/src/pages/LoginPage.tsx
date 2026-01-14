import { useState, FC, FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Wallet, LogIn, UserPlus, Mail, Phone, Eye, EyeOff } from 'lucide-react';
import { auth } from '../lib/firebase';
import { signInWithEmailAndPassword, createUserWithEmailAndPassword, updateProfile, sendEmailVerification, sendPasswordResetEmail } from 'firebase/auth';
import { toast } from 'sonner';
import { motion } from 'framer-motion';
import { api } from '../lib/api';

interface FormData {
  name: string;
  email: string;
  phone: string;
  password: string;
}

export const LoginPage: FC = () => {
  const navigate = useNavigate();
  const [isLogin, setIsLogin] = useState(true);
  const [isForgotPassword, setIsForgotPassword] = useState(false);
  const [emailVerificationSent, setEmailVerificationSent] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [formData, setFormData] = useState<FormData>({
    name: '',
    email: '',
    phone: '',
    password: '',
  });

  const validatePhoneNumber = (phone: string): boolean => {
    // Basic phone validation - at least 10 digits
    const phoneRegex = /^[0-9]{10,15}$/;
    return phoneRegex.test(phone.replace(/[\s\-\(\)]/g, ''));
  };

  const handleResendVerification = async () => {
    try {
      if (auth.currentUser) {
        await sendEmailVerification(auth.currentUser);
        toast.success('Verification email sent! Please check your inbox.');
      }
    } catch (error: any) {
      toast.error('Failed to send verification email: ' + error.message);
    }
  };

  const handleForgotPassword = async (e: FormEvent) => {
    e.preventDefault();
    if (!formData.email) {
      toast.error('Please enter your email address');
      return;
    }
    try {
      await sendPasswordResetEmail(auth, formData.email);
      toast.success('Password reset email sent! Please check your inbox.');
      setIsForgotPassword(false);
      setIsLogin(true);
    } catch (error: any) {
      toast.error(error.message);
    }
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!formData.email || !formData.password) {
      toast.error('Please fill in all fields');
      return;
    }
    if (!isLogin && !formData.name) {
      toast.error('Please enter your name');
      return;
    }
    if (!isLogin && !formData.phone) {
      toast.error('Please enter your phone number');
      return;
    }
    if (!isLogin && !validatePhoneNumber(formData.phone)) {
      toast.error('Please enter a valid phone number (10-15 digits)');
      return;
    }
    try {
      if (isLogin) {
        // Login flow - NO verification email sent here
        const userCredential = await signInWithEmailAndPassword(auth, formData.email, formData.password);

        // Check if email is verified
        if (!userCredential.user.emailVerified) {
          toast.warning('Please verify your email before logging in.');
          setEmailVerificationSent(true);
          return;
        }

        toast.success('Welcome back!');
        navigate('/dashboard');
      } else {
        // Signup flow - verification email ONLY sent here
        const userCredential = await createUserWithEmailAndPassword(auth, formData.email, formData.password);

        // Update profile with name
        await updateProfile(userCredential.user, { displayName: formData.name });

        // Store phone number in backend
        try {
          await api.createOrGetProfile(formData.phone);
        } catch (error) {
          console.error('Failed to store phone number:', error);
          // Continue even if profile creation fails
        }

        // Send email verification
        await sendEmailVerification(userCredential.user);

        toast.success('Account created! Please verify your email.');
        setEmailVerificationSent(true);
      }
    } catch (error: any) {
      toast.error(error.message);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-hero flex items-center justify-center p-4">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="w-full max-w-md"
      >
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center gap-2 mb-4">
            <div className="w-12 h-12 rounded-xl gradient-primary flex items-center justify-center shadow-glow">
              <Wallet className="w-7 h-7 text-primary-foreground" />
            </div>
            <div className="text-left">
              <h1 className="text-2xl font-bold gradient-text">SplitEase</h1>
              <p className="text-sm text-muted-foreground">Smart expense splitting</p>
            </div>
          </div>
        </div>

        <Card className="glass-card shadow-elegant">
          <CardHeader className="space-y-1">
            <CardTitle className="text-3xl text-center">
              {isForgotPassword ? 'Reset Password' : isLogin ? 'Welcome Back' : 'Create Account'}
            </CardTitle>
            <CardDescription className="text-center">
              {isForgotPassword
                ? 'Enter your email to receive a password reset link'
                : isLogin
                  ? 'Sign in to manage your shared expenses'
                  : 'Get started with SplitEase'}
            </CardDescription>
          </CardHeader>

          <CardContent>
            {emailVerificationSent ? (
              <div className="space-y-4">
                <div className="text-center p-6 bg-primary/10 rounded-lg border border-primary/20">
                  <Mail className="w-12 h-12 mx-auto mb-4 text-primary" />
                  <h3 className="text-lg font-semibold mb-2">Verify Your Email</h3>
                  <p className="text-sm text-muted-foreground mb-4">
                    We've sent a verification link to <strong>{formData.email}</strong>.
                    Please check your inbox and click the link to verify your account.
                  </p>
                  <Button
                    onClick={handleResendVerification}
                    variant="outline"
                    className="w-full mb-2"
                  >
                    <Mail className="w-4 h-4 mr-2" />
                    Resend Verification Email
                  </Button>
                  <Button
                    onClick={() => {
                      setEmailVerificationSent(false);
                      setIsLogin(true);
                    }}
                    variant="ghost"
                    className="w-full"
                  >
                    Back to Login
                  </Button>
                </div>
              </div>
            ) : isForgotPassword ? (
              <form onSubmit={handleForgotPassword} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="name@example.com"
                    value={formData.email}
                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                    required
                  />
                </div>
                <Button type="submit" variant="premium" className="w-full" size="lg">
                  <Mail className="w-5 h-5 mr-2" />
                  Send Reset Link
                </Button>
              </form>
            ) : (
              <form onSubmit={handleSubmit} className="space-y-4">
                {!isLogin && (
                  <div className="space-y-2">
                    <Label htmlFor="name">Full Name</Label>
                    <Input
                      id="name"
                      placeholder="Enter your name"
                      value={formData.name}
                      onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                      required={!isLogin}
                    />
                  </div>
                )}

                {!isLogin && (
                  <div className="space-y-2">
                    <Label htmlFor="phone">Phone Number</Label>
                    <div className="relative">
                      <Phone className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                      <Input
                        id="phone"
                        type="tel"
                        placeholder="Enter your phone number"
                        value={formData.phone}
                        onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
                        className="pl-10"
                        required={!isLogin}
                      />
                    </div>
                    <p className="text-xs text-muted-foreground">10-15 digits, no spaces or special characters</p>
                  </div>
                )}

                <div className="space-y-2">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="name@example.com"
                    value={formData.email}
                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                    required
                  />
                </div>

                {!isForgotPassword && (
                  <div className="space-y-2">
                    <Label htmlFor="password">Password</Label>
                    <div className="relative">
                      <Input
                        id="password"
                        type={showPassword ? "text" : "password"}
                        placeholder="Enter your password"
                        value={formData.password}
                        onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                        className="pr-10"
                        required
                      />
                      <button
                        type="button"
                        onClick={() => setShowPassword(!showPassword)}
                        className="absolute right-3 top-1/2 transform -translate-y-1/2 text-muted-foreground hover:text-primary transition-colors"
                      >
                        {showPassword ? (
                          <EyeOff className="w-4 h-4" />
                        ) : (
                          <Eye className="w-4 h-4" />
                        )}
                      </button>
                    </div>
                  </div>
                )}

                <Button type="submit" variant="premium" className="w-full" size="lg">
                  {isForgotPassword ? (
                    <>
                      <Mail className="w-5 h-5 mr-2" />
                      Send Reset Link
                    </>
                  ) : isLogin ? (
                    <>
                      <LogIn className="w-5 h-5 mr-2" />
                      Sign In
                    </>
                  ) : (
                    <>
                      <UserPlus className="w-5 h-5 mr-2" />
                      Create Account
                    </>
                  )}
                </Button>

                {isLogin && !isForgotPassword && (
                  <div className="text-center mt-4">
                    <button
                      type="button"
                      onClick={() => setIsForgotPassword(true)}
                      className="text-sm text-primary hover:underline"
                    >
                      Forgot password?
                    </button>
                  </div>
                )}
              </form>
            )}

            {!emailVerificationSent && (
              <div className="mt-6 text-center">
                <button
                  type="button"
                  onClick={() => {
                    if (isForgotPassword) {
                      setIsForgotPassword(false);
                      setIsLogin(true);
                    } else {
                      setIsLogin(!isLogin);
                    }
                  }}
                  className="text-sm text-muted-foreground hover:text-primary transition-colors"
                >
                  {isForgotPassword ? (
                    <>
                      Remember your password?{' '}
                      <span className="font-semibold text-primary">Sign in</span>
                    </>
                  ) : isLogin ? (
                    <>
                      Don't have an account?{' '}
                      <span className="font-semibold text-primary">Sign up</span>
                    </>
                  ) : (
                    <>
                      Already have an account?{' '}
                      <span className="font-semibold text-primary">Sign in</span>
                    </>
                  )}
                </button>
              </div>
            )}

            <div className="mt-6 pt-6 border-t text-center">
              <p className="text-xs text-muted-foreground">
                Authentication powered by Firebase. Your credentials are securely managed.
              </p>
            </div>
          </CardContent>
        </Card>

        <div className="text-center mt-6">
          <button
            onClick={() => navigate('/')}
            className="text-sm text-muted-foreground hover:text-primary transition-colors"
          >
            ← Back to Home
          </button>
        </div>
      </motion.div>
    </div>
  );
};
