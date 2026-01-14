// API utility for backend requests
import axios from 'axios';
import { auth } from './firebase';

const BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:5001';

interface Group {
  id: string;
  name: string;
  members: Member[];
  expenses: Expense[];
  createdAt: string;
}

interface Member {
  id: string;
  name: string;
}

interface Expense {
  id: string;
  description: string;
  amount: number;
  paidBy: string;
  participants: string[];
  date: string;
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

  createGroup: async (group: Partial<Group>): Promise<Group> => {
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
  addExpense: async (groupId: string, expense: Partial<Expense>): Promise<Expense> => {
    const headers = await getAuthHeader();
    const res = await axios.post(`${BASE_URL}/api/groups/${groupId}/expenses`, expense, { headers });
    return res.data.data || res.data;
  },

  deleteExpense: async (groupId: string, expenseId: string): Promise<any> => {
    const headers = await getAuthHeader();
    const res = await axios.delete(`${BASE_URL}/api/groups/${groupId}/expenses/${expenseId}`, { headers });
    return res.data.data || res.data;
  },

  getGroup: async (groupId: string): Promise<Group> => {
    const headers = await getAuthHeader();
    const res = await axios.get(`${BASE_URL}/api/groups/${groupId}`, { headers });
    return res.data.data || res.data;
  },
};

