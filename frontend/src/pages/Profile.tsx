import { useState, useEffect, FC } from 'react';
import { useNavigate } from 'react-router-dom';
import { Layout } from '../components/Layout';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card';
import { User, Mail, Phone, LogOut, ArrowLeft, Edit2, Save, X } from 'lucide-react';
import { auth } from '../lib/firebase';
import { api } from '../lib/api';
import { toast } from 'sonner';
import { motion } from 'framer-motion';
import { useClerk } from '@clerk/react';

interface UserProfile {
    id: string;
    firebaseUid: string;
    email: string;
    name: string;
    phone: string;
    emailVerified: boolean;
    createdAt: string;
}

export const Profile: FC = () => process.env.REACT_APP_AUTH_PROVIDER === 'clerk' ? <ClerkProfile /> : <ProfileContent signOut={() => auth.signOut()} />;

const ClerkProfile: FC = () => {
    const clerk = useClerk();
    return <ProfileContent signOut={() => clerk.signOut()} />;
};

const ProfileContent: FC<{ signOut: () => Promise<void> }> = ({ signOut }) => {
    const navigate = useNavigate();
    const [profile, setProfile] = useState<UserProfile | null>(null);
    const [isEditing, setIsEditing] = useState(false);
    const [editData, setEditData] = useState({ name: '', phone: '' });
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        loadProfile();
    }, []);

    const loadProfile = async () => {
        try {
            const data = await api.createOrGetProfile();
            setProfile(data);
            setEditData({ name: data.name, phone: data.phone });
            setLoading(false);
        } catch (err) {
            toast.error('Failed to load profile');
            setLoading(false);
        }
    };

    const handleSaveProfile = async () => {
        try {
            await api.updateProfile(editData);
            toast.success('Profile updated successfully!');
            setIsEditing(false);
            loadProfile();
        } catch (err) {
            toast.error('Failed to update profile');
        }
    };

    const handleLogout = async () => {
        try {
            await signOut();
            toast.success('Logged out successfully');
            navigate('/login');
        } catch (error) {
            toast.error('Logout failed');
        }
    };

    if (loading) {
        return (
            <Layout>
                <div className="flex items-center justify-center min-h-[60vh]">
                    <div className="text-center">
                        <div className="w-16 h-16 border-4 border-primary border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
                        <p className="text-muted-foreground">Loading profile...</p>
                    </div>
                </div>
            </Layout>
        );
    }

    return (
        <Layout>
            <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ duration: 0.5 }}
                className="max-w-2xl mx-auto"
            >
                {/* Header */}
                <div className="mb-8">
                    <Button
                        variant="ghost"
                        onClick={() => navigate('/dashboard')}
                        className="mb-4"
                    >
                        <ArrowLeft className="w-4 h-4 mr-2" />
                        Back to Dashboard
                    </Button>
                    <h1 className="text-4xl font-bold mb-2">My Profile</h1>
                    <p className="text-muted-foreground text-lg">
                        Manage your account information and settings
                    </p>
                </div>

                <div>
                    <Card className="glass-card">
                        <CardHeader>
                            <div className="flex items-center justify-between">
                                <CardTitle>Profile Information</CardTitle>
                                {!isEditing ? (
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={() => setIsEditing(true)}
                                    >
                                        <Edit2 className="w-4 h-4 mr-2" />
                                        Edit
                                    </Button>
                                ) : (
                                    <div className="flex gap-2">
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            onClick={() => {
                                                setIsEditing(false);
                                                setEditData({ name: profile?.name || '', phone: profile?.phone || '' });
                                            }}
                                        >
                                            <X className="w-4 h-4 mr-2" />
                                            Cancel
                                        </Button>
                                        <Button
                                            variant="premium"
                                            size="sm"
                                            onClick={handleSaveProfile}
                                        >
                                            <Save className="w-4 h-4 mr-2" />
                                            Save
                                        </Button>
                                    </div>
                                )}
                            </div>
                        </CardHeader>
                        <CardContent className="space-y-6">
                            {/* Avatar */}
                            <div className="flex items-center gap-4">
                                <div className="w-20 h-20 rounded-full gradient-primary flex items-center justify-center text-primary-foreground font-bold text-3xl shadow-glow">
                                    {profile?.name?.charAt(0).toUpperCase() || profile?.email?.charAt(0).toUpperCase()}
                                </div>
                                <div>
                                    <h3 className="text-xl font-semibold">{profile?.name || 'User'}</h3>
                                    <p className="text-sm text-muted-foreground">{profile?.email}</p>
                                </div>
                            </div>

                            {/* Name */}
                            <div className="space-y-2">
                                <Label htmlFor="name">Full Name</Label>
                                {isEditing ? (
                                    <Input
                                        id="name"
                                        value={editData.name}
                                        onChange={(e) => setEditData({ ...editData, name: e.target.value })}
                                        placeholder="Enter your name"
                                    />
                                ) : (
                                    <div className="flex items-center gap-3 p-3 rounded-lg bg-muted/50">
                                        <User className="w-5 h-5 text-muted-foreground" />
                                        <span>{profile?.name || 'Not set'}</span>
                                    </div>
                                )}
                            </div>

                            {/* Email */}
                            <div className="space-y-2">
                                <Label htmlFor="email">Email Address</Label>
                                <div className="flex items-center gap-3 p-3 rounded-lg bg-muted/50">
                                    <Mail className="w-5 h-5 text-muted-foreground" />
                                    <span>{profile?.email}</span>
                                    {profile?.emailVerified && (
                                        <span className="ml-auto text-xs bg-green-500/20 text-green-600 px-2 py-1 rounded-full">
                                            Verified
                                        </span>
                                    )}
                                </div>
                                <p className="text-xs text-muted-foreground">Email cannot be changed</p>
                            </div>

                            {/* Phone */}
                            <div className="space-y-2">
                                <Label htmlFor="phone">Phone Number</Label>
                                {isEditing ? (
                                    <Input
                                        id="phone"
                                        type="tel"
                                        value={editData.phone}
                                        onChange={(e) => setEditData({ ...editData, phone: e.target.value })}
                                        placeholder="Enter your phone number"
                                    />
                                ) : (
                                    <div className="flex items-center gap-3 p-3 rounded-lg bg-muted/50">
                                        <Phone className="w-5 h-5 text-muted-foreground" />
                                        <span>{profile?.phone || 'Not set'}</span>
                                    </div>
                                )}
                            </div>

                            {/* Account Created */}
                            <div className="space-y-2">
                                <Label>Member Since</Label>
                                <div className="p-3 rounded-lg bg-muted/50">
                                    <span className="text-sm">
                                        {profile?.createdAt ? new Date(profile.createdAt).toLocaleDateString('en-US', {
                                            year: 'numeric',
                                            month: 'long',
                                            day: 'numeric'
                                        }) : 'Unknown'}
                                    </span>
                                </div>
                            </div>
                            <div className="border-t pt-5 space-y-4">
                                <h3 className="font-semibold">Account</h3>
                                <div className="flex items-center justify-between text-sm">
                                    <span className="text-muted-foreground">Status</span>
                                    <span className="font-medium text-green-600">Active</span>
                                </div>
                                <div className="flex items-center justify-between text-sm">
                                    <span className="text-muted-foreground">Email verified</span>
                                    <span className="font-medium">{profile?.emailVerified ? 'Yes' : 'No'}</span>
                                </div>
                                <Button variant="outline" className="w-full sm:w-auto" onClick={handleLogout}>
                                    <LogOut className="w-4 h-4 mr-2" />
                                    Log out
                                </Button>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </motion.div>
        </Layout>
    );
};
