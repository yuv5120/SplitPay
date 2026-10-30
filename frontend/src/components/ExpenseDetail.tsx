import { useEffect, useState } from 'react';
import { api, Expense, ExpenseComment, Member } from '../lib/api';
import { parseRupeesToPaise } from '../lib/money';
import { formatCurrency, formatDateTime } from '../lib/utils';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Checkbox } from './ui/checkbox';
import { toast } from 'sonner';

interface Props {
  groupId: string;
  expense: Expense;
  members: Member[];
  currentMemberId?: string;
  onChanged: () => Promise<void>;
}

export function ExpenseDetail({ groupId, expense, members, currentMemberId, onChanged }: Props) {
  const [comments, setComments] = useState<ExpenseComment[]>([]);
  const [comment, setComment] = useState('');
  const [busy, setBusy] = useState(false);
  const [editing, setEditing] = useState(false);
  const [description, setDescription] = useState(expense.description);
  const [amount, setAmount] = useState((expense.amountPaise / 100).toFixed(2));
  const [paidBy, setPaidBy] = useState(expense.paidBy);
  const [participants, setParticipants] = useState(expense.participants);
  const joined = members.filter(member => member.isJoined);
  const memberName = (id: string) => members.find(member => member.id === id)?.name || 'Former member';
  const myApproval = expense.approvals?.find(approval => approval.memberId === currentMemberId);

  useEffect(() => {
    let live = true;
    api.getExpenseComments(groupId, expense.id).then(data => {
      if (live) setComments(data);
    }).catch(() => toast.error('Could not load comments'));
    return () => { live = false; };
  }, [groupId, expense.id]);

  const perform = async (action: () => Promise<unknown>, message: string) => {
    setBusy(true);
    try {
      await action();
      await onChanged();
      toast.success(message);
      setEditing(false);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Action failed. Refresh and try again.');
    } finally {
      setBusy(false);
    }
  };

  const submitComment = async () => {
    const text = comment.trim();
    if (!text || text.length > 1000) return;
    setBusy(true);
    try {
      const created = await api.addExpenseComment(groupId, expense.id, text);
      setComments(previous => [...previous, created]);
      setComment('');
      await onChanged();
    } catch {
      toast.error('Could not post comment');
    } finally {
      setBusy(false);
    }
  };

  const submitRevision = () => {
    const amountPaise = parseRupeesToPaise(amount);
    if (!description.trim() || amountPaise === null || !paidBy || participants.length === 0) {
      toast.error('Enter a description, positive amount, payer, and participants');
      return;
    }
    perform(() => api.resubmitExpense(groupId, expense.id, { description: description.trim(), amountPaise, paidBy, participants }), 'Expense resubmitted for approval');
  };

  return (
    <div className="space-y-6 mt-4 max-h-[75vh] overflow-y-auto pr-1">
      <div className="rounded-lg border p-4 space-y-2">
        <div className="flex justify-between gap-3"><strong>{expense.description}</strong><strong>{formatCurrency(expense.amountPaise / 100)}</strong></div>
        <p className="text-sm text-muted-foreground">Paid by {memberName(expense.paidBy)} · {expense.status} · approval cycle {expense.cycle || 1}</p>
        {expense.status === 'proposed' && <p className="text-sm">This proposal does not affect balances until everyone approves.</p>}
        {expense.recurringId && <p className="text-sm">Scheduled bill for {expense.periodKey}</p>}
        {expense.anomalyExplanation && <p className="text-sm rounded-md border border-amber-400/50 bg-amber-400/10 p-2" role="alert">{expense.anomalyExplanation}</p>}
        <div className="space-y-1">
          {(expense.approvals || []).map(approval => (
            <p key={approval.memberId} className="text-sm">{memberName(approval.memberId)}: <strong>{approval.status}</strong></p>
          ))}
        </div>
        {expense.status === 'proposed' && myApproval?.status === 'pending' && (
          <div className="flex gap-2 pt-2">
            <Button disabled={busy} onClick={() => perform(() => api.decideExpense(groupId, expense.id, 'approved'), 'Share approved')}>Approve my share</Button>
            <Button disabled={busy} variant="outline" onClick={() => perform(() => api.decideExpense(groupId, expense.id, 'rejected'), 'Proposal rejected')}>Reject</Button>
          </div>
        )}
        {expense.canWithdraw && <Button disabled={busy} variant="outline" onClick={() => {
          if (window.confirm('Withdraw this proposal? Its comments and decision history will remain visible.')) {
            perform(() => api.withdrawExpense(groupId, expense.id), 'Proposal withdrawn');
          }
        }}>Withdraw proposal</Button>}
      </div>

      {expense.canEdit && (
        <div className="rounded-lg border p-4 space-y-3">
          <Button variant="outline" onClick={() => setEditing(value => !value)}>{editing ? 'Cancel edit' : 'Edit and resubmit'}</Button>
          {editing && <>
            <p className="text-sm text-muted-foreground">Resubmitting starts a new approval cycle. Earlier decisions stay in the history.</p>
            <div><Label htmlFor="revision-description">Description</Label><Input id="revision-description" maxLength={200} value={description} onChange={event => setDescription(event.target.value)} /></div>
            <div><Label htmlFor="revision-amount">Amount (₹)</Label><Input id="revision-amount" value={amount} onChange={event => setAmount(event.target.value)} /></div>
            <div><Label htmlFor="revision-payer">Paid by</Label><select id="revision-payer" className="w-full rounded-md border bg-background p-2" value={paidBy} onChange={event => setPaidBy(event.target.value)}>{joined.map(member => <option key={member.id} value={member.id}>{member.name}</option>)}</select></div>
            <div><Label>Split between</Label>{joined.map(member => <label key={member.id} className="flex gap-2 items-center py-1"><Checkbox checked={participants.includes(member.id)} onCheckedChange={() => setParticipants(current => current.includes(member.id) ? current.filter(id => id !== member.id) : [...current, member.id])} />{member.name}</label>)}</div>
            <Button disabled={busy} onClick={submitRevision}>Resubmit for approval</Button>
          </>}
        </div>
      )}

      <div>
        <h3 className="font-semibold mb-2">Decision history</h3>
        {expense.decisionHistory?.length ? <ol className="space-y-1">{expense.decisionHistory.map((entry, index) => <li className="text-sm text-muted-foreground" key={index}>{memberName(entry.memberId)} {entry.decision} · cycle {entry.cycle} · {formatDateTime(entry.at)}</li>)}</ol> : <p className="text-sm text-muted-foreground">No decisions recorded for this older expense.</p>}
      </div>

      <div className="border-t pt-4">
        <h3 className="font-semibold mb-3">Discussion</h3>
        {comments.length ? <ol className="space-y-3 mb-4">{comments.map(item => <li key={item.id} className="rounded-lg bg-muted/50 p-3"><div className="flex justify-between gap-2 text-sm"><strong>{item.authorName}</strong><time className="text-muted-foreground" dateTime={item.createdAt}>{formatDateTime(item.createdAt)}</time></div><p className="text-sm whitespace-pre-wrap break-words mt-1">{item.text}</p></li>)}</ol> : <p className="text-sm text-muted-foreground mb-4">No comments yet.</p>}
        <Label htmlFor="expense-comment">Add a comment</Label>
        <textarea id="expense-comment" className="w-full min-h-20 rounded-md border bg-background p-2 mt-1" maxLength={1000} value={comment} onChange={event => setComment(event.target.value)} placeholder="Add context or ask a question about this expense" />
        <Button disabled={busy || !comment.trim()} onClick={submitComment}>Post comment</Button>
      </div>
    </div>
  );
}
