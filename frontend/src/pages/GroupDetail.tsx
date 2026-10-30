import { useState, useEffect, FC } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Layout } from '../components/Layout';
import { ExpenseCard } from '../components/ExpenseCard';
import { ExpenseDetail } from '../components/ExpenseDetail';
import { RecurringRadar } from '../components/RecurringRadar';
import { BudgetPulse } from '../components/BudgetPulse';
import { PlanTogether } from '../components/PlanTogether';
import { BalanceCard } from '../components/BalanceCard';
import { SettlementCard } from '../components/SettlementCard';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card';
import { Badge } from '../components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '../components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select';
import { Checkbox } from '../components/ui/checkbox';
import { 
  Plus, 
  ArrowLeft, 
  Users, 
  Receipt, 
  TrendingUp,
  Calculator,
  ChevronDown,
  Bell,
  Mail
} from 'lucide-react';
import { api, Expense, ExpenseNotification, Group, PaymentAttempt } from '../lib/api';
import { formatCurrency, formatDateTime } from '../lib/utils';
import { parseRupeesToPaise } from '../lib/money';
import { openRazorpayCheckout, RazorpayCheckoutResponse } from '../lib/razorpay';
import { toast } from 'sonner';
import { motion } from 'framer-motion';

interface NewExpense {
  description: string;
  amount: string;
  paidBy: string;
  participants: string[];
}

export const GroupDetail: FC = () => {
  const { groupId } = useParams<{ groupId: string }>();
  const navigate = useNavigate();
  const [group, setGroup] = useState<Group | null>(null);
  const [showAddExpense, setShowAddExpense] = useState(false);
  const [showBalances, setShowBalances] = useState(false);
  const [showSettlements, setShowSettlements] = useState(false);
  const [showNotifications, setShowNotifications] = useState(false);
  const [inviteLink, setInviteLink] = useState('');
  const [inviteMemberName, setInviteMemberName] = useState('');
  const [inviteEmail, setInviteEmail] = useState('');
  const [selectedExpenseId, setSelectedExpenseId] = useState<string | null>(null);
  const [notifications, setNotifications] = useState<ExpenseNotification[]>([]);
  const [checkout, setCheckout] = useState<{ from: string; to: string; maxPaise: number } | null>(null);
  const [paymentAmount, setPaymentAmount] = useState('');
  const [paymentAttempt, setPaymentAttempt] = useState<PaymentAttempt | null>(null);
  const [paymentBusy, setPaymentBusy] = useState(false);
  const [paymentMethod, setPaymentMethod] = useState<PaymentAttempt['method']>('simulated');
  const [paymentConfig, setPaymentConfig] = useState<{ razorpayTestEnabled: boolean; razorpayTestKeyId?: string }>({ razorpayTestEnabled: false });
  const [newExpense, setNewExpense] = useState<NewExpense>({
    description: '',
    amount: '',
    paidBy: '',
    participants: [],
  });

  useEffect(() => {
    if (groupId) {
      loadGroup();
    }
  }, [groupId]);

  useEffect(() => { api.getPaymentConfig().then(setPaymentConfig).catch(() => setPaymentConfig({ razorpayTestEnabled: false })); }, []);

  const loadGroup = async () => {
    try {
      const data = await api.getGroup(groupId!);
      setGroup(data);
      try { setNotifications(await api.getExpenseNotifications(groupId!)); } catch { setNotifications([]); }
    } catch (err) {
      toast.error('Group not found');
      navigate('/dashboard');
    }
  };

  const handleAddExpense = async () => {
    if (!newExpense.description.trim()) {
      toast.error('Please enter a description');
      return;
    }
    const amountPaise = parseRupeesToPaise(newExpense.amount);
    if (amountPaise === null) {
      toast.error('Enter a positive amount with up to two decimal places');
      return;
    }
    if (!newExpense.paidBy) {
      toast.error('Please select who paid');
      return;
    }
    if (newExpense.participants.length === 0) {
      toast.error('Please select at least one participant');
      return;
    }
    const expense = {
      description: newExpense.description.trim(),
      amountPaise,
      paidBy: newExpense.paidBy,
      participants: newExpense.participants,
    };
    try {
      const created = await api.addExpense(groupId!, expense);
      toast.success(created.status === 'proposed' ? 'Expense proposed for approval' : 'Expense added');
      setShowAddExpense(false);
      setNewExpense({ description: '', amount: '', paidBy: '', participants: [] });
      await loadGroup();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to add expense');
    }
  };

  const handleDeleteExpense = async (expenseId: string) => {
    try {
      await api.deleteExpense(groupId!, expenseId);
      toast.success('Expense deleted successfully!');
      await loadGroup();
    } catch (err) {
      toast.error('Failed to delete expense');
    }
  };

  const handleClaimMember = async (memberId: string) => {
    try {
      await api.claimMember(groupId!, memberId);
      toast.success('Your account is now linked to this member');
      await loadGroup();
    } catch (err) {
      toast.error('Could not claim this member slot');
    }
  };

  const handleInviteMember = async (memberId: string, memberName: string) => {
    try {
      const invite = await api.createInvite(groupId!, memberId);
      setInviteLink(`${window.location.origin}/join/${invite.token}`);
      setInviteMemberName(memberName);
      setInviteEmail('');
    } catch (err) {
      toast.error('Could not create an invite for this member');
    }
  };

  const composeGmailInvite = () => {
    const email = inviteEmail.trim();
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      toast.error('Enter a valid email address');
      return;
    }
    const subject = `Join ${group?.name || 'my group'} on SplitPay`;
    const body = `Hi ${inviteMemberName},\n\nJoin our SplitPay group using this invitation link:\n${inviteLink}\n\nThe link expires in seven days. Sign in with your email, then choose Join group.`;
    const url = `https://mail.google.com/mail/?view=cm&fs=1&to=${encodeURIComponent(email)}&su=${encodeURIComponent(subject)}&body=${encodeURIComponent(body)}`;
    window.open(url, '_blank', 'noopener,noreferrer');
  };

  const openNotification = async (notification: ExpenseNotification) => {
    if (notification.expenseId) setSelectedExpenseId(notification.expenseId);
    else setShowSettlements(true);
    try {
      await api.readExpenseNotification(groupId!, notification.id);
      setNotifications(previous => previous.map(item => item.id === notification.id ? { ...item, readAt: new Date().toISOString() } : item));
    } catch {
      toast.error('Could not mark notification as read');
    }
  };

  const openCheckout = (settlement: { from: string; to: string; amountPaise: number }) => {
    setCheckout({ from: settlement.from, to: settlement.to, maxPaise: settlement.amountPaise });
    setPaymentAmount((settlement.amountPaise / 100).toFixed(2));
    setPaymentAttempt(null);
    setPaymentMethod('simulated');
    setShowSettlements(false);
  };

  const resumeCheckout = (attempt: PaymentAttempt) => {
    setCheckout({ from: attempt.from, to: attempt.to, maxPaise: attempt.amountPaise });
    setPaymentAmount((attempt.amountPaise / 100).toFixed(2));
    setPaymentAttempt(attempt);
    setPaymentMethod(attempt.method);
  };

  const openGateway = async (attempt: PaymentAttempt) => {
    if (!attempt.gatewayOrderId || !paymentConfig.razorpayTestKeyId) { toast.error('Razorpay Test Mode is not configured'); return; }
    try {
      await openRazorpayCheckout({
        key: paymentConfig.razorpayTestKeyId,
        amount: attempt.amountPaise,
        currency: 'INR',
        name: 'SplitPay Test Mode',
        description: 'Demo settlement — no real transfer to group member',
        order_id: attempt.gatewayOrderId,
        theme: { color: '#7c3aed' },
        handler: async (response: RazorpayCheckoutResponse) => {
          setPaymentBusy(true);
          try {
            const result = await api.verifyRazorpayPayment(groupId!, attempt.id, response);
            setPaymentAttempt(result);
            await loadGroup();
            if (result.status === 'succeeded') { toast.success('Razorpay test payment recorded'); setCheckout(null); setPaymentAttempt(null); }
            else toast.error('Test payment captured, but balances changed. Review this settlement.');
          } catch { toast.error('Verification is pending. Use Check payment status to retry.'); }
          finally { setPaymentBusy(false); }
        }
      });
    } catch (error) { toast.error(error instanceof Error ? error.message : 'Could not open Razorpay checkout'); }
  };

  const checkGatewayStatus = async () => {
    if (!paymentAttempt) return;
    setPaymentBusy(true);
    try {
      const result = await api.reconcileRazorpayPayment(groupId!, paymentAttempt.id);
      setPaymentAttempt(result);
      await loadGroup();
      if (result.status === 'succeeded') { toast.success('Razorpay test payment recorded'); setCheckout(null); setPaymentAttempt(null); }
      else toast.error('Test payment captured, but balances changed. Review this settlement.');
    } catch { toast.error('No captured test payment found yet. You can resume checkout.'); }
    finally { setPaymentBusy(false); }
  };

  const startCheckout = async () => {
    if (!checkout) return;
    const amountPaise = parseRupeesToPaise(paymentAmount);
    if (amountPaise === null || amountPaise > checkout.maxPaise) {
      toast.error('Enter an amount up to the exact outstanding balance');
      return;
    }
    setPaymentBusy(true);
    try {
      const attempt = await api.startPayment(groupId!, checkout.to, amountPaise, paymentMethod);
      setPaymentAttempt(attempt);
      await loadGroup();
      if (attempt.method === 'razorpay_test') await openGateway(attempt);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Could not start payment');
    } finally { setPaymentBusy(false); }
  };

  const finishCheckout = async (outcome: 'success' | 'failure' | 'cancel') => {
    if (!paymentAttempt) return;
    setPaymentBusy(true);
    try {
      const result = await api.confirmPayment(groupId!, paymentAttempt.id, outcome);
      toast.success(result.status === 'succeeded' ? 'Simulated payment recorded' : `Simulated payment ${result.status}`);
      setCheckout(null);
      setPaymentAttempt(null);
      await loadGroup();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Could not finish simulated payment');
    } finally { setPaymentBusy(false); }
  };

  const remindMember = async (from: string) => {
    try {
      await api.sendReminder(groupId!, from);
      toast.success('Reminder sent');
      await loadGroup();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Could not send reminder');
    }
  };

  const setRemindersEnabled = async (enabled: boolean) => {
    try {
      await api.setReminderPreference(groupId!, enabled);
      setGroup(previous => previous ? { ...previous, remindersEnabled: enabled } : previous);
      toast.success(enabled ? 'Reminders enabled' : 'Reminders turned off');
    } catch { toast.error('Could not update reminder preference'); }
  };

  const toggleParticipant = (memberId: string) => {
    setNewExpense(prev => ({
      ...prev,
      participants: prev.participants.includes(memberId)
        ? prev.participants.filter(id => id !== memberId)
        : [...prev.participants, memberId],
    }));
  };

  const selectAllParticipants = () => {
    if (group) {
      setNewExpense(prev => ({
        ...prev,
        participants: group.members.filter(m => m.isJoined).map(m => m.id),
      }));
    }
  };

  if (!group) {
    return (
      <Layout>
        <div className="text-center py-20">
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </Layout>
    );
  }

  const balances = Object.fromEntries(Object.entries(group.balancesPaise || {}).map(([id, paise]) => [id, paise / 100]));
  const settlements = (group.settlements || []).map(item => ({ ...item, amount: item.amountPaise / 100 }));
  const activeExpenses = (group.expenses || []).filter(exp => exp.status === 'active');
  const totalExpenses = activeExpenses.reduce((sum, exp) => sum + exp.amountPaise, 0) / 100;
  const selectedExpense: Expense | undefined = group.expenses?.find(expense => expense.id === selectedExpenseId);
  const pendingExpenses = group.expenses?.filter(expense => expense.status === 'proposed') || [];
  const unreadNotifications = notifications.filter(item => !item.readAt);
  const timeline = [
    ...(group.activity || []).filter(item => !item.kind.startsWith('payment_')).map(item => ({ id: item.id, at: item.createdAt, label: `${item.actorName} ${item.detail.charAt(0).toLowerCase() + item.detail.slice(1)}`, payment: undefined as PaymentAttempt | undefined })),
    ...(group.payments || []).map(item => ({ id: item.id, at: item.completedAt || item.createdAt, label: `${group.members.find(member => member.id === item.from)?.name || 'Member'} → ${group.members.find(member => member.id === item.to)?.name || 'Member'}: ${item.method === 'razorpay_test' ? 'Razorpay test' : 'simulated'} payment ${formatCurrency(item.amountPaise / 100)} ${item.status.replace('_', ' ')}`, payment: item as PaymentAttempt | undefined }))
  ].sort((a, b) => new Date(b.at).getTime() - new Date(a.at).getTime());

  return (
    <Layout>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.5 }}
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
          
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div>
              <h1 className="text-4xl font-bold mb-2">{group.name}</h1>
              <div className="flex flex-wrap gap-2">
                <Badge variant="glass">
                  <Users className="w-3 h-3 mr-1" />
                  {group.members.length} members
                </Badge>
                <Badge variant="glass">
                  <Receipt className="w-3 h-3 mr-1" />
                  {activeExpenses.length} active expenses
                </Badge>
                {pendingExpenses.length > 0 && <Badge variant="outline">{pendingExpenses.length} awaiting approval</Badge>}
              </div>
            </div>
            
            <div className="flex items-center gap-3">
              <Button variant="outline" size="icon" className="relative" aria-label={`Notifications${unreadNotifications.length ? `, ${unreadNotifications.length} unread` : ''}`} title="Notifications" onClick={() => setShowNotifications(true)}>
                <Bell className="w-5 h-5" />
                {unreadNotifications.length > 0 && <span className="absolute -right-2 -top-2 flex h-5 min-w-5 items-center justify-center rounded-full bg-primary px-1 text-[11px] font-semibold text-primary-foreground">{unreadNotifications.length > 99 ? '99+' : unreadNotifications.length}</span>}
              </Button>
              <Button variant="premium" size="lg" onClick={() => setShowAddExpense(true)} disabled={!group.currentMemberId}>
                <Plus className="w-5 h-5 mr-2" />
                Add Expense
              </Button>
            </div>
          </div>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <Card className="glass-card hover:shadow-elegant transition-all">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground mb-1">Total Spent</p>
                  <p className="text-3xl font-bold gradient-text">
                    {formatCurrency(totalExpenses)}
                  </p>
                </div>
                <div className="w-12 h-12 rounded-xl gradient-primary flex items-center justify-center shadow-md">
                  <TrendingUp className="w-6 h-6 text-primary-foreground" />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card 
            className="glass-card hover:shadow-elegant transition-all cursor-pointer"
            onClick={() => setShowBalances(true)}
          >
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground mb-1">View Balances</p>
                  <p className="text-lg font-semibold text-primary">
                    See who owes whom →
                  </p>
                </div>
                <div className="w-12 h-12 rounded-xl bg-accent/20 flex items-center justify-center shadow-md">
                  <Calculator className="w-6 h-6 text-accent" />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card 
            className="glass-card hover:shadow-elegant transition-all cursor-pointer"
            onClick={() => setShowSettlements(true)}
          >
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground mb-1">Settlements</p>
                  <p className="text-lg font-semibold text-primary">
                    {settlements.length} transaction{settlements.length !== 1 ? 's' : ''} →
                  </p>
                </div>
                <div className="w-12 h-12 rounded-xl gradient-secondary flex items-center justify-center shadow-md">
                  <ChevronDown className="w-6 h-6 text-secondary-foreground rotate-[-90deg]" />
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        <BudgetPulse group={group} onChanged={loadGroup} />
        <PlanTogether group={group} onChanged={loadGroup} onExpenseCreated={setSelectedExpenseId} />

        {/* Recurring costs */}
        <RecurringRadar group={group} onChanged={loadGroup} />

        {/* Members Section */}
        <Card className="glass-card mb-8">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Users className="w-5 h-5" />
              Members
            </CardTitle>
          </CardHeader>
          <CardContent>
            {group.canManage && !group.currentMemberId && (
              <p className="text-sm text-muted-foreground mb-4">This group was created before account-linked members. Choose your own name below, then invite the others.</p>
            )}
            <div className="flex flex-wrap gap-3">
              {group.members.map(member => (
                <div key={member.id} className="flex items-center gap-2 rounded-lg border px-3 py-2">
                  <Badge variant="outline" className="text-sm">
                    {member.name} {member.id === group.currentMemberId ? '(you)' : ''}
                  </Badge>
                  <span className="text-xs text-muted-foreground">{member.isJoined ? 'Joined' : 'Not joined'}</span>
                  {group.canManage && !member.isJoined && !group.currentMemberId && (
                    <Button size="sm" variant="outline" onClick={() => handleClaimMember(member.id)}>This is me</Button>
                  )}
                  {group.canManage && !!group.currentMemberId && !member.isJoined && (
                    <Button size="sm" variant="outline" onClick={() => handleInviteMember(member.id, member.name)}>Invite</Button>
                  )}
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {pendingExpenses.length > 0 && <Card className="glass-card mb-8"><CardHeader><CardTitle>Awaiting approval</CardTitle></CardHeader><CardContent className="space-y-2">{pendingExpenses.map(expense => <button key={expense.id} className="w-full rounded-lg border p-3 text-left hover:bg-accent/10" onClick={() => setSelectedExpenseId(expense.id)}><span className="font-medium">{expense.description}</span><span className="block text-sm text-muted-foreground">{expense.approvals.filter(item => item.status === 'pending').map(item => group.members.find(member => member.id === item.memberId)?.name).join(', ')} still to decide</span></button>)}</CardContent></Card>}

        <Card className="glass-card mb-8">
          <CardHeader><CardTitle>Settlement and expense timeline</CardTitle></CardHeader>
          <CardContent>
            {timeline.length ? (
              <ol className="space-y-3">
                {timeline.map(event => (
                  <li key={event.id} className="flex flex-col sm:flex-row sm:justify-between gap-1 border-b pb-2 last:border-0">
                    <span className="text-sm">{event.label} {event.payment?.status === 'pending' && event.payment.from === group.currentMemberId && <Button size="sm" variant="outline" className="ml-2" onClick={() => resumeCheckout(event.payment!)}>Continue checkout</Button>}{event.payment?.status === 'review_required' && <span className="block text-destructive">Captured test payment needs review; it did not change the balance.</span>}</span>
                    <time className="text-xs text-muted-foreground" dateTime={event.at}>{formatDateTime(event.at)}</time>
                  </li>
                ))}
              </ol>
            ) : <p className="text-sm text-muted-foreground">No activity recorded yet.</p>}
          </CardContent>
        </Card>

        {/* Expenses List */}
        <div className="mb-8">
          <h2 className="text-2xl font-bold mb-4">Expenses</h2>
          {group.expenses && group.expenses.length > 0 ? (
            <div className="space-y-4">
              {[...group.expenses].reverse().map((expense, idx) => (
                <motion.div
                  key={expense.id}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.3, delay: idx * 0.05 }}
                >
                  <ExpenseCard expense={expense} members={group.members} onOpen={setSelectedExpenseId} onDelete={expense.canDelete ? handleDeleteExpense : undefined} />
                </motion.div>
              ))}
            </div>
          ) : (
            <Card className="glass-card">
              <CardContent className="p-12 text-center">
                <div className="w-20 h-20 rounded-2xl gradient-primary flex items-center justify-center mx-auto mb-6 shadow-glow">
                  <Receipt className="w-10 h-10 text-primary-foreground" />
                </div>
                <h3 className="text-2xl font-bold mb-2">No expenses yet</h3>
                <p className="text-muted-foreground mb-6">
                  Add your first expense to start tracking
                </p>
                <Button variant="premium" disabled={!group.currentMemberId} onClick={() => setShowAddExpense(true)}>
                  <Plus className="w-5 h-5 mr-2" />
                  Add First Expense
                </Button>
              </CardContent>
            </Card>
          )}
        </div>
      </motion.div>

      <Dialog open={!!selectedExpense} onOpenChange={open => !open && setSelectedExpenseId(null)}>
        <DialogContent onClose={() => setSelectedExpenseId(null)} className="max-w-2xl">
          <DialogHeader><DialogTitle>Expense details</DialogTitle></DialogHeader>
          {selectedExpense && <ExpenseDetail key={selectedExpense.id} groupId={groupId!} expense={selectedExpense} members={group.members} currentMemberId={group.currentMemberId} onChanged={loadGroup} />}
        </DialogContent>
      </Dialog>

      <Dialog open={showNotifications} onOpenChange={setShowNotifications}>
        <DialogContent onClose={() => setShowNotifications(false)} className="max-w-lg">
          <DialogHeader><DialogTitle>Notifications</DialogTitle></DialogHeader>
          <div className="max-h-[65vh] overflow-y-auto space-y-2 mt-3">
            {notifications.length ? notifications.map(item => <button key={item.id} className={`w-full text-left rounded-lg border p-3 hover:bg-accent/10 ${item.readAt ? 'opacity-70' : ''}`} onClick={async () => { await openNotification(item); setShowNotifications(false); }}><span className="block text-sm">{item.message}</span><time className="text-xs text-muted-foreground" dateTime={item.createdAt}>{formatDateTime(item.createdAt)}</time></button>) : <p className="text-sm text-muted-foreground">No notifications yet.</p>}
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={!!inviteLink} onOpenChange={(open) => !open && setInviteLink('')}>
        <DialogContent onClose={() => setInviteLink('')}>
          <DialogHeader><DialogTitle>Invite {inviteMemberName}</DialogTitle></DialogHeader>
          <p className="text-sm text-muted-foreground">This link claims one member slot and expires in seven days. Anyone with the link can claim that slot, so share it only with {inviteMemberName || 'the intended person'}.</p>
          <Input aria-label="Invite link" value={inviteLink} readOnly onFocus={(event) => event.target.select()} />
          <Button variant="outline" onClick={async () => {
            try {
              await navigator.clipboard.writeText(inviteLink);
              toast.success('Invite link copied');
            } catch {
              toast.error('Select and copy the link above');
            }
          }}>Copy link</Button>
          <div className="space-y-2 border-t pt-4">
            <Label htmlFor="invite-email">Send using Gmail</Label>
            <Input id="invite-email" type="email" autoComplete="email" placeholder="friend@gmail.com" value={inviteEmail} onChange={event => setInviteEmail(event.target.value)} />
            <Button onClick={composeGmailInvite}><Mail className="w-4 h-4 mr-2" />Open Gmail draft</Button>
            <p className="text-xs text-muted-foreground">Gmail opens with the invite ready. Review it and click Send there.</p>
          </div>
        </DialogContent>
      </Dialog>

      {/* Add Expense Dialog */}
      <Dialog open={showAddExpense} onOpenChange={setShowAddExpense}>
        <DialogContent onClose={() => setShowAddExpense(false)} className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Add New Expense</DialogTitle>
          </DialogHeader>
          
          <div className="space-y-6 mt-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Input
                  id="description"
                  placeholder="e.g., Groceries, Dinner, Movie tickets"
                  value={newExpense.description}
                  onChange={(e) => setNewExpense({ ...newExpense, description: e.target.value })}
                />
              </div>
              
              <div className="space-y-2">
                <Label htmlFor="amount">Amount (₹)</Label>
                <Input
                  id="amount"
                  type="number"
                  step="0.01"
                  placeholder="0.00"
                  value={newExpense.amount}
                  onChange={(e) => setNewExpense({ ...newExpense, amount: e.target.value })}
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="paidBy">Paid By</Label>
              <Select
                value={newExpense.paidBy}
                onValueChange={(value) => setNewExpense({ ...newExpense, paidBy: value })}
              >
                <SelectTrigger placeholder="Select who paid">
                  <SelectValue placeholder="Select who paid">
                    {newExpense.paidBy ? group.members.find(m => m.id === newExpense.paidBy)?.name : null}
                  </SelectValue>
                </SelectTrigger>
                <SelectContent>
                  {group.members.filter(member => member.isJoined).map(member => (
                    <SelectItem key={member.id} value={member.id}>
                      {member.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <Label>Split Between</Label>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={selectAllParticipants}
                >
                  Select All
                </Button>
              </div>
              
              <div className="space-y-2 max-h-60 overflow-y-auto p-4 rounded-lg border bg-card/50">
                {group.members.filter(member => member.isJoined).map(member => (
                  <div
                    key={member.id}
                    className="flex items-center space-x-3 p-2 rounded-lg hover:bg-accent/10 transition-colors"
                  >
                    <Checkbox
                      checked={newExpense.participants.includes(member.id)}
                      onCheckedChange={() => toggleParticipant(member.id)}
                    />
                    <div className="flex items-center gap-2 flex-1">
                      <div className="w-8 h-8 rounded-full gradient-primary flex items-center justify-center text-primary-foreground text-xs">
                        {member.name.charAt(0).toUpperCase()}
                      </div>
                      <span className="font-medium">{member.name}</span>
                    </div>
                    {newExpense.participants.includes(member.id) && newExpense.amount && (
                      <span className="text-sm text-muted-foreground">
                        {formatCurrency((parseRupeesToPaise(newExpense.amount) || 0) / 100 / newExpense.participants.length)}
                      </span>
                    )}
                  </div>
                ))}
              </div>
            </div>

            <div className="flex gap-3 pt-4 border-t">
              <Button
                variant="outline"
                className="flex-1"
                onClick={() => setShowAddExpense(false)}
              >
                Cancel
              </Button>
              <Button
                variant="premium"
                className="flex-1"
                onClick={handleAddExpense}
              >
                Add Expense
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      {/* Balances Dialog */}
      <Dialog open={showBalances} onOpenChange={setShowBalances}>
        <DialogContent onClose={() => setShowBalances(false)} className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Member Balances</DialogTitle>
          </DialogHeader>
          
          <div className="space-y-4 mt-4">
            {group.members.map(member => (
              <BalanceCard
                key={member.id}
                member={member}
                balance={balances[member.id] || 0}
              />
            ))}
          </div>

          <div className="pt-4 border-t text-center text-sm text-muted-foreground">
            Positive balance means the person gets money back.
            <br />
            Negative balance means the person owes money.
          </div>
        </DialogContent>
      </Dialog>

      {/* Settlements Dialog */}
      <Dialog open={showSettlements} onOpenChange={setShowSettlements}>
        <DialogContent onClose={() => setShowSettlements(false)} className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Settlement Suggestions</DialogTitle>
          </DialogHeader>
          
          <div className="mt-4">
            <label className="flex items-center gap-2 text-sm mb-4"><Checkbox checked={group.remindersEnabled} onCheckedChange={value => setRemindersEnabled(value === true)} /> Allow group members to send me settlement reminders</label>
            {settlements.length > 0 ? (
              <>
                <div className="mb-4 p-4 rounded-lg bg-accent/10 border border-accent/20">
                  <p className="text-sm text-center">
                    <span className="font-semibold text-accent">
                      {settlements.length} transaction{settlements.length !== 1 ? 's' : ''}
                    </span>
                    {' '}needed to settle all balances
                  </p>
                </div>
                
                <div className="space-y-4">
                  {settlements.map((settlement, idx) => (
                    <SettlementCard
                      key={idx}
                      settlement={settlement}
                      members={group.members}
                      onSettle={settlement.from === group.currentMemberId && group.members.find(member => member.id === settlement.to)?.isJoined ? () => openCheckout({ from: settlement.from, to: settlement.to, amountPaise: Math.round(settlement.amount * 100) }) : undefined}
                      onRemind={settlement.to === group.currentMemberId && group.members.find(member => member.id === settlement.from)?.isJoined ? () => remindMember(settlement.from) : undefined}
                    />
                  ))}
                </div>
              </>
            ) : (
              <div className="text-center py-8">
                <div className="w-16 h-16 rounded-2xl gradient-primary flex items-center justify-center mx-auto mb-4 shadow-glow">
                  <Calculator className="w-8 h-8 text-primary-foreground" />
                </div>
                <h3 className="text-xl font-bold mb-2">All Settled!</h3>
                <p className="text-muted-foreground">
                  Everyone is square. No settlements needed.
                </p>
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={!!checkout} onOpenChange={open => { if (!open) setCheckout(null); }}>
        <DialogContent onClose={() => setCheckout(null)} className="max-w-lg">
          <DialogHeader><DialogTitle>Demo settlement payment</DialogTitle></DialogHeader>
          {checkout && <div className="space-y-4 mt-3">
            <p className="text-sm text-muted-foreground">Demo checkout only. No real money will move.</p>
            <p className="text-sm">From <strong>{group.members.find(member => member.id === checkout.from)?.name}</strong> to <strong>{group.members.find(member => member.id === checkout.to)?.name}</strong></p>
            <p className="text-sm">Outstanding: <strong>{formatCurrency(checkout.maxPaise / 100)}</strong></p>
            {!paymentAttempt ? <><div><Label htmlFor="payment-amount">Amount to settle (₹)</Label><Input id="payment-amount" value={paymentAmount} onChange={event => setPaymentAmount(event.target.value)} /></div><div><Label>Method</Label><Select value={paymentMethod} onValueChange={value => setPaymentMethod(value as PaymentAttempt['method'])}><SelectTrigger><SelectValue /></SelectTrigger><SelectContent><SelectItem value="simulated">Simulated payment</SelectItem>{paymentConfig.razorpayTestEnabled && <SelectItem value="razorpay_test">Razorpay Test Mode</SelectItem>}</SelectContent></Select></div>{paymentMethod === 'razorpay_test' && <p className="text-sm text-muted-foreground">Razorpay test checkout uses test cards or UPI. No real transfer reaches your group member.</p>}<Button disabled={paymentBusy} onClick={startCheckout}>Continue to demo checkout</Button></> : paymentAttempt.method === 'simulated' ? <><p className="text-sm">Choose a demo result for this payment attempt.</p><div className="flex flex-wrap gap-2"><Button disabled={paymentBusy} onClick={() => finishCheckout('success')}>Simulate success</Button><Button disabled={paymentBusy} variant="outline" onClick={() => finishCheckout('failure')}>Simulate failure</Button><Button disabled={paymentBusy} variant="outline" onClick={() => finishCheckout('cancel')}>Cancel payment</Button></div></> : <div className="space-y-3"><p className="text-sm">Razorpay Test Mode: {paymentAttempt.status.replace('_', ' ')}. No real money moves to the group member.</p>{paymentAttempt.status === 'pending' && <div className="flex flex-wrap gap-2"><Button disabled={paymentBusy} onClick={() => openGateway(paymentAttempt)}>Open test checkout</Button><Button disabled={paymentBusy} variant="outline" onClick={checkGatewayStatus}>Check payment status</Button></div>}{paymentAttempt.status === 'review_required' && <p className="text-sm text-destructive">The test payment was captured after balances changed. This attempt did not reduce the group balance; review it before settling again.</p>}</div>}
          </div>}
        </DialogContent>
      </Dialog>
    </Layout>
  );
};
