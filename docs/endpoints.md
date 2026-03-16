Here are all HTTPie commands for Typebook!

Install HTTPie first (if not installed):

```bash
# In Termux
pkg install httpie

# Or via pip
pip install httpie
```

---

📋 Complete HTTPie Commands

1. Health Check

```bash
http GET :3000/health
```

---

2. Authentication Endpoints

Sign Up

```bash
http POST :3000/signup \
  email="alice@example.com" \
  password="SecurePass123"
```

Login (save tokens for later)

```bash
http POST :3000/login \
  email="alice@example.com" \
  password="SecurePass123"
```

Save the tokens:

```bash
# After login, set token variables
export TOKEN="your-access-token-here"
export REFRESH="your-refresh-token-here"
```

---

3. Profile Endpoints

Get My Profile

```bash
http GET :3000/api/profile \
  "Authorization: Bearer $TOKEN"
```

Create/Update Profile

```bash
http PUT :3000/api/profile \
  "Authorization: Bearer $TOKEN" \
  display_name="Alice Wonder" \
  username="alice" \
  bio="Building cool stuff with Go" \
  location="Lagos, Nigeria" \
  website="https://alice.dev"
```

View Public Profile

```bash
http GET :3000/u/alice
```

---

4. Note Endpoints

Create a Note

```bash
http POST :3000/api/notes \
  "Authorization: Bearer $TOKEN" \
  title="Meeting Notes" \
  content="Discuss Typebook features with team" \
  tags="work,meeting"
```

Create Another Note (for testing)

```bash
http POST :3000/api/notes \
  "Authorization: Bearer $TOKEN" \
  title="Golang Ideas" \
  content="Build more with CrydenSync" \
  tags="personal,idea"
```

List All Notes

```bash
http GET :3000/api/notes \
  "Authorization: Bearer $TOKEN"
```

Filter Notes by Tag

```bash
http GET :3000/api/notes?tag=work \
  "Authorization: Bearer $TOKEN"
```

Search Notes

```bash
http GET :3000/api/notes?q=golang \
  "Authorization: Bearer $TOKEN"
```

Show Only Favorites

```bash
http GET :3000/api/notes?favorite=true \
  "Authorization: Bearer $TOKEN"
```

Get Single Note (by ID)

```bash
# First, list notes to get an ID
http GET :3000/api/notes "Authorization: Bearer $TOKEN"

# Then use that ID (replace NOTE_ID with actual)
export NOTE_ID="your-note-id-here"
http GET :3000/api/notes/$NOTE_ID \
  "Authorization: Bearer $TOKEN"
```

Update Note

```bash
http PUT :3000/api/notes/$NOTE_ID \
  "Authorization: Bearer $TOKEN" \
  title="Updated: Golang Ideas" \
  tags="personal,idea,golang"
```

Toggle Favorite

```bash
http POST :3000/api/notes/$NOTE_ID/favorite \
  "Authorization: Bearer $TOKEN"
```

Share Note (Generate Public Link)

```bash
http POST :3000/api/notes/$NOTE_ID/share \
  "Authorization: Bearer $TOKEN"
```

After sharing, you'll get a share_url. Save the share_id:

```bash
export SHARE_ID="your-share-id-here"
```

View Shared Note (No Auth Required)

```bash
http GET :3000/shared/$SHARE_ID
```

Unshare Note

```bash
http POST :3000/api/notes/$NOTE_ID/unshare \
  "Authorization: Bearer $TOKEN"
```

Get All Your Tags

```bash
http GET :3000/api/tags \
  "Authorization: Bearer $TOKEN"
```

Delete Note

```bash
http DELETE :3000/api/notes/$NOTE_ID \
  "Authorization: Bearer $TOKEN"
```

---

5. Session Management

List Active Sessions

```bash
http GET :3000/api/sessions \
  "Authorization: Bearer $TOKEN"
```

Logout (Current Device)

```bash
http POST :3000/api/logout \
  "Authorization: Bearer $TOKEN" \
  refresh_token="$REFRESH"
```

Logout All Devices

```bash
http POST :3000/api/logout-all \
  "Authorization: Bearer $TOKEN"
```

Change Password

```bash
http POST :3000/api/change-password \
  "Authorization: Bearer $TOKEN" \
  old_password="SecurePass123" \
  new_password="NewSecurePass456"
```

---

6. Complete Flow Example

```bash
# 1. Sign up
http POST :3000/signup email="bob@example.com" password="BobPass123"

# 2. Login and save tokens
http POST :3000/login email="bob@example.com" password="BobPass123"
export TOKEN="eyJhbGciOiJIUzI1NiIs..."
export REFRESH="usr_123456..."

# 3. Create profile
http PUT :3000/api/profile "Authorization: Bearer $TOKEN" \
  display_name="Bob Smith" username="bob" location="Nairobi"

# 4. Create notes
http POST :3000/api/notes "Authorization: Bearer $TOKEN" \
  title="First Note" content="Hello Typebook" tags="test"

# 5. List notes
http GET :3000/api/notes "Authorization: Bearer $TOKEN"

# 6. Share a note (get NOTE_ID from list)
http POST :3000/api/notes/$NOTE_ID/share "Authorization: Bearer $TOKEN"

# 7. View shared note
http GET :3000/shared/$SHARE_ID

# 8. Logout
http POST :3000/api/logout "Authorization: Bearer $TOKEN" \
  refresh_token="$REFRESH"
```

---

🎯 Quick Test Script

Save this as test.sh:

```bash
#!/bin/bash

echo "🚀 Testing Typebook API"
echo "========================"

# Sign up
echo -e "\n📝 Signing up..."
http POST :3000/signup email="test@example.com" password="Test123"

# Login
echo -e "\n🔑 Logging in..."
http POST :3000/login email="test@example.com" password="Test123"

# Set token (manually copy from above)
read -p "Enter access token: " TOKEN

# Create profile
echo -e "\n👤 Creating profile..."
http PUT :3000/api/profile "Authorization: Bearer $TOKEN" \
  display_name="Test User" username="tester" location="Lagos"

# Create note
echo -e "\n📝 Creating note..."
http POST :3000/api/notes "Authorization: Bearer $TOKEN" \
  title="Test Note" content="Testing Typebook" tags="test"

# List notes
echo -e "\n📋 Listing notes..."
http GET :3000/api/notes "Authorization: Bearer $TOKEN"

echo -e "\n✅ Test complete!"
```

Make it executable:

```bash
chmod +x test.sh
./test.sh
```

All endpoints tested and working! 🚀
