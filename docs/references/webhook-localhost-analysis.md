# Webhook Feasibility Analysis - Localhost Deployment

**Date**: 2025-11-16  
**Context**: User runs ceye locally on laptop, monitors ~10 repos (own repos only)

## Updated Assessment

### The Localhost Problem

**Webhooks fundamentally require**:
- GitHub/Azure must POST to your endpoint
- Requires publicly accessible HTTPS URL
- Localhost (127.0.0.1) is not accessible from internet

### Solutions for Localhost

#### Option 1: ngrok (Persistent Tunnel)
**Pros**:
- Works perfectly for webhooks
- Can get static subdomain ($8/month)
- Inspect all webhook traffic
- Professional solution

**Cons**:
- ❌ Requires ngrok to be running 24/7
- ❌ Paid plan needed for static URLs ($8/mo)
- ❌ If ngrok stops, webhooks stop working
- ❌ Free tier = random URLs that change on restart

**Verdict**: Works but adds operational complexity for local-only use

#### Option 2: Cloudflare Tunnel (Free)
**Pros**:
- ✅ Free forever
- ✅ Static subdomain
- ✅ Automatic reconnection
- ✅ Built-in SSL

**Cons**:
- Still requires tunnel to be running
- More setup than ngrok

#### Option 3: Tailscale + Webhook Proxy
**Complexity**: High, probably not worth it

### Reality Check: Do You Actually Need Webhooks?

Let's analyze your actual situation:

**Your Setup**:
- 10 repos max
- Own repos (full control)
- Laptop deployment (not 24/7 server)
- Local-only preference

**Current Polling Math**:
```
10 repos × 4 polls/minute = 40 requests/minute
40 × 60 minutes = 2,400 requests/hour

GitHub limit: 5,000 requests/hour
Your usage: 2,400 requests/hour (48% of limit)
```

**Conclusion**: You're only using ~50% of your rate limit quota!

### The Real Problem

If you're hitting rate limits with only 10 repos, the issue might be:

1. **Polling too frequently** (every 15 seconds is aggressive)
   - Solution: Increase interval to 30s → 1,200 requests/hour (24% of limit)
   
2. **Multiple instances running** (accidentally?)
   - Check: `ps aux | grep ci-dash`
   
3. **Other tools using the same token** (gh CLI, other apps)
   - Each tool shares the same 5,000/hour quota
   
4. **Unauthenticated requests** (60/hour limit)
   - Verify: `GITHUB_TOKEN` is set correctly

### Recommended Solution: Optimize Polling First

**Before webhooks, try this**:

1. **Increase poll interval** (30-60 seconds instead of 15)
2. **Smart polling** (only poll active repos)
3. **Rate limit monitoring** (log when approaching limits)
4. **Exponential backoff** (slow down when rate limited)

This is **much simpler** than webhooks + tunneling for local-only use.

### If You Still Want Webhooks

**Best option for localhost**: Use ngrok with static domain

**Setup**:
```bash
# One-time setup
ngrok config add-authtoken YOUR_TOKEN
ngrok http --domain=your-static-name.ngrok-free.app 9090

# Keep this running while ceye runs
# Set webhook URL: https://your-static-name.ngrok-free.app/webhooks/github
```

**Reality**:
- Adds ~$8/month cost
- Must keep ngrok running alongside ceye
- More complexity for marginal benefit with only 10 repos

### My Honest Recommendation

**Don't implement webhooks yet. Instead**:

1. **Optimize current polling** (see detailed plan below)
2. **Add rate limit monitoring** (know when you're actually hitting limits)
3. **Implement smart polling** (adaptive intervals based on activity)
4. **Consider webhooks later** when:
   - You're monitoring 50+ repos, OR
   - You deploy to a server/cloud, OR
   - You're willing to pay for ngrok static domain

With only 10 repos, optimized polling is simpler, cheaper, and "good enough."

---

## Alternative Plan: Optimize Polling Instead

### Phase 1: Rate Limit Awareness (Week 1)

**Goal**: Know when and why you're hitting limits

#### 1.1 Rate Limit Monitoring (2 days)

**Features**:
- Capture rate limit headers from GitHub API responses
- Log current limit usage
- Alert when approaching limits (> 80%)
- Display in TUI/Web UI

**GitHub Response Headers**:
```
X-RateLimit-Limit: 5000
X-RateLimit-Remaining: 4532
X-RateLimit-Reset: 1703001234
```

**Implementation**:
```go
type RateLimitInfo struct {
    Limit     int
    Remaining int
    Reset     time.Time
    Used      int
}

func (c *HTTPClient) GetRateLimitInfo() RateLimitInfo {
    // Parse from last response headers
}
```

**UI Addition**:
```
Rate Limits: 2,400 / 5,000 (48%) - Resets in 42m
```

#### 1.2 Request Logging (1 day)

**Track**:
- Total requests per hour
- Requests per repo
- Failed requests (rate limited)
- Identify which repo/workflow uses most quota

**Output**:
```
GitHub API Usage (last hour):
  joelklabo/ceye: 120 requests (5%)
  joelklabo/other: 80 requests (3%)
  Total: 200 requests (4% of quota)
```

#### 1.3 Testing (1 day)

- Verify rate limit parsing
- Test alert thresholds
- Simulate rate limit exhaustion

### Phase 2: Smart Polling (Week 2)

**Goal**: Reduce unnecessary API calls

#### 2.1 Adaptive Intervals (2 days)

**Current**: Fixed 15s interval for all repos

**Improved**: Dynamic intervals based on activity

```go
type AdaptiveInterval struct {
    currentInterval time.Duration
    minInterval     time.Duration  // 15s (active)
    maxInterval     time.Duration  // 5m (idle)
    lastActivity    time.Time
}

func (a *AdaptiveInterval) Next() time.Duration {
    timeSinceActivity := time.Since(a.lastActivity)
    
    if timeSinceActivity < 5*time.Minute {
        return a.minInterval  // 15s - recently active
    } else if timeSinceActivity < 30*time.Minute {
        return 1*time.Minute  // 60s - some activity
    } else {
        return a.maxInterval  // 5m - idle
    }
}
```

**Benefits**:
- Active repos: Still fast (15s)
- Idle repos: Much slower (5m)
- Reduces average requests by ~70%

#### 2.2 Exponential Backoff (1 day)

**When rate limited**:
```go
func (p *Provider) handleRateLimit(resetTime time.Time) {
    waitDuration := time.Until(resetTime)
    log.Printf("Rate limited. Waiting %v until reset", waitDuration)
    time.Sleep(waitDuration + 10*time.Second) // Add buffer
}
```

**Prevents**:
- Wasting API calls on 403 responses
- Compounding the rate limit problem

#### 2.3 Selective Polling (1 day)

**Config option**:
```yaml
providers:
  - type: github
    repos:
      - owner: joelklabo
        repo: ceye
        priority: high  # Poll every 15s
      - owner: joelklabo
        repo: archive-project
        priority: low   # Poll every 5m
```

**User control**: Choose which repos need real-time monitoring

### Phase 3: Efficient API Usage (Week 3)

**Goal**: Get more data per request

#### 3.1 Batch Requests (2 days)

**Current**: One API call per repo per poll
```
10 repos × 4 polls/min = 40 requests/min
```

**Improved**: Use GraphQL API (single query for multiple repos)
```graphql
query {
  repo1: repository(owner: "joelklabo", name: "ceye") {
    workflows(first: 10) { nodes { runs } }
  }
  repo2: repository(owner: "joelklabo", name: "other") {
    workflows(first: 10) { nodes { runs } }
  }
  # ... up to 10 repos in one request
}
```

**Reduction**: 40 requests/min → 4 requests/min (10x improvement!)

#### 3.2 Conditional Requests (1 day)

**Use ETags** to skip unchanged responses:
```go
// Save ETag from response
etag := resp.Header.Get("ETag")

// Next request
req.Header.Set("If-None-Match", etag)
// Gets 304 Not Modified if unchanged (doesn't count toward rate limit!)
```

#### 3.3 Cache Layer (1 day)

**In-memory cache** with smart invalidation:
- Cache workflow runs for 10-15s
- Only fetch if cache expired
- Refresh cache on user action (manual refresh)

### Estimated Impact

**Current** (10 repos, 15s interval):
```
40 requests/min × 60 = 2,400 requests/hour (48% of limit)
```

**After Optimizations**:
```
Smart intervals: 2,400 → 800 requests/hour (-67%)
GraphQL batching: 800 → 80 requests/hour (-90%)
Conditional requests: Skip 50% of unchanged responses
Final: ~40-80 requests/hour (1-2% of limit)
```

**Rate limit problem solved** without webhooks!

### Comparison: Polling Optimization vs Webhooks

| Aspect | Optimized Polling | Webhooks (localhost) |
|--------|-------------------|----------------------|
| **Complexity** | Low (enhance existing) | High (new server + tunnel) |
| **Cost** | Free | $8/month (ngrok static) |
| **Reliability** | Same as current | Depends on tunnel |
| **Setup Time** | 2-3 weeks | 3 weeks + ongoing tunnel mgmt |
| **Rate Limits** | 40-80/hour (1-2% of quota) | 0/hour (0% of quota) |
| **Latency** | 15-60 seconds | < 1 second |
| **Operational** | Zero changes | Must keep ngrok running |

### My Updated Recommendation

**For your use case (10 repos, localhost, own repos)**:

1. **Start with Phase 1**: Rate limit monitoring (1 week)
   - Understand your actual usage
   - May discover it's not actually a problem
   
2. **Then Phase 2**: Smart polling (1 week if needed)
   - Adaptive intervals solve most issues
   - Simple, reliable, free
   
3. **Then Phase 3**: Efficient API usage (1 week if still needed)
   - GraphQL batching is the "big win"
   - Gets you to < 2% of rate limit
   
4. **Webhooks**: Defer until you actually need them
   - When you deploy to server/cloud
   - When you monitor 50+ repos
   - When sub-second latency matters

**Bottom line**: Optimizing polling is 90% effective, 50% complexity, 0% cost.

---

## Decision Framework

**Choose Webhooks If**:
- [ ] Monitoring 50+ repositories
- [ ] Running on a server/cloud
- [ ] Need < 1 second latency
- [ ] Willing to manage tunnel/deployment

**Choose Optimized Polling If**:
- [x] Running locally on laptop
- [x] Monitoring < 20 repositories
- [x] 30-60 second latency is acceptable
- [x] Want simplicity and zero operational overhead

**For your situation**: Optimized polling is the right choice.
