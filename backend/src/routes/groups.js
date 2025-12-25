import express from 'express';
import { Group } from '../models/Group.js';
import { authenticateUser } from '../middleware/auth.js';

const router = express.Router();

// Get all groups for authenticated user
router.get('/', authenticateUser, async (req, res) => {
  try {
    const groups = await Group.find({ userId: req.user.uid })
      .sort({ createdAt: -1 });

    res.json({
      success: true,
      data: groups.map(g => ({
        id: g.id,
        name: g.name,
        members: g.members,
        expenses: g.expenses,
        createdAt: g.createdAt,
      })),
    });
  } catch (error) {
    console.error('Get groups error:', error);
    res.status(500).json({
      success: false,
      message: 'Error fetching groups',
    });
  }
});

// Get single group
router.get('/:groupId', authenticateUser, async (req, res) => {
  try {
    const group = await Group.findOne({
      id: req.params.groupId,
      userId: req.user.uid,
    });

    if (!group) {
      return res.status(404).json({
        success: false,
        message: 'Group not found',
      });
    }

    res.json({
      success: true,
      data: {
        id: group.id,
        name: group.name,
        members: group.members,
        expenses: group.expenses,
        createdAt: group.createdAt,
      },
    });
  } catch (error) {
    console.error('Get group error:', error);
    res.status(500).json({
      success: false,
      message: 'Error fetching group',
    });
  }
});

// Create new group
router.post('/', authenticateUser, async (req, res) => {
  try {
    console.log('Create group payload:', JSON.stringify(req.body));
    const { id, name, members } = req.body;

    if (!name || !members || members.length < 2) {
      return res.status(400).json({
        success: false,
        message: 'Group name and at least 2 members are required',
      });
    }

    const group = await Group.create({
      id: id || Date.now().toString(),
      name,
      members,
      expenses: [],
      userId: req.user.uid,
    });

    res.status(201).json({
      success: true,
      data: {
        id: group.id,
        name: group.name,
        members: group.members,
        expenses: group.expenses,
        createdAt: group.createdAt,
      },
    });
  } catch (error) {
    console.error('Create group error:', error);
    res.status(500).json({
      success: false,
      message: 'Error creating group',
    });
  }
});

// Update group
router.put('/:groupId', authenticateUser, async (req, res) => {
  try {
    const { name, members, expenses } = req.body;

    const group = await Group.findOneAndUpdate(
      { id: req.params.groupId, userId: req.user.uid },
      { name, members, expenses },
      { new: true }
    );

    if (!group) {
      return res.status(404).json({
        success: false,
        message: 'Group not found',
      });
    }

    res.json({
      success: true,
      data: {
        id: group.id,
        name: group.name,
        members: group.members,
        expenses: group.expenses,
        createdAt: group.createdAt,
      },
    });
  } catch (error) {
    console.error('Update group error:', error);
    res.status(500).json({
      success: false,
      message: 'Error updating group',
    });
  }
});

// Delete group
router.delete('/:groupId', authenticateUser, async (req, res) => {
  try {
    const group = await Group.findOneAndDelete({
      id: req.params.groupId,
      userId: req.user.uid,
    });

    if (!group) {
      return res.status(404).json({
        success: false,
        message: 'Group not found',
      });
    }

    res.json({
      success: true,
      message: 'Group deleted successfully',
    });
  } catch (error) {
    console.error('Delete group error:', error);
    res.status(500).json({
      success: false,
      message: 'Error deleting group',
    });
  }
});

// Add expense to group
router.post('/:groupId/expenses', authenticateUser, async (req, res) => {
  try {
    const { id, description, amount, paidBy, participants } = req.body;

    if (!description || !amount || !paidBy || !participants) {
      return res.status(400).json({
        success: false,
        message: 'All expense fields are required',
      });
    }

    const group = await Group.findOne({
      id: req.params.groupId,
      userId: req.user.uid,
    });

    if (!group) {
      return res.status(404).json({
        success: false,
        message: 'Group not found',
      });
    }

    group.expenses.push({
      id: id || Date.now().toString(),
      description,
      amount,
      paidBy,
      participants,
      date: new Date(),
    });

    await group.save();

    res.status(201).json({
      success: true,
      data: group.expenses[group.expenses.length - 1],
    });
  } catch (error) {
    console.error('Add expense error:', error);
    res.status(500).json({
      success: false,
      message: 'Error adding expense',
    });
  }
});

// Delete expense from group
router.delete('/:groupId/expenses/:expenseId', authenticateUser, async (req, res) => {
  try {
    const group = await Group.findOne({
      id: req.params.groupId,
      userId: req.user.uid,
    });

    if (!group) {
      return res.status(404).json({
        success: false,
        message: 'Group not found',
      });
    }

    group.expenses = group.expenses.filter(exp => exp.id !== req.params.expenseId);
    await group.save();

    res.json({
      success: true,
      message: 'Expense deleted successfully',
    });
  } catch (error) {
    console.error('Delete expense error:', error);
    res.status(500).json({
      success: false,
      message: 'Error deleting expense',
    });
  }
});

export default router;
