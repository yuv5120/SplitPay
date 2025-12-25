// API utility for backend requests
import axios from 'axios';
import { auth } from './firebase';

const BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:5001';

const getAuthHeader = async () => {
  const user = auth.currentUser;
  if (user) {
    const token = await user.getIdToken();
    return { Authorization: `Bearer ${token}` };
  }
  return {};
};

export const api = {
  // Groups
  getGroups: async () => {
    const headers = await getAuthHeader();
    const res = await axios.get(`${BASE_URL}/api/groups`, { headers });
    return res.data.data || res.data;
  },
  createGroup: async (group) => {
    const headers = await getAuthHeader();
    try {
      const res = await axios.post(`${BASE_URL}/api/groups`, group, { headers });
      return res.data.data || res.data;
    } catch (error) {
      console.error('Create group error:', error.response?.data || error.message);
      throw new Error(error.response?.data?.message || 'Failed to create group. Please make sure you are logged in.');
    }
  },
  deleteGroup: async (groupId) => {
    const headers = await getAuthHeader();
    const res = await axios.delete(`${BASE_URL}/api/groups/${groupId}`, { headers });
    return res.data.data || res.data;
  },
  // Expenses
  addExpense: async (groupId, expense) => {
    const headers = await getAuthHeader();
    const res = await axios.post(`${BASE_URL}/api/groups/${groupId}/expenses`, expense, { headers });
    return res.data.data || res.data;
  },
  deleteExpense: async (groupId, expenseId) => {
    const headers = await getAuthHeader();
    const res = await axios.delete(`${BASE_URL}/api/groups/${groupId}/expenses/${expenseId}`, { headers });
    return res.data.data || res.data;
  },
  getGroup: async (groupId) => {
    const headers = await getAuthHeader();
    const res = await axios.get(`${BASE_URL}/api/groups/${groupId}`, { headers });
    return res.data.data || res.data;
  },
};
