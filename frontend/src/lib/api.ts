// API utility for backend requests
import axios from 'axios';
import { auth } from './firebase';

const BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:5001';
let clerkTokenGetter: (() => Promise<string | null>) | null = null;
export const setClerkTokenGetter = (getter: typeof clerkTokenGetter) => { clerkTokenGetter = getter; };

export interface Group {
  id: string;
  name: string;
  members: Member[];
  expenses: Expense[];
  canManage: boolean;
  currentMemberId?: string;
  activity: { id: string; kind: string; actorName: string; detail: string; createdAt: string }[];
  balancesPaise: Record<string, number>;
  settlements: { from: string; to: string; amountPaise: number }[];
  payments: PaymentAttempt[];
  remindersEnabled: boolean;
  recurring: RecurringCost[];
  monthAheadForecastPaise: number;
  budget: BudgetOverview;
  purchasePlans: PurchasePlan[];
  createdAt: string;
}

export interface BudgetOverview {
  period: string;
  limitPaise: number;
  approvedPaise: number;
  pendingPaise: number;
  scheduledPaise: number;
  projectedPaise: number;
  remainingPaise: number;
  overBudgetPaise: number;
}

export interface PurchasePlan {
  id: string;
  title: string;
  details: string;
  options: { id: string; label: string; amountPaise: number }[];
  voteCounts: Record<string, number>;
  myVote: string;
  comments: { id: string; authorName: string; text: string; createdAt: string }[];
  status: 'open' | 'closed' | 'cancelled';
  selectedOptionId?: string;
  convertedExpenseId?: string;
  canVote: boolean;
  canClose: boolean;
  canCancel: boolean;
  canConvert: boolean;
  deadlineAt: string;
  createdAt: string;
}

export interface Member {
  id: string;
  name: string;
  isJoined: boolean;
}

export interface Expense {
  id: string;
  description: string;
  amount: number;
  amountPaise: number;
  paidBy: string;
  participants: string[];
  date: string;
  canDelete: boolean;
  canEdit: boolean;
  canWithdraw: boolean;
  status: 'proposed' | 'active' | 'rejected' | 'withdrawn';
  cycle: number;
  approvals: { memberId: string; status: 'pending' | 'approved' | 'rejected'; decidedAt?: string }[];
  decisionHistory: { memberId: string; decision: string; cycle: number; at: string }[];
  recurringId?: string;
  periodKey?: string;
  anomalyExplanation?: string;
  purchasePlanId?: string;
}

export interface RecurringCost {
  id: string;
  description: string;
  estimatePaise: number;
  dueDay: number;
  paidBy: string;
  participants: string[];
  nextDueDate: string;
  estimatedSharesPaise: Record<string, number>;
  recentAveragePaise: number;
  recentCount: number;
  lastProposedPeriod: string;
  canPropose: boolean;
  canManage: boolean;
}

export interface RecurringInput {
  description: string;
  estimatePaise: number;
  dueDay: number;
  paidBy: string;
  participants: string[];
}

export interface ExpenseComment {
  id: string;
  authorName: string;
  text: string;
  createdAt: string;
}

export interface ExpenseNotification {
  id: string;
  expenseId: string;
  kind: string;
  message: string;
  createdAt: string;
  readAt?: string;
}

export interface PaymentAttempt {
  id: string;
  from: string;
  to: string;
  amountPaise: number;
  method: 'simulated' | 'razorpay_test';
  status: 'pending' | 'succeeded' | 'failed' | 'cancelled' | 'review_required';
  gatewayOrderId?: string;
  gatewayPaymentId?: string;
  createdAt: string;
  completedAt?: string;
}

interface UserProfile {
  id: string;
  firebaseUid: string;
  email: string;
  name: string;
  phone: string;
  emailVerified: boolean;
  createdAt: string;
}

const getAuthHeader = async (): Promise<{ Authorization: string } | {}> => {
  if (process.env.REACT_APP_AUTH_PROVIDER === 'clerk') {
    const token = await clerkTokenGetter?.();
    if (!token) throw new Error('Sign in to continue');
    return { Authorization: `Bearer ${token}` };
  }
  const user = auth.currentUser;
  if (user) {
    const token = await user.getIdToken();
    return { Authorization: `Bearer ${token}` };
  }
  return {};
};

export const api = {
  // User Profile
  createOrGetProfile: async (phoneNumber?: string): Promise<UserProfile> => {
    const headers = await getAuthHeader();
    try {
      const body = phoneNumber ? { phone: phoneNumber } : {};
      const res = await axios.post(`${BASE_URL}/api/users/profile`, body, { headers });
      return res.data.data || res.data;
    } catch (error) {
      console.error('Profile creation error:', axios.isAxiosError(error) ? error.response?.data : error);
      throw error;
    }
  },

  updateProfile: async (data: { name?: string; phone?: string }): Promise<UserProfile> => {
    const headers = await getAuthHeader();
    const res = await axios.put(`${BASE_URL}/api/users/profile`, data, { headers });
    return res.data.data || res.data;
  },

  // Groups
  getGroups: async (): Promise<Group[]> => {
    const headers = await getAuthHeader();
    const res = await axios.get(`${BASE_URL}/api/groups`, { headers });
    return res.data.data || res.data;
  },

  createGroup: async (group: { name: string; members: { name: string }[] }): Promise<Group> => {
    const headers = await getAuthHeader();
    try {
      const res = await axios.post(`${BASE_URL}/api/groups`, group, { headers });
      return res.data.data || res.data;
    } catch (error) {
      const errorMessage = axios.isAxiosError(error)
        ? error.response?.data?.message || error.message
        : 'Failed to create group';
      console.error('Create group error:', errorMessage);
      throw new Error(errorMessage || 'Failed to create group. Please make sure you are logged in.');
    }
  },

  deleteGroup: async (groupId: string): Promise<any> => {
    const headers = await getAuthHeader();
    const res = await axios.delete(`${BASE_URL}/api/groups/${groupId}`, { headers });
    return res.data.data || res.data;
  },

  // Expenses
  addExpense: async (groupId: string, expense: Pick<Expense, 'description' | 'amountPaise' | 'paidBy' | 'participants'>): Promise<Expense> => {
    const headers = await getAuthHeader();
    const res = await axios.post(`${BASE_URL}/api/groups/${groupId}/expenses`, expense, { headers });
    return res.data.data || res.data;
  },

  deleteExpense: async (groupId: string, expenseId: string): Promise<any> => {
    const headers = await getAuthHeader();
    const res = await axios.delete(`${BASE_URL}/api/groups/${groupId}/expenses/${expenseId}`, { headers });
    return res.data.data || res.data;
  },

  decideExpense: async (groupId: string, expenseId: string, decision: 'approved' | 'rejected'): Promise<Expense> => {
    const headers = await getAuthHeader();
    const res = await axios.post(`${BASE_URL}/api/groups/${groupId}/expenses/${expenseId}/decision`, { decision }, { headers });
    return res.data.data;
  },

  resubmitExpense: async (groupId: string, expenseId: string, expense: Pick<Expense, 'description' | 'amountPaise' | 'paidBy' | 'participants'>): Promise<Expense> => {
    const headers = await getAuthHeader();
    const res = await axios.put(`${BASE_URL}/api/groups/${groupId}/expenses/${expenseId}`, expense, { headers });
    return res.data.data;
  },

  withdrawExpense: async (groupId: string, expenseId: string): Promise<Expense> => {
    const headers = await getAuthHeader();
    const res = await axios.post(`${BASE_URL}/api/groups/${groupId}/expenses/${expenseId}/withdraw`, {}, { headers });
    return res.data.data;
  },

  getExpenseComments: async (groupId: string, expenseId: string): Promise<ExpenseComment[]> => {
    const headers = await getAuthHeader();
    const res = await axios.get(`${BASE_URL}/api/groups/${groupId}/expenses/${expenseId}/comments`, { headers });
    return res.data.data;
  },

  addExpenseComment: async (groupId: string, expenseId: string, text: string): Promise<ExpenseComment> => {
    const headers = await getAuthHeader();
    const res = await axios.post(`${BASE_URL}/api/groups/${groupId}/expenses/${expenseId}/comments`, { text }, { headers });
    return res.data.data;
  },

  getExpenseNotifications: async (groupId: string): Promise<ExpenseNotification[]> => {
    const headers = await getAuthHeader();
    const res = await axios.get(`${BASE_URL}/api/groups/${groupId}/notifications`, { headers });
    return res.data.data;
  },

  readExpenseNotification: async (groupId: string, notificationId: string): Promise<void> => {
    const headers = await getAuthHeader();
    await axios.post(`${BASE_URL}/api/groups/${groupId}/notifications/${notificationId}/read`, {}, { headers });
  },

  getPaymentConfig: async (): Promise<{ razorpayTestEnabled: boolean; razorpayTestKeyId?: string }> => {
    const headers = await getAuthHeader();
    const res = await axios.get(`${BASE_URL}/api/payment-config`, { headers });
    return res.data.data;
  },

  startPayment: async (groupId: string, to: string, amountPaise: number, method: PaymentAttempt['method']): Promise<PaymentAttempt> => {
    const headers = await getAuthHeader();
    const res = await axios.post(`${BASE_URL}/api/groups/${groupId}/payments`, { to, amountPaise, method }, { headers });
    return res.data.data;
  },

  verifyRazorpayPayment: async (groupId: string, paymentId: string, response: { razorpay_order_id: string; razorpay_payment_id: string; razorpay_signature: string }): Promise<PaymentAttempt> => {
    const headers = await getAuthHeader();
    const res = await axios.post(`${BASE_URL}/api/groups/${groupId}/payments/${paymentId}/verify`, response, { headers });
    return res.data.data;
  },

  reconcileRazorpayPayment: async (groupId: string, paymentId: string): Promise<PaymentAttempt> => {
    const headers = await getAuthHeader();
    const res = await axios.post(`${BASE_URL}/api/groups/${groupId}/payments/${paymentId}/reconcile`, {}, { headers });
    return res.data.data;
  },

  confirmPayment: async (groupId: string, paymentId: string, outcome: 'success' | 'failure' | 'cancel'): Promise<PaymentAttempt> => {
    const headers = await getAuthHeader();
    const res = await axios.post(`${BASE_URL}/api/groups/${groupId}/payments/${paymentId}/confirm`, { outcome }, { headers });
    return res.data.data;
  },

  setReminderPreference: async (groupId: string, enabled: boolean): Promise<void> => {
    const headers = await getAuthHeader();
    await axios.put(`${BASE_URL}/api/groups/${groupId}/reminders/preference`, { enabled }, { headers });
  },

  sendReminder: async (groupId: string, from: string): Promise<void> => {
    const headers = await getAuthHeader();
    await axios.post(`${BASE_URL}/api/groups/${groupId}/reminders`, { from }, { headers });
  },

  createRecurring: async (groupId: string, input: RecurringInput): Promise<void> => {
    const headers = await getAuthHeader();
    await axios.post(`${BASE_URL}/api/groups/${groupId}/recurring`, input, { headers });
  },

  updateRecurring: async (groupId: string, recurringId: string, input: RecurringInput): Promise<void> => {
    const headers = await getAuthHeader();
    await axios.put(`${BASE_URL}/api/groups/${groupId}/recurring/${recurringId}`, input, { headers });
  },

  updateBudget: async (groupId: string, limitPaise: number): Promise<BudgetOverview> => {
    const headers = await getAuthHeader();
    const res = await axios.put(`${BASE_URL}/api/groups/${groupId}/budget`, { limitPaise }, { headers });
    return res.data.data;
  },

  createPurchasePlan: async (groupId: string, input: { title: string; details: string; votingDays: number; options: { label: string; amountPaise: number }[] }): Promise<void> => {
    const headers = await getAuthHeader();
    await axios.post(`${BASE_URL}/api/groups/${groupId}/purchase-plans`, input, { headers });
  },

  votePurchasePlan: async (groupId: string, planId: string, optionId: string): Promise<void> => {
    const headers = await getAuthHeader();
    await axios.post(`${BASE_URL}/api/groups/${groupId}/purchase-plans/${planId}/vote`, { optionId }, { headers });
  },

  commentPurchasePlan: async (groupId: string, planId: string, text: string): Promise<void> => {
    const headers = await getAuthHeader();
    await axios.post(`${BASE_URL}/api/groups/${groupId}/purchase-plans/${planId}/comments`, { text }, { headers });
  },

  closePurchasePlan: async (groupId: string, planId: string): Promise<void> => {
    const headers = await getAuthHeader();
    await axios.post(`${BASE_URL}/api/groups/${groupId}/purchase-plans/${planId}/close`, {}, { headers });
  },

  cancelPurchasePlan: async (groupId: string, planId: string): Promise<void> => {
    const headers = await getAuthHeader();
    await axios.post(`${BASE_URL}/api/groups/${groupId}/purchase-plans/${planId}/cancel`, {}, { headers });
  },

  convertPurchasePlan: async (groupId: string, planId: string, paidBy: string, participants: string[]): Promise<Expense> => {
    const headers = await getAuthHeader();
    const res = await axios.post(`${BASE_URL}/api/groups/${groupId}/purchase-plans/${planId}/convert`, { paidBy, participants }, { headers });
    return res.data.data;
  },

  archiveRecurring: async (groupId: string, recurringId: string): Promise<void> => {
    const headers = await getAuthHeader();
    await axios.post(`${BASE_URL}/api/groups/${groupId}/recurring/${recurringId}/archive`, {}, { headers });
  },

  proposeRecurring: async (groupId: string, recurringId: string, amountPaise: number): Promise<Expense> => {
    const headers = await getAuthHeader();
    const res = await axios.post(`${BASE_URL}/api/groups/${groupId}/recurring/${recurringId}/proposals`, { amountPaise }, { headers });
    return res.data.data;
  },

  getGroup: async (groupId: string): Promise<Group> => {
    const headers = await getAuthHeader();
    const res = await axios.get(`${BASE_URL}/api/groups/${groupId}`, { headers });
    return res.data.data || res.data;
  },

  claimMember: async (groupId: string, memberId: string): Promise<void> => {
    const headers = await getAuthHeader();
    await axios.post(`${BASE_URL}/api/groups/${groupId}/members/${memberId}/claim`, {}, { headers });
  },

  createInvite: async (groupId: string, memberId: string): Promise<{ token: string; expiresAt: string }> => {
    const headers = await getAuthHeader();
    const res = await axios.post(`${BASE_URL}/api/groups/${groupId}/members/${memberId}/invite`, {}, { headers });
    return res.data.data;
  },

  acceptInvite: async (token: string): Promise<{ groupId: string }> => {
    const headers = await getAuthHeader();
    const res = await axios.post(`${BASE_URL}/api/invitations/${encodeURIComponent(token)}/accept`, {}, { headers });
    return res.data.data;
  },
};

