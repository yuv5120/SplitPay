# Expense Tracker Backend

Backend API for the Expense Tracker application with Firebase Authentication and MongoDB.

## 🏗️ Architecture

- **Authentication**: Firebase Admin SDK
- **Database**: MongoDB (local)
- **Framework**: Express.js
- **ODM**: Mongoose

## 📁 Project Structure

```
backend/
├── src/
│   ├── config/
│   │   ├── database.js       # MongoDB connection
│   │   └── firebase.js       # Firebase Admin setup
│   ├── middleware/
│   │   └── auth.js           # Firebase token verification
│   ├── models/
│   │   ├── User.js           # User schema
│   │   └── Group.js          # Group & Expense schemas
│   ├── routes/
│   │   ├── users.js          # User endpoints
│   │   └── groups.js         # Group & Expense endpoints
│   └── server.js             # Main server file
├── .env                      # Environment variables
├── .env.example             # Environment template
├── package.json
└── README.md
```

## 🚀 Setup Instructions

### 1. Install Dependencies

```bash
cd backend
npm install
```

### 2. Install and Start MongoDB

**macOS (using Homebrew):**
```bash
brew tap mongodb/brew
brew install mongodb-community@7.0
brew services start mongodb-community@7.0
```

**Linux (Ubuntu):**
```bash
sudo apt-get install -y mongodb-org
sudo systemctl start mongod
sudo systemctl enable mongod
```

**Windows:**
Download and install from [MongoDB Download Center](https://www.mongodb.com/try/download/community)

**Verify MongoDB is running:**
```bash
mongosh
# Should connect to: mongodb://localhost:27017
```

### 3. Setup Firebase Admin SDK

1. Go to [Firebase Console](https://console.firebase.google.com/)
2. Select your project (or create one)
3. Go to **Project Settings** > **Service Accounts**
4. Click **Generate New Private Key**
5. Download the JSON file
6. Rename it to `firebase-service-account.json`
7. Place it in the `backend/` directory

### 4. Configure Environment Variables

Copy `.env.example` to `.env`:
```bash
cp .env.example .env
```

Update the values:
```env
PORT=5000
NODE_ENV=development
MONGODB_URI=mongodb://localhost:27017/expense-tracker
FIREBASE_SERVICE_ACCOUNT_PATH=./firebase-service-account.json
CLIENT_URL=http://localhost:3000
```

### 5. Start the Server

**Development mode (with auto-restart):**
```bash
npm run dev
```

**Production mode:**
```bash
npm start
```

Server will run on `http://localhost:5000`

## 📡 API Endpoints

### Authentication

All endpoints (except `/health`) require a Firebase ID token in the Authorization header:
```
Authorization: Bearer <firebase-id-token>
```

### Health Check

**GET** `/health`
- Check if server is running
- No authentication required

### User Endpoints

**POST** `/api/users/profile`
- Get or create user profile
- Auto-creates user on first login

**PUT** `/api/users/profile`
- Update user name
- Body: `{ "name": "New Name" }`

### Group Endpoints

**GET** `/api/groups`
- Get all groups for authenticated user

**GET** `/api/groups/:groupId`
- Get single group by ID

**POST** `/api/groups`
- Create new group
- Body:
```json
{
  "id": "generated-id",
  "name": "Group Name",
  "members": [
    { "id": "m1", "name": "Alice" },
    { "id": "m2", "name": "Bob" }
  ]
}
```

**PUT** `/api/groups/:groupId`
- Update group (name, members, or expenses)
- Body: Same as create

**DELETE** `/api/groups/:groupId`
- Delete a group

### Expense Endpoints

**POST** `/api/groups/:groupId/expenses`
- Add expense to group
- Body:
```json
{
  "id": "expense-id",
  "description": "Dinner",
  "amount": 90,
  "paidBy": "m1",
  "participants": ["m1", "m2", "m3"]
}
```

**DELETE** `/api/groups/:groupId/expenses/:expenseId`
- Delete expense from group

## 🔐 Security Features

- **Helmet**: Sets security HTTP headers
- **CORS**: Configured for frontend origin only
- **Rate Limiting**: 100 requests per 15 minutes per IP
- **Firebase Auth**: Token-based authentication
- **Input Validation**: Required fields checked

## 🗄️ Database Schema

### User Collection
```javascript
{
  firebaseUid: String (unique, indexed),
  email: String (unique),
  name: String,
  createdAt: Date,
  updatedAt: Date
}
```

### Group Collection
```javascript
{
  id: String (unique, indexed),
  name: String,
  members: [
    {
      id: String,
      name: String
    }
  ],
  expenses: [
    {
      id: String,
      description: String,
      amount: Number,
      paidBy: String,
      participants: [String],
      date: Date
    }
  ],
  userId: String (indexed),
  createdAt: Date,
  updatedAt: Date
}
```

## 🧪 Testing with cURL

**Health check:**
```bash
curl http://localhost:5000/health
```

**Get groups (with auth):**
```bash
curl -H "Authorization: Bearer YOUR_FIREBASE_TOKEN" \
     http://localhost:5000/api/groups
```

## 🐛 Troubleshooting

### MongoDB Connection Error
- Check if MongoDB is running: `brew services list` (macOS) or `sudo systemctl status mongod` (Linux)
- Verify connection string in `.env`
- Try connecting with mongosh: `mongosh mongodb://localhost:27017/expense-tracker`

### Firebase Authentication Error
- Verify `firebase-service-account.json` is in the backend directory
- Check file path in `.env` is correct
- Ensure Firebase project is set up correctly
- Test token generation in frontend

### Port Already in Use
- Change PORT in `.env` to another port (e.g., 5001)
- Or kill process using port 5000: `lsof -ti:5000 | xargs kill -9` (macOS/Linux)

## 📚 Next Steps

1. **Connect Frontend**: Update frontend to use this API instead of localStorage
2. **Add Tests**: Write unit and integration tests
3. **Add Validation**: Use Joi or express-validator for input validation
4. **Add Logging**: Use Winston or Morgan for better logging
5. **Deploy**: Deploy to Heroku, Railway, or AWS

## 🔗 Related Documentation

- [Express.js](https://expressjs.com/)
- [Mongoose](https://mongoosejs.com/)
- [Firebase Admin SDK](https://firebase.google.com/docs/admin/setup)
- [MongoDB](https://www.mongodb.com/docs/)

## 📝 License

MIT
