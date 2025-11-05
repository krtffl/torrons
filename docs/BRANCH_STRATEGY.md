# Branch Strategy - Torrons Project

**Project:** Torrons Vicens 2025 Campaign
**Timeline:** November 5 - November 30, 2025 (3.5 weeks)
**Deployment:** End of November 2025
**Campaign Launch:** Early December 2025
**Results Release:** January 6, 2026

---

## 🌿 Branch Structure

### **Main Branches**

```
main (protected)
├── develop (integration branch)
└── production (deployment branch)
```

### **Feature Branches**

Each feature gets its own branch for parallel development:

```
feature/database-migrations       - Schema updates, new tables
feature/user-tracking            - Cookie-based user identification
feature/personalized-elo         - User-specific ELO calculations
feature/countdown-results        - Campaign countdown & results page
feature/global-category          - Cross-category comparisons
feature/vote-requirements        - Minimum vote thresholds
feature/ui-redesign              - CSS/template overhaul
feature/2025-inventory           - Product data updates
feature/security-hardening       - Security improvements
```

---

## 🔄 Development Workflow

### **Branch Creation Pattern**

```bash
# Create feature branch from develop
git checkout develop
git pull origin develop
git checkout -b feature/[feature-name]

# Example:
git checkout -b feature/database-migrations
```

### **Development Cycle**

```
1. Create feature branch
2. Implement feature with tests
3. Commit regularly with clear messages
4. Push to remote branch
5. Create Pull Request to develop
6. Code review (if needed)
7. Merge to develop
8. Delete feature branch
```

### **Commit Message Convention**

```
feat: Add user tracking middleware
fix: Correct ELO calculation for edge case
refactor: Extract pairing logic to separate function
docs: Update inventory requirements
test: Add unit tests for countdown logic
style: Apply consistent formatting to templates
chore: Update dependencies

Examples:
feat(user): implement cookie-based identification
fix(elo): handle division by zero in rating calculation
refactor(db): extract transaction logic to helper
test(pairing): add tests for global category selection
```

---

## 📅 Feature Branch Timeline

### **Week 1: November 5-11 (Foundation)**

**Priority: CRITICAL - Must complete for others to build on**

#### **feature/database-migrations** (Nov 5-7)
```
Tasks:
- Create Users table
- Create Campaigns table
- Create UserVotes table
- Create UserEloSnapshots table
- Add Timestamp to Results
- Add extended fields to Torrons (Description, Allergens, etc.)
- Create migrations (up and down)
- Test locally

Dependencies: None
Estimated: 2-3 days
Merge to develop by: Nov 7
```

#### **feature/security-hardening** (Nov 5-6)
```
Tasks:
- Move DB credentials to environment variables
- Add .env file support
- Initialize random seed properly
- Add HSTS header
- Update CSP
- Add security documentation

Dependencies: None
Estimated: 1-2 days
Merge to develop by: Nov 6
```

#### **feature/user-tracking** (Nov 7-9)
```
Tasks:
- Create user identification middleware
- Generate/validate user cookies
- Track user sessions
- Link votes to users
- Add user repository layer
- Test cookie persistence

Dependencies: feature/database-migrations
Estimated: 2-3 days
Merge to develop by: Nov 9
```

---

### **Week 2: November 12-18 (Core Features)**

**Priority: HIGH - Main functionality**

#### **feature/2025-inventory** (Nov 12-14)
```
Tasks:
- Wait for inventory data collection
- Create data migration scripts
- Add new products
- Mark discontinued products
- Upload new images
- Validate all pairings generate
- Test image loading

Dependencies: Inventory data from client, feature/database-migrations
Estimated: 2-3 days
Merge to develop by: Nov 14
```

#### **feature/personalized-elo** (Nov 10-13)
```
Tasks:
- Implement user-specific ELO tracking
- Create UserEloSnapshots repository
- Update vote handler to calculate user ELO
- Build personalized leaderboard logic
- Add API endpoint for user rankings
- Test ELO calculations

Dependencies: feature/user-tracking, feature/database-migrations
Estimated: 3-4 days
Merge to develop by: Nov 13
```

#### **feature/vote-requirements** (Nov 13-15)
```
Tasks:
- Implement vote counting per user per category
- Add minimum vote thresholds
- Create "unlock" logic for results
- Update progress bar with thresholds
- Add vote stats API
- Show "X more votes needed" messaging

Dependencies: feature/user-tracking, feature/personalized-elo
Estimated: 2-3 days
Merge to develop by: Nov 15
```

#### **feature/countdown-results** (Nov 14-17)
```
Tasks:
- Create Campaigns management
- Build countdown timer component
- Implement global leaderboard
- Add category-specific results
- Create results page template
- Add countdown API endpoint
- Test countdown expiry behavior

Dependencies: feature/database-migrations
Estimated: 3-4 days
Merge to develop by: Nov 17
```

---

### **Week 3: November 19-25 (Enhancement & Polish)**

**Priority: MEDIUM-HIGH - Value-add features**

#### **feature/global-category** (Nov 16-19)
```
Tasks:
- Add Global category to database
- Implement smart pairing algorithm
- Balance pairing across all categories
- Add global leaderboard view
- Test with large pairing set
- Optimize query performance

Dependencies: feature/2025-inventory, feature/database-migrations
Estimated: 3-4 days
Merge to develop by: Nov 19
```

#### **feature/ui-redesign** (Nov 18-24)
```
Tasks:
- Extract brand guidelines from Vicens website (manual)
- Create external CSS files
- Refactor templates (remove inline styles)
- Add animations and transitions
- Implement loading states
- Add engagement features (confetti, etc.)
- Mobile optimization
- Accessibility improvements (ARIA, keyboard nav)
- Test across devices

Dependencies: feature/countdown-results, feature/vote-requirements
Estimated: 5-7 days
Merge to develop by: Nov 24
```

---

### **Week 4: November 26-30 (Testing & Deployment)**

**Priority: CRITICAL - Launch preparation**

#### **Integration Testing** (Nov 25-27)
```
Tasks:
- Merge all features to develop
- End-to-end testing
- Load testing
- Security testing
- Bug fixes
- Performance optimization

Branch: develop
Estimated: 3 days
```

#### **Production Deployment** (Nov 28-30)
```
Tasks:
- Merge develop to production
- Set up production environment
- Configure SSL/HTTPS
- Set environment variables
- Database migration in production
- Deploy application
- Post-deployment verification
- Monitoring setup

Branch: production
Estimated: 2-3 days
Complete by: Nov 30
```

---

## 🔀 Merge Strategy

### **Feature → Develop**

```bash
# Update feature branch with latest develop
git checkout feature/your-feature
git fetch origin
git rebase origin/develop

# Push updated feature branch
git push origin feature/your-feature --force-with-lease

# Create Pull Request on GitHub/GitLab
# After approval, merge using "Squash and Merge" for clean history
```

### **Develop → Production**

```bash
# Only after all features merged and tested
git checkout production
git merge develop --no-ff -m "Release: November 2025 - All features"
git push origin production
```

### **Hotfix Workflow**

If critical bugs found after deployment:

```bash
git checkout production
git checkout -b hotfix/issue-description
# Fix the issue
git commit -m "hotfix: description of fix"
git checkout production
git merge hotfix/issue-description
git checkout develop
git merge hotfix/issue-description
```

---

## 🔐 Branch Protection Rules

### **main**
- Protected: YES
- Require pull request: YES
- Require approvals: 1 (if team, otherwise can merge own)
- Dismiss stale reviews: YES
- Allow force push: NO

### **develop**
- Protected: YES
- Require pull request: YES
- Allow force push: NO

### **production**
- Protected: YES
- Require pull request: YES
- Require approvals: 1
- Allow force push: NO
- Only allow merges from: develop

### **feature/**
- Protected: NO
- Allow force push: YES (for rebasing)

---

## 📋 Branch Status Tracking

### **Week 1 Checklist**

- [ ] feature/database-migrations (Nov 7)
- [ ] feature/security-hardening (Nov 6)
- [ ] feature/user-tracking (Nov 9)

### **Week 2 Checklist**

- [ ] feature/2025-inventory (Nov 14)
- [ ] feature/personalized-elo (Nov 13)
- [ ] feature/vote-requirements (Nov 15)
- [ ] feature/countdown-results (Nov 17)

### **Week 3 Checklist**

- [ ] feature/global-category (Nov 19)
- [ ] feature/ui-redesign (Nov 24)

### **Week 4 Checklist**

- [ ] Integration testing complete (Nov 27)
- [ ] Production deployment (Nov 30)

---

## 🚀 Parallel Development Strategy

### **Can Work Simultaneously:**

**Group A (No Dependencies):**
- feature/database-migrations
- feature/security-hardening

**Group B (After database-migrations):**
- feature/user-tracking
- feature/countdown-results

**Group C (After user-tracking):**
- feature/personalized-elo
- feature/vote-requirements

**Group D (After inventory data):**
- feature/2025-inventory
- feature/global-category (after inventory)

**Group E (After most features):**
- feature/ui-redesign (can start earlier with planning)

### **Critical Path:**

```
database-migrations → user-tracking → personalized-elo → vote-requirements
                   → countdown-results
                   → 2025-inventory → global-category
                   → ui-redesign → integration → deployment
```

**Total Critical Path: ~22-25 days** (fits within 3.5 week deadline with buffer)

---

## 🎯 Daily Standup Format

Track progress daily:

```
What was completed yesterday?
- Merged feature/database-migrations
- Started feature/user-tracking

What will be done today?
- Complete cookie middleware
- Test user session persistence

Any blockers?
- Waiting for inventory data (Nov 12 deadline)
- Need brand guidelines from website review
```

---

## 📊 Progress Tracking

Use GitHub/GitLab Issues and Project Board:

```
Columns:
- Backlog
- In Progress
- In Review
- Testing
- Done

Labels:
- priority: critical
- priority: high
- priority: medium
- type: feature
- type: bug
- type: refactor
- blocked
- needs-review
```

---

## 🔧 Local Development Setup

```bash
# Clone repository
git clone <repo-url>
cd torrons

# Checkout develop branch
git checkout develop

# Pull latest changes
git pull origin develop

# Create .env file (after feature/security-hardening)
cp .env.example .env
# Edit .env with your local database credentials

# Run migrations
make migrate-up

# Start server
make run
```

---

## 📝 Notes

1. **Keep branches short-lived:** Max 3-5 days before merging
2. **Rebase regularly:** Keep feature branches up-to-date with develop
3. **Test before merging:** All tests must pass
4. **Small commits:** Commit early and often
5. **Clear messages:** Future you will thank you
6. **Review before merge:** Even if self-review
7. **Delete after merge:** Clean up merged feature branches

---

## ⚠️ Risk Mitigation

### **If falling behind schedule:**

1. **Priority 1 (Must Have):**
   - User tracking
   - Personalized ELO
   - Countdown/Results
   - 2025 Inventory

2. **Priority 2 (Should Have):**
   - Vote requirements
   - Global category

3. **Priority 3 (Nice to Have):**
   - Full UI redesign (can do minimal styling if needed)
   - Advanced engagement features

**Decision Point: November 20**
- If behind, cut Priority 3 features
- Focus on core functionality + basic styling
- Can enhance UI post-launch

---

This strategy allows for **parallel development while maintaining stability and clear dependencies**. Let's build this! 🚀
