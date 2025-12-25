import mongoose from 'mongoose';

const memberSchema = new mongoose.Schema({
  id: {
    type: String,
    required: true,
  },
  name: {
    type: String,
    required: true,
  },
});

const expenseSchema = new mongoose.Schema({
  id: {
    type: String,
    required: true,
  },
  description: {
    type: String,
    required: true,
  },
  amount: {
    type: Number,
    required: true,
    min: 0,
  },
  paidBy: {
    type: String,
    required: true,
  },
  participants: [{
    type: String,
    required: true,
  }],
  date: {
    type: Date,
    default: Date.now,
  },
});

const groupSchema = new mongoose.Schema({
  id: {
    type: String,
    required: true,
    unique: true,
    index: true,
  },
  name: {
    type: String,
    required: true,
  },
  members: [memberSchema],
  expenses: [expenseSchema],
  userId: {
    type: String,
    required: true,
    index: true,
  },
  createdAt: {
    type: Date,
    default: Date.now,
  },
  updatedAt: {
    type: Date,
    default: Date.now,
  },
});

// Update timestamp on save
groupSchema.pre('save', function(next) {
  this.updatedAt = Date.now();
  next();
});

// Index for efficient queries
groupSchema.index({ userId: 1, createdAt: -1 });

export const Group = mongoose.model('Group', groupSchema);
