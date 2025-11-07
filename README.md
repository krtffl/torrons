# Torrorèndum 2025 - Torrons Vicens

Torrorèndum is an interactive voting application built with **Go** and **HTMX** that uses the **ELO rating system** to determine the best Torrons Vicens nougat products based on user preferences.

## 🎯 Overview

This application allows users to vote on pairwise comparisons of torrons (Spanish nougat) to create personalized and global rankings. With over 100 torron varieties, the system uses sophisticated algorithms to provide meaningful results with minimal voting requirements.

## ✨ Key Features

### 1. **Cookie-Based User Tracking**
- Automatic user identification without requiring login
- 90-day persistent cookies for returning users
- Privacy-preserving anonymous tracking
- Vote history and statistics per user

### 2. **Dual ELO Rating System**
- **Global ELO**: Community-wide ratings visible to all
- **Personalized ELO**: Individual user ratings based on their voting history
- K-factor of 42 for optimal rating convergence
- Transaction-based updates ensuring data consistency

### 3. **Personalized Leaderboards**
- Each user sees rankings tailored to their preferences
- Minimum vote requirements per category (25-50 votes)
- Category-specific and global personalized rankings
- Real-time updates after each vote

### 4. **Campaign Management & Countdown**
- Time-bound voting campaigns with start/end dates
- Countdown timer to results reveal (January 6, 2026)
- Campaign status management (active, ended, archived)
- Historical campaign data preservation

### 5. **Smart Category System**
Four traditional categories plus cross-category comparison:
- **Clàssics** - Traditional favorites (min. 30 votes)
- **Novetats** - Revolutionary flavors (min. 25 votes)
- **Xocolata** - Chocolate vs nougat boundary (min. 30 votes)
- **Albert Adrià** - Essència Adrià collection (min. 40 votes)
- **Global** - Cross-category absolute rankings (min. 50 votes)

### 6. **Strategic Pairing Algorithm**
- Regular categories: All possible pairings (O(n²))
- Global category: Smart pairing strategy (O(n*k))
  - Each torron paired with top 5 from other categories
  - Ensures ELO convergence without explosion of comparisons
- Cryptographically secure randomization
- Discontinued products excluded

### 7. **Modern Responsive UI**
- External CSS with design tokens and variables
- Mobile-first responsive design
- Smooth animations and transitions
- Accessibility features (ARIA labels, keyboard navigation)
- Progress tracking visual feedback
- Reduced motion support

## 🏗️ Architecture

### Technology Stack
- **Backend**: Go 1.24.7
- **Router**: Chi v5
- **Database**: PostgreSQL with golang-migrate
- **Frontend**: HTMX 1.9.9
- **Styling**: Custom CSS with CSS variables
- **Fonts**: Google Fonts (Montserrat)

### Database Schema
- **Users**: Anonymous user tracking with vote statistics
- **Campaigns**: Time-bound voting periods
- **UserEloSnapshots**: Personalized ratings per user per torron
- **Torrons**: Extended product information (allergens, dietary attributes)
- **Pairings**: Strategic matchups for voting
- **Results**: Vote history with user and campaign links

### API Endpoints

#### User API
- `GET /api/user/stats` - User voting statistics
- `GET /api/user/leaderboard/class/{classId}` - Personalized class leaderboard
- `GET /api/user/leaderboard/global` - Personalized global leaderboard

#### Campaign API
- `GET /api/campaign/countdown` - Time remaining until results reveal
- `GET /api/campaign/info` - Active campaign information
- `GET /api/leaderboard/global` - Global community leaderboard
- `GET /api/leaderboard/class/{classId}` - Class-specific global leaderboard

## 🚀 How It Works

1. **User Arrives**: Automatic cookie-based identification creates or retrieves user profile

2. **Category Selection**: User chooses from 5 categories (4 traditional + 1 global)

3. **Pairing Presentation**: System presents two torrons using secure random selection

4. **Vote Submission**: User selects preferred torron

5. **Dual Rating Update**:
   - Global ELO updated for community ranking
   - Personal ELO updated for user-specific ranking
   - Both operations in single atomic transaction

6. **Progress Tracking**: Visual progress bar encourages minimum votes

7. **Results Access**: Users meeting minimum votes see personalized leaderboards

## 📊 ELO System

### Parameters
- **K-factor**: 42 (optimal for food preference convergence)
- **Initial Rating**: 1500 (inherited from global rating for new user snapshots)
- **Update Formula**: Standard ELO with expected score calculation

### Minimum Vote Requirements
Statistical significance thresholds per category:
- Clàssics: 30 votes
- Novetats: 25 votes
- Xocolata: 30 votes
- Albert Adrià: 40 votes
- Global: 50 votes

## 🔒 Security Features

- Cryptographically secure random pairing selection
- Environment variable configuration for sensitive data
- Enhanced security headers (CSP, X-Frame-Options, HSTS-ready)
- SQL injection prevention via parameterized queries
- XSS protection through proper escaping
- Rate limiting (100 requests/minute per IP)
- Cookie security (httpOnly, secure, SameSite=Lax)

## 🌍 Environment Configuration

```bash
# Server
PORT=3000

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=myUser
DB_PASSWORD=myPassword
DB_NAME=databaseName
DB_SSL_MODE=disable

# Logging
LOGGER_FORMAT=json
LOGGER_LEVEL=info
LOGGER_PATH=logs/torro.log
```

See `.env.example` for full configuration template.

## 📦 Deployment

### Recommended Platforms
1. **Railway.app** (simplest, $10-20/month)
2. **Fly.io** (EU hosting, similar pricing)

See `docs/DEPLOYMENT_OPTIONS.md` for detailed comparison.

## 🎨 Brand Guidelines

Updated for 2025 with Torrons Vicens branding:
- Primary Color: `#4E0011` (Burgundy)
- Typography: Montserrat (300, 400, 500, 700 weights)
- Design: Modern, accessible, mobile-first

## 📝 Future Enhancements

### Pending: 2025 Inventory Update
Waiting for client to provide:
- Current 2025 product catalog
- New product images
- Updated descriptions and attributes
- Discontinued product list

See `docs/INVENTORY_REQUIREMENTS.md` for data format specifications.

## 🤝 Contributing

This is a private project for Torrons Vicens. For questions or suggestions, contact the development team.

## 📄 License

Proprietary - © 2025 Torrons Vicens

---

Built with ❤️ for Torrons Vicens - Torrorèndum 2025
