# Local Testing Guide - Torrorèndum 2025

This guide will help you set up and test all the new features locally before deploying.

## 🚀 Quick Start (5 minutes)

### Prerequisites

- Go 1.24.7+ installed
- PostgreSQL 12+ installed and running
- Port 3000 and 5432 available

### Step 1: Set Up Database

```bash
# Create database
createdb torrons

# Or using psql
psql -U postgres
CREATE DATABASE torrons;
\q
```

### Step 2: Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your database credentials
# Minimal config for local testing:
PORT=3000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=torrons
DB_SSL_MODE=disable
LOGGER_FORMAT=common
LOGGER_LEVEL=info
```

### Step 3: Start the Application (Migrations Run Automatically!)

```bash
# The server now runs migrations automatically on startup!
make run

# Or run directly
go run cmd/server/main.go

# To skip automatic migrations (production use):
go run cmd/server/main.go --skip-migrations
```

**✨ NEW:** Migrations now run automatically on server startup! No manual migration step needed.

### Step 4: Open in Browser

```
http://localhost:3000
```

---

## 📦 Migration Commands (Optional)

If you need to run migrations manually:

```bash
# Run all pending migrations
make migrate

# Rollback last migration
make migrate-down

# Create a new migration
make migrate-create

# Check current version
make migrate-version

# See all available commands
make help
```

---

## 🧪 Testing Checklist

### ✅ New Features to Test

#### 1. **Homepage & Navigation** (`http://localhost:3000`)
- [ ] Countdown widget displays and updates
- [ ] "How It Works" section shows 4 steps
- [ ] Category icons visible (🏛️ ✨ 🍫 👨‍🍳 🌍)
- [ ] Navigation bar shows: Estadístiques, Historial, Classificació, Vota
- [ ] Onboarding modal appears on first visit
- [ ] Modal can be dismissed and doesn't reappear

#### 2. **Onboarding Modal** (First visit)
- [ ] Modal auto-shows after 500ms
- [ ] Shows 4 steps with emojis
- [ ] "Comença a votar!" button works
- [ ] Overlay click closes modal
- [ ] X button closes modal
- [ ] Modal doesn't reappear after closing
- [ ] Mobile responsive (test on narrow screen)

#### 3. **Voting Flow** (`/classes` → Vote)
- [ ] Select a category
- [ ] Vote for torrons (click or press Enter/Space)
- [ ] Loading states appear (top bar, button spinner)
- [ ] Next comparison loads smoothly
- [ ] Achievement toast appears at 1st, 5th, 10th vote
- [ ] Progress tracking updates
- [ ] Keyboard navigation works (Tab, Enter, Space)

#### 4. **Achievement Toasts** (During voting)
- [ ] Toast at 1 vote: "🎯 Primer vot! Continua així!"
- [ ] Toast at 5 votes: "🔓 5 vots! Has desbloqueejat els resultats..."
- [ ] Toast at 10 votes: "⭐ 10 vots! Torronaire entusiasta!"
- [ ] Toasts slide in from right
- [ ] Auto-dismiss after 4 seconds
- [ ] Multiple toasts stack properly
- [ ] Mobile positioning correct

#### 5. **Stats Dashboard** (`/stats`)
- [ ] Total votes count displays
- [ ] Unlocked categories count shows
- [ ] User rank displays (Principiant → Mestre torronaire)
- [ ] Category progress bars show percentage
- [ ] Votes remaining displays correctly
- [ ] Locked categories show "🔒 Bloquejat"
- [ ] Achievement badges show locked/unlocked state
- [ ] CTAs work ("Vota ara", "Veure resultats")

#### 6. **Voting History** (`/history`)
- [ ] Shows list of past votes
- [ ] Winner highlighted in green
- [ ] Loser appears grayed out
- [ ] "✓ Escollit" badge on winner
- [ ] Category filter buttons work
- [ ] "Tots" shows all votes
- [ ] Time ago format displays ("Fa 2 hores")
- [ ] "Load more" pagination works
- [ ] Empty state for new users

#### 7. **Leaderboard** (`/leaderboard`)
- [ ] Personal view toggle works
- [ ] Global view toggle works
- [ ] Category filter buttons work
- [ ] Rank changes show (↑ ↓ — NOU)
- [ ] Position change numbers display
- [ ] Top 3 medals show (🥇 🥈 🥉)
- [ ] Progress bars fill correctly
- [ ] Share section appears in personal view
- [ ] "No tens prou vots" error shows correctly

#### 8. **Rank Change Tracking** (Leaderboard personal view)
Test this by:
1. Vote 5+ times in a category
2. View leaderboard → note positions
3. Clear cookies or use incognito
4. View leaderboard again → all should show "NOU"
5. Vote more and revisit → should show ↑ or ↓

- [ ] First visit shows "NOU" (orange badge)
- [ ] Position improvements show "↑" (green)
- [ ] Position drops show "↓" (red)
- [ ] No change shows "—" (gray)
- [ ] Number of positions displayed ("↑ 3")

#### 9. **Social Share** (Leaderboard personal view, after 5+ votes)
- [ ] Share section visible
- [ ] Twitter button opens with pre-filled text
- [ ] Facebook button opens with share dialog
- [ ] WhatsApp button opens with message
- [ ] Copy link button works
- [ ] "✓ Copiat!" feedback shows
- [ ] Share text includes top 3 torrons with medals
- [ ] URLs are properly encoded

#### 10. **Accessibility**
Use keyboard only:
- [ ] Tab through navigation links
- [ ] Tab through voting cards
- [ ] Enter/Space activates cards
- [ ] Tab through filter buttons
- [ ] Focus indicators visible (burgundy outline)
- [ ] All interactive elements reachable
- [ ] Screen reader announces content correctly
- [ ] ARIA labels present

Test with screen reader (optional):
- [ ] Navigation announces roles
- [ ] Voting instructions read aloud
- [ ] Leaderboard entries meaningful

#### 11. **Loading States**
- [ ] Global loading bar at top during requests
- [ ] Button spinners during vote submission
- [ ] Card overlays during updates
- [ ] Smooth transitions between states

#### 12. **SEO & Meta Tags**
View page source:
- [ ] Title tag: "Torrorèndum 2025 - Vota el millor torró | Torrons Vicens"
- [ ] Description meta tag present
- [ ] Open Graph tags present (og:title, og:image, etc.)
- [ ] Twitter Card tags present
- [ ] Canonical URL set

---

## 🐛 Common Issues & Solutions

### Database Connection Failed
```bash
# Check PostgreSQL is running
pg_isready

# Check credentials in .env
# Make sure DB_NAME exists
psql -U postgres -l | grep torrons
```

### Port 3000 Already in Use
```bash
# Find process using port
lsof -i :3000

# Kill it or change PORT in .env
PORT=3001
```

### Migrations Failed
```bash
# Reset database
psql -U postgres
DROP DATABASE torrons;
CREATE DATABASE torrons;
\q

# Run migrations again
migrate -path migrations -database "postgresql://..." up
```

### No Torrons Appearing
```bash
# Check if data was seeded
psql -U postgres -d torrons
SELECT COUNT(*) FROM "Torrons";
\q

# Should show 100+ torrons
```

### Achievement Toasts Not Appearing
- Clear localStorage: Open DevTools → Application → Local Storage → Clear
- Refresh page and vote again

### Rank Changes Not Showing
- Clear cookies: DevTools → Application → Cookies → Clear
- Visit leaderboard → all should show "NOU"

---

## 🎯 Performance Testing

### Test Image Loading
```bash
# Check network tab in DevTools
# Images over 200KB will load slowly on 3G

# To optimize now:
./optimize-images.sh
```

### Test Mobile Experience
1. Open DevTools (F12)
2. Toggle device toolbar (Ctrl+Shift+M)
3. Test on:
   - iPhone SE (375px)
   - iPhone 12 Pro (390px)
   - iPad (768px)
   - Desktop (1920px)

### Test Loading Speed
```bash
# Use lighthouse (in Chrome DevTools)
# Target scores:
# - Performance: 80+
# - Accessibility: 95+
# - Best Practices: 90+
# - SEO: 95+
```

---

## 📊 Sample Test Scenarios

### Scenario 1: New User Journey
1. Open in incognito mode
2. Should see onboarding modal
3. Close modal, click "Vota"
4. Select "Clàssics"
5. Vote 5 times
6. Check toast at 1st and 5th vote
7. Go to /stats → see progress
8. Go to /leaderboard → see personal results
9. Check share section appears

### Scenario 2: Returning User
1. Open with cookies from Scenario 1
2. No onboarding modal
3. Vote in different category
4. Check history page shows all votes
5. Filter history by category
6. Go to leaderboard
7. Note rank positions
8. Vote more
9. Return to leaderboard
10. Check rank changes show ↑ or ↓

### Scenario 3: Accessibility User
1. Close all pointing devices
2. Tab through entire site
3. Use Enter/Space to interact
4. Complete a full voting cycle
5. Navigate to all pages
6. Verify all functionality accessible

---

## 🔍 Developer Tools Tips

### Check for JavaScript Errors
```
F12 → Console
Should be clean except for INFO logs
```

### Monitor Network Requests
```
F12 → Network → Filter by XHR
Watch HTMX requests
Should be fast (<100ms)
```

### Check Cookies
```
F12 → Application → Cookies
Should see:
- torrons_user_id (90 days)
- onboarding_seen (1 year)
- ranks_global (30 days)
- ranks_1, ranks_2, etc. (30 days per category)
```

### Check Local Storage
```
F12 → Application → Local Storage
Should see:
- vote_count (for achievement tracking)
```

---

## ✅ Sign-Off Checklist

Before creating PR, verify:

- [ ] All 12 feature categories tested
- [ ] No JavaScript console errors
- [ ] All navigation links work
- [ ] Mobile responsive on 3+ screen sizes
- [ ] Keyboard navigation works completely
- [ ] No database errors in server logs
- [ ] Images load properly
- [ ] Cookies persist across sessions
- [ ] Achievement toasts trigger correctly
- [ ] Rank tracking updates properly

---

## 🆘 Need Help?

1. **Check server logs**: Look in console output for errors
2. **Check browser console**: F12 → Console for JS errors
3. **Check database**: `psql -U postgres -d torrons` to query data
4. **Clear everything**: Cookies + LocalStorage + restart server

---

## 🚀 After Testing

Once everything works:

```bash
# Commit any fixes
git add .
git commit -m "fix: any issues found during testing"
git push

# Create PR
# See main README for PR template
```

---

**Happy Testing!** 🎉

If all features work as expected, you're ready to deploy to production!
