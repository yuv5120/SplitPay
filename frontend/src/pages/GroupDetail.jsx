import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Layout } from '../components/Layout';
import { ExpenseCard } from '../components/ExpenseCard';
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
  ChevronDown
} from 'lucide-react';
import { api } from '../lib/api';
import { calculateBalances, simplifyDebts } from '../lib/debtSimplification';
import { formatCurrency } from '../lib/utils';
import { toast } from 'sonner';
import { motion } from 'framer-motion';

export const GroupDetail = () => {
  const { groupId } = useParams();
  const navigate = useNavigate();
  const [group, setGroup] = useState(null);
  const [showAddExpense, setShowAddExpense] = useState(false);
  const [showBalances, setShowBalances] = useState(false);
  const [showSettlements, setShowSettlements] = useState(false);
  const [newExpense, setNewExpense] = useState({
    description: '',
    amount: '',
    paidBy: '',
    participants: [],
  });

  useEffect(() => {
    loadGroup();
  }, [groupId]);

  const loadGroup = async () => {
    try {
      const data = await api.getGroup(groupId);
      setGroup(data);
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
    const amount = parseFloat(newExpense.amount);
    if (isNaN(amount) || amount <= 0) {
      toast.error('Please enter a valid amount');
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
      amount: amount,
      paidBy: newExpense.paidBy,
      participants: newExpense.participants,
    };
    try {
      await api.addExpense(groupId, expense);
      toast.success('Expense added successfully!');
      setShowAddExpense(false);
      setNewExpense({ description: '', amount: '', paidBy: '', participants: [] });
      loadGroup();
    } catch (err) {
      toast.error('Failed to add expense');
    }
  };

  const handleDeleteExpense = async (expenseId) => {
    try {
      await api.deleteExpense(groupId, expenseId);
      toast.success('Expense deleted successfully!');
      loadGroup();
    } catch (err) {
      toast.error('Failed to delete expense');
    }
  };

  const toggleParticipant = (memberId) => {
    setNewExpense(prev => ({
      ...prev,
      participants: prev.participants.includes(memberId)
        ? prev.participants.filter(id => id !== memberId)
        : [...prev.participants, memberId],
    }));
  };

  const selectAllParticipants = () => {
    setNewExpense(prev => ({
      ...prev,
      participants: group.members.map(m => m.id),
    }));
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

  const balances = calculateBalances(group.expenses || [], group.members);
  const settlements = simplifyDebts(balances);
  const totalExpenses = (group.expenses || []).reduce((sum, exp) => sum + exp.amount, 0);

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
                  {group.expenses?.length || 0} expenses
                </Badge>
              </div>
            </div>
            
            <Button
              variant="premium"
              size="lg"
              onClick={() => setShowAddExpense(true)}
            >
              <Plus className="w-5 h-5 mr-2" />
              Add Expense
            </Button>
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

        {/* Members Section */}
        <Card className="glass-card mb-8">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Users className="w-5 h-5" />
              Members
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-3">
              {group.members.map(member => (
                <Badge key={member.id} variant="outline" className="text-sm px-4 py-2">
                  <div className="w-6 h-6 rounded-full gradient-primary flex items-center justify-center text-primary-foreground text-xs mr-2">
                    {member.name.charAt(0).toUpperCase()}
                  </div>
                  {member.name}
                </Badge>
              ))}
            </div>
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
                  <ExpenseCard expense={expense} members={group.members} onDelete={handleDeleteExpense} />
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
                <Button variant="premium" onClick={() => setShowAddExpense(true)}>
                  <Plus className="w-5 h-5 mr-2" />
                  Add First Expense
                </Button>
              </CardContent>
            </Card>
          )}
        </div>
      </motion.div>

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
                  {group.members.map(member => (
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
                {group.members.map(member => (
                  <div
                    key={member.id}
                    className="flex items-center space-x-3 p-2 rounded-lg hover:bg-accent/10 transition-colors cursor-pointer"
                    onClick={() => toggleParticipant(member.id)}
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
                        {formatCurrency(parseFloat(newExpense.amount) / newExpense.participants.length)}
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
    </Layout>
  );
};
