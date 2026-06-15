#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
PASSED=0
FAILED=0

# Generate unique email for this test run
RANDOM_SUFFIX=$(date +%s)
TEST_EMAIL="apitest_${RANDOM_SUFFIX}@example.com"
TEST_PASSWORD="Password1"

assert_status() {
    local test_name="$1"
    local expected="$2"
    local actual="$3"

    if [ "$actual" -eq "$expected" ]; then
        echo "PASS: $test_name (status $actual)"
        PASSED=$((PASSED + 1))
    else
        echo "FAIL: $test_name (expected $expected, got $actual)"
        FAILED=$((FAILED + 1))
    fi
}

echo "=== API Tests ==="
echo "Base URL: $BASE_URL"
echo ""

# Health check
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")
assert_status "GET /health" 200 "$STATUS"

# Register
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
assert_status "POST /register" 201 "$STATUS"

# Register duplicate
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
assert_status "POST /register duplicate" 400 "$STATUS"

# Register weak password
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/register" \
    -H "Content-Type: application/json" \
    -d '{"email":"weak@example.com","password":"weak"}')
assert_status "POST /register weak password" 400 "$STATUS"

# Register invalid email
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/register" \
    -H "Content-Type: application/json" \
    -d '{"email":"not-an-email","password":"Password1"}')
assert_status "POST /register invalid email" 400 "$STATUS"

# Login
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
STATUS=$(echo "$LOGIN_RESPONSE" | head -c 0; curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}")
assert_status "POST /login" 200 "$STATUS"

TOKEN=$(curl -s -X POST "$BASE_URL/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"$TEST_PASSWORD\"}" | \
    python3 -c "import sys,json; print(json.load(sys.stdin)['token'])" 2>/dev/null || echo "")

if [ -z "$TOKEN" ]; then
    echo "FAIL: Could not extract token from login response"
    FAILED=$((FAILED + 1))
    echo ""
    echo "=== Results: $PASSED passed, $FAILED failed ==="
    exit 1
fi

# Login wrong password
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\",\"password\":\"WrongPass1\"}")
assert_status "POST /login wrong password" 401 "$STATUS"

# Me
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/me" \
    -H "Authorization: Bearer $TOKEN")
assert_status "GET /me" 200 "$STATUS"

# Unauthorized access
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/me")
assert_status "GET /me unauthorized" 401 "$STATUS"

# Create notification
CREATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/notifications" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"title":"Test Title","message":"Test Message"}')
STATUS=$(echo "$CREATE_RESPONSE" | tail -1)
BODY=$(echo "$CREATE_RESPONSE" | head -n -1)
assert_status "POST /notifications" 201 "$STATUS"

NOTIFICATION_ID=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null || echo "0")

# Create notification empty title
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/notifications" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"title":"","message":"Test Message"}')
assert_status "POST /notifications empty title" 400 "$STATUS"

# List notifications
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/notifications" \
    -H "Authorization: Bearer $TOKEN")
assert_status "GET /notifications" 200 "$STATUS"

# Get notification by ID
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/notifications/$NOTIFICATION_ID" \
    -H "Authorization: Bearer $TOKEN")
assert_status "GET /notifications/:id" 200 "$STATUS"

# Mark as read
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "$BASE_URL/notifications/$NOTIFICATION_ID/read" \
    -H "Authorization: Bearer $TOKEN")
assert_status "PUT /notifications/:id/read" 200 "$STATUS"

# Delete notification
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/notifications/$NOTIFICATION_ID" \
    -H "Authorization: Bearer $TOKEN")
assert_status "DELETE /notifications/:id" 200 "$STATUS"

# Create webhook
CREATE_WH_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/webhooks" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"url":"https://example.com/webhook"}')
STATUS=$(echo "$CREATE_WH_RESPONSE" | tail -1)
BODY=$(echo "$CREATE_WH_RESPONSE" | head -n -1)
assert_status "POST /webhooks" 201 "$STATUS"

WEBHOOK_ID=$(echo "$BODY" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null || echo "0")

# Create webhook invalid URL
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/webhooks" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"url":"not-a-url"}')
assert_status "POST /webhooks invalid URL" 400 "$STATUS"

# List webhooks
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/webhooks" \
    -H "Authorization: Bearer $TOKEN")
assert_status "GET /webhooks" 200 "$STATUS"

# Delete webhook
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/webhooks/$WEBHOOK_ID" \
    -H "Authorization: Bearer $TOKEN")
assert_status "DELETE /webhooks/:id" 200 "$STATUS"

# Admin endpoints - should be forbidden for regular user
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/admin/users" \
    -H "Authorization: Bearer $TOKEN")
assert_status "GET /admin/users forbidden" 403 "$STATUS"

STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/admin/notifications" \
    -H "Authorization: Bearer $TOKEN")
assert_status "GET /admin/notifications forbidden" 403 "$STATUS"

# X-Request-ID header present
REQUEST_ID=$(curl -s -D - -o /dev/null "$BASE_URL/health" | grep -i "x-request-id" | tr -d '\r' | awk '{print $2}')
if [ -n "$REQUEST_ID" ]; then
    echo "PASS: X-Request-ID header present ($REQUEST_ID)"
    PASSED=$((PASSED + 1))
else
    echo "FAIL: X-Request-ID header not present"
    FAILED=$((FAILED + 1))
fi

echo ""
echo "=== Results: $PASSED passed, $FAILED failed ==="

if [ "$FAILED" -gt 0 ]; then
    exit 1
fi
