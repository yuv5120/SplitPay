import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Layout } from '../components/Layout';
import { GroupCard } from '../components/GroupCard';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '../components/ui/dialog';
import { Card, CardContent } from '../components/ui/card';
import { Plus, Users, Receipt, TrendingUp, Search } from 'lucide-react';
import { api } from '../lib/api';
import { toast } from 'sonner';
import { motion } from 'framer-motion';

export const Dashboard = () => {
  const navigate = useNavigate();
  const [groups, setGroups] = useState([]);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [newGroup, setNewGroup] = useState({
    name: '',
    members: [{ name: '', id: Date.now().toString() }],
  });

  useEffect(() => {
    // Auth handled globally
    loadGroups();
  }, [navigate]);

  const loadGroups = async () => {
    try {
      const data = await api.getGroups();
      setGroups(data);
    } catch (err) {
      toast.error('Failed to load groups');
    }
  };

  const handleAddMember = () => {
    setNewGroup({
      ...newGroup,
      members: [...newGroup.members, { name: '', id: Date.now().toString() }],
    });
  };

  const handleRemoveMember = (index) => {
    if (newGroup.members.length > 1) {
      const updatedMembers = newGroup.members.filter((_, idx) => idx !== index);
      setNewGroup({ ...newGroup, members: updatedMembers });
    }
  };

  const handleMemberNameChange = (index, name) => {
    const updatedMembers = [...newGroup.members];
    updatedMembers[index] = { ...updatedMembers[index], name };
    setNewGroup({ ...newGroup, members: updatedMembers });
  };

  const handleCreateGroup = async () => {
    if (!newGroup.name.trim()) {
      toast.error('Please enter a group name');
      return;
    }
    let validMembers = newGroup.members
      .filter(m => m.name.trim() !== '')
      .map(m => ({
        id: m.id && typeof m.id === 'string' && m.id.length > 0 ? m.id : `${Date.now()}_${Math.random().toString(36).slice(2)}`,
        name: m.name.trim()
      }));
    // Debug: log validMembers before API call
    console.log('Valid members sent to backend:', validMembers);
    if (validMembers.length < 2) {
      toast.error('Add at least 2 members to the group');
      return;
    }
    const group = {
      name: newGroup.name,
      members: validMembers.map(m => ({
        id: m.id && typeof m.id === 'string' && m.id.length > 0 ? m.id : `${Date.now()}_${Math.random().toString(36).slice(2)}`,
        name: m.name.trim()
      })),
    };
    console.log('Payload sent to api.createGroup:', group);
    try {
      await api.createGroup(group);
      toast.success('Group created successfully!');
      setShowCreateDialog(false);
      setNewGroup({ name: '', members: [{ name: '', id: Date.now().toString() }] });
      loadGroups();
    } catch (err) {
      console.error('Create group error:', err);
      toast.error(err.message || 'Failed to create group');
    }
  };

  const handleDeleteGroup = async (groupId) => {
    try {
      await api.deleteGroup(groupId);
      toast.success('Group deleted successfully!');
      loadGroups();
    } catch (err) {
      toast.error('Failed to delete group');
    }
  };

  const filteredGroups = groups.filter(group => 
    group.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const totalExpenses = groups.reduce((sum, group) => 
    sum + (group.expenses?.reduce((s, e) => s + e.amount, 0) || 0), 0
  );

  return (
    <Layout>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.5 }}
      >
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-4xl font-bold mb-2">My Groups</h1>
          <p className="text-muted-foreground text-lg">
            Manage your expense groups and track shared costs
          </p>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <Card className="glass-card hover:shadow-elegant transition-all">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground mb-1">Total Groups</p>
                  <p className="text-3xl font-bold gradient-text">{groups.length}</p>
                </div>
                <div className="w-12 h-12 rounded-xl gradient-primary flex items-center justify-center shadow-md">
                  <Users className="w-6 h-6 text-primary-foreground" />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="glass-card hover:shadow-elegant transition-all">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground mb-1">Total Expenses</p>
                  <p className="text-3xl font-bold gradient-text">
                    ₹{totalExpenses.toLocaleString('en-IN')}
                  </p>
                </div>
                <div className="w-12 h-12 rounded-xl gradient-secondary flex items-center justify-center shadow-md">
                  <Receipt className="w-6 h-6 text-secondary-foreground" />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="glass-card hover:shadow-elegant transition-all">
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground mb-1">Active Members</p>
                  <p className="text-3xl font-bold gradient-text">
                    {groups.reduce((sum, g) => sum + g.members.length, 0)}
                  </p>
                </div>
                <div className="w-12 h-12 rounded-xl bg-accent/20 flex items-center justify-center shadow-md">
                  <TrendingUp className="w-6 h-6 text-accent" />
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Actions Bar */}
        <div className="flex flex-col sm:flex-row gap-4 mb-8">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-muted-foreground" />
            <Input
              placeholder="Search groups..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
          <Button
            variant="premium"
            size="lg"
            onClick={() => setShowCreateDialog(true)}
          >
            <Plus className="w-5 h-5 mr-2" />
            New Group
          </Button>
        </div>

        {/* Groups Grid */}
        {filteredGroups.length > 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredGroups.map((group, idx) => (
              <motion.div
                key={group.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.3, delay: idx * 0.05 }}
              >
                <GroupCard group={group} onDelete={handleDeleteGroup} />
              </motion.div>
            ))}
          </div>
        ) : (
          <Card className="glass-card">
            <CardContent className="p-12 text-center">
              <div className="w-20 h-20 rounded-2xl gradient-primary flex items-center justify-center mx-auto mb-6 shadow-glow">
                <Users className="w-10 h-10 text-primary-foreground" />
              </div>
              <h3 className="text-2xl font-bold mb-2">
                {searchQuery ? 'No groups found' : 'No groups yet'}
              </h3>
              <p className="text-muted-foreground mb-6">
                {searchQuery 
                  ? 'Try adjusting your search query'
                  : 'Create your first group to start tracking expenses'}
              </p>
              {!searchQuery && (
                <Button variant="premium" onClick={() => setShowCreateDialog(true)}>
                  <Plus className="w-5 h-5 mr-2" />
                  Create Your First Group
                </Button>
              )}
            </CardContent>
          </Card>
        )}
      </motion.div>

      {/* Create Group Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent onClose={() => setShowCreateDialog(false)}>
          <DialogHeader>
            <DialogTitle>Create New Group</DialogTitle>
          </DialogHeader>
          
          <div className="space-y-6 mt-4">
            <div className="space-y-2">
              <Label htmlFor="groupName">Group Name</Label>
              <Input
                id="groupName"
                placeholder="e.g., Weekend Trip, Roommates, Office Lunch"
                value={newGroup.name}
                onChange={(e) => setNewGroup({ ...newGroup, name: e.target.value })}
              />
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <Label>Members</Label>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleAddMember}
                >
                  <Plus className="w-4 h-4 mr-1" />
                  Add Member
                </Button>
              </div>
              
              <div className="space-y-2 max-h-60 overflow-y-auto">
                {newGroup.members.map((member, index) => (
                  <div key={member.id} className="flex gap-2">
                    <Input
                      placeholder={`Member ${index + 1} name`}
                      value={member.name}
                      onChange={(e) => handleMemberNameChange(index, e.target.value)}
                    />
                    {newGroup.members.length > 1 && (
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        onClick={() => handleRemoveMember(index)}
                        className="text-destructive hover:text-destructive hover:bg-destructive/10"
                      >
                        <Plus className="w-5 h-5 rotate-45" />
                      </Button>
                    )}
                  </div>
                ))}
              </div>
            </div>

            <div className="flex gap-3 pt-4">
              <Button
                variant="outline"
                className="flex-1"
                onClick={() => setShowCreateDialog(false)}
              >
                Cancel
              </Button>
              <Button
                variant="premium"
                className="flex-1"
                onClick={handleCreateGroup}
              >
                Create Group
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </Layout>
  );
};
