import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Layout } from '../components/Layout';
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { api } from '../lib/api';

export function JoinGroup() {
  const { token } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const join = async () => {
    if (!token) return;
    setLoading(true);
    setError('');
    try {
      const result = await api.acceptInvite(token);
      navigate(`/group/${result.groupId}`, { replace: true });
    } catch (err) {
      setError('This invite is invalid, expired, or its member slot has already been claimed. Ask the group owner for a new link.');
      setLoading(false);
    }
  };

  return (
    <Layout>
      <Card className="glass-card max-w-lg mx-auto mt-12">
        <CardHeader><CardTitle>Join a SplitPay group</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <p className="text-muted-foreground">This invite links one group member slot to your account. You will be able to view the group and add expenses.</p>
          {error && <p role="alert" className="text-destructive">{error}</p>}
          <Button onClick={join} disabled={loading || !token}>{loading ? 'Joining…' : 'Join group'}</Button>
        </CardContent>
      </Card>
    </Layout>
  );
}
