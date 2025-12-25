import express from 'express';
import { User } from '../models/User.js';
import { authenticateUser } from '../middleware/auth.js';

const router = express.Router();

// Get or create user profile
router.post('/profile', authenticateUser, async (req, res) => {
  try {
    const { uid, email, name } = req.user;

    let user = await User.findOne({ firebaseUid: uid });

    if (!user) {
      user = await User.create({
        firebaseUid: uid,
        email,
        name,
      });
    }

    res.json({
      success: true,
      data: {
        id: user._id,
        firebaseUid: user.firebaseUid,
        email: user.email,
        name: user.name,
        createdAt: user.createdAt,
      },
    });
  } catch (error) {
    console.error('Get/Create user error:', error);
    res.status(500).json({
      success: false,
      message: 'Error fetching user profile',
    });
  }
});

// Update user profile
router.put('/profile', authenticateUser, async (req, res) => {
  try {
    const { name } = req.body;
    
    const user = await User.findOneAndUpdate(
      { firebaseUid: req.user.uid },
      { name },
      { new: true }
    );

    if (!user) {
      return res.status(404).json({
        success: false,
        message: 'User not found',
      });
    }

    res.json({
      success: true,
      data: {
        id: user._id,
        firebaseUid: user.firebaseUid,
        email: user.email,
        name: user.name,
      },
    });
  } catch (error) {
    console.error('Update user error:', error);
    res.status(500).json({
      success: false,
      message: 'Error updating user profile',
    });
  }
});

export default router;
