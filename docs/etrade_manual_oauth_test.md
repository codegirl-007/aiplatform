# Manual Testing Instructions for E*TRADE OAuth (COD-14)

## Status

**OAuth signature issue FIXED** ✅ (2026-02-14)
- Request token step now succeeds
- Ready for full manual verification

## Prerequisites

1. E*TRADE developer account (https://developer.etrade.com/)
2. Sandbox credentials (Consumer Key + Consumer Secret)
3. E*TRADE sandbox account for login

## Setup

1. Create `.env` file in project root:

```bash
ETRADE_CONSUMER_KEY=your_sandbox_consumer_key_here
ETRADE_CONSUMER_SECRET=your_sandbox_consumer_secret_here
ETRADE_SANDBOX=true
```

## Test Procedure

### Test 1: Initial OAuth Flow (No Saved Token)

**Expected behavior**: Full OOB OAuth flow with user interaction

```bash
go run ./cmd/etrade-oauth-test
```

**Expected output**:

```
No saved token found or token expired. Starting OAuth flow...

Please visit the following URL to authorize the application:
https://us.etrade.com/e/t/etws/authorize?key=<consumer_key>&token=<request_token>

After authorizing, enter the verification code: _
```

**Steps**:
1. Copy the authorization URL from terminal
2. Open URL in browser
3. Log in to E*TRADE sandbox account
4. Click "Accept" to authorize the application
5. Copy the 5-character verification code displayed
6. Paste verification code into terminal and press Enter

**Expected output (continued)**:

```
Access token obtained successfully
Token saved to: /path/to/workspace/.aiplatform/credentials/etrade_tokens.json

Testing API call: GET /v1/accounts/list
Response status: 200 OK
Response body:
<AccountListResponse>
  <Accounts>
    <Account>
      <accountId>...</accountId>
      <accountIdKey>...</accountIdKey>
      ...
    </Account>
  </Accounts>
</AccountListResponse>

OAuth flow completed successfully!
```

**Verification**:
- [ ] Authorization URL printed and opens in browser
- [ ] E*TRADE login page loads
- [ ] Verification code displayed after authorization
- [ ] Verification code accepted by tool
- [ ] Access token saved to `.aiplatform/credentials/etrade_tokens.json`
- [ ] API call to `/v1/accounts/list` succeeds with 200 status
- [ ] Response body contains XML with account data

### Test 2: Token Reuse (Saved Token Exists)

**Expected behavior**: Load saved token, skip OAuth flow, make API call immediately

```bash
go run ./cmd/etrade-oauth-test
```

**Expected output**:

```
Token loaded from: /path/to/workspace/.aiplatform/credentials/etrade_tokens.json

Testing API call: GET /v1/accounts/list
Response status: 200 OK
Response body:
<AccountListResponse>
  ...
</AccountListResponse>

OAuth flow completed successfully!
```

**Verification**:
- [ ] No authorization URL printed
- [ ] No user interaction required
- [ ] Token loaded from disk
- [ ] API call succeeds immediately

### Test 3: Expired Token Handling

**Expected behavior**: Detect expired token, restart OAuth flow

**Setup**: Manually edit `.aiplatform/credentials/etrade_tokens.json` and set `expires_at` to a past date:

```json
{
  "access_token": "...",
  "access_secret": "...",
  "expires_at": "2024-01-01T00:00:00Z",
  "sandbox": true
}
```

```bash
go run ./cmd/etrade-oauth-test
```

**Expected output**: Should behave like Test 1 (full OAuth flow)

**Verification**:
- [ ] Tool detects expired token
- [ ] Full OAuth flow runs (authorization URL printed)
- [ ] New token obtained and saved

### Test 4: Invalid Credentials

**Expected behavior**: Clear error message for authentication failure

**Setup**: Temporarily modify `.env` with invalid credentials:

```bash
ETRADE_CONSUMER_KEY=invalid_key
ETRADE_CONSUMER_SECRET=invalid_secret
ETRADE_SANDBOX=true
```

```bash
go run ./cmd/etrade-oauth-test
```

**Expected output**: Error message indicating authentication failure

**Verification**:
- [ ] Tool fails gracefully with clear error message
- [ ] No panic or crash
- [ ] Error indicates authentication problem

## Success Criteria

All tests must pass:
- ✅ Test 1: Initial OAuth flow works end-to-end
- ✅ Test 2: Saved token reused correctly
- ✅ Test 3: Expired token detected and re-authenticated
- ✅ Test 4: Invalid credentials handled gracefully

## Troubleshooting

**"No saved token found" on second run**:
- Check that `.aiplatform/credentials/etrade_tokens.json` was created
- Verify file has correct JSON structure

**"Token expired" immediately after obtaining**:
- E*TRADE tokens expire at midnight US Eastern time
- If testing late in the day, token may expire quickly

**"401 Unauthorized" on API call**:
- Verify E*TRADE sandbox account is active
- Check that consumer key/secret are for sandbox environment
- Verify `ETRADE_SANDBOX=true` in `.env`

**Browser doesn't open automatically**:
- This is expected (OOB flow requires manual copy/paste)
- Copy URL from terminal and paste into browser

## Cleanup

After testing, you can remove the saved token:

```bash
rm -rf .aiplatform/credentials/
```

## Next Steps After Successful Testing

1. Mark COD-14 as "Done" in Linear
2. Update Linear ticket with test results
3. Move to COD-15 (documentation) or COD-16 (secure token storage)
