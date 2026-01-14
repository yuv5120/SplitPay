// LocalStorage utilities for data persistence

const STORAGE_KEYS = {
  USER: 'expense_splitter_user',
  GROUPS: 'expense_splitter_groups',
};

interface Group {
  id: string;
  name: string;
  [key: string]: any;
}

export const storage = {
  // User operations (removed: handled by Firebase Auth)
  getUser: () => null,
  setUser: () => {},
  removeUser: () => {},
  
  // Groups operations
  getGroups: (): Group[] => {
    const groups = localStorage.getItem(STORAGE_KEYS.GROUPS);
    return groups ? JSON.parse(groups) : [];
  },
  
  setGroups: (groups: Group[]): void => {
    localStorage.setItem(STORAGE_KEYS.GROUPS, JSON.stringify(groups));
  },
  
  // Get single group
  getGroup: (groupId: string): Group | undefined => {
    const groups = storage.getGroups();
    return groups.find(g => g.id === groupId);
  },
  
  // Add new group
  addGroup: (group: Group): void => {
    const groups = storage.getGroups();
    groups.push(group);
    storage.setGroups(groups);
  },
  
  // Update group
  updateGroup: (groupId: string, updatedGroup: Partial<Group>): void => {
    const groups = storage.getGroups();
    const index = groups.findIndex(g => g.id === groupId);
    if (index !== -1) {
      groups[index] = { ...groups[index], ...updatedGroup };
      storage.setGroups(groups);
    }
  },
  
  // Delete group
  deleteGroup: (groupId: string): void => {
    const groups = storage.getGroups();
    const filtered = groups.filter(g => g.id !== groupId);
    storage.setGroups(filtered);
  },
};
