# Cost Structure Reference

> **Source:** ROADMAP__15_COSTS.md
> **Updated:** 2026-03-08

---

## Day-0 Cost: Indexing 300 Libraries

### One-Time Indexing

| Step | Service | Details | Cost |
|------|---------|---------|------|
| GitHub API calls | GitHub | ~4,500 calls (free with PAT) | $0.00 |
| Doc parsing | Local compute | Go parser, 300 repos in <5 min | $0.00 |
| LLM condensation | Anthropic Haiku Batch | 15M input + 2.4M output @ $0.50/$2.50 per MTok | $13.50 |
| API surface generation | Anthropic Haiku Batch | 2.4M input + 450K output | $2.33 |
| Embedding | Voyage AI voyage-code-3 | 1.1M tokens @ $0.06/MTok (or $0 with free 200M) | $0.07 |
| Vectorize upload | Cloudflare | 3,600 vectors (requires Workers paid: $5/mo) | ~$0.15/mo |
| R2 upload | Cloudflare | ~7.2MB (free tier: 10GB) | $0.00 |
| **Total one-time indexing** | | | **$15.85** |

### How the $5 Free Anthropic Credits Are Spent

$5 covers ~10M Haiku Batch input tokens → enough to index ~200 repos. To index all 300, top up ~$11 (credit card required at that point).

---

## Monthly Infrastructure (Pre-Revenue)

| Service | Tier | Cost/Month |
|---------|------|------------|
| Cloudflare Vectorize + Workers | Paid plan (required for Vectorize) | $5.00 |
| Cloudflare R2 | Free (10GB storage, 10M reads) | $0.00 |
| Neon Postgres | Free (0.5GB, auto-suspend) | $0.00 |
| WorkOS (auth) | Free (1M MAU) | $0.00 |
| Voyage AI | Free tier / pay-as-go | ~$0.00 |
| GitHub repos | Free (unlimited private) | $0.00 |
| Domain (coderank.ai) | Annual | ~$1.00 |
| **Total monthly pre-revenue** | | **~$6** |

---

## Cost Per Unit

| Operation | Cost |
|-----------|------|
| Index 1 library (condensation + surface + embedding) | ~$0.05 |
| Embed 1 query (Voyage AI) | ~$0.000003 |
| Serve 1 query (embed + Vectorize search + R2 fetch) | ~$0.00003 |
| Serve 1,000 queries | ~$0.03 |
| Serve 1,000,000 queries | ~$30 |
| Reindex 1 library | ~$0.05 |
| Reindex 300 libraries (full) | ~$16 |
| Reindex 50 hot libraries (daily) | ~$2.65 |

---

## Reindexing Strategy

Not all libraries need the same frequency. Tiered approach:

| Tier | Libraries | Frequency | Trigger | Cost/Run |
|------|-----------|-----------|---------|----------|
| **Hot** (top 50) | React, Next.js, Prisma, etc. | Daily | Cron + webhook on new release | ~$2.65 |
| **Warm** (51–300) | Popular, actively maintained | Every 3 days | Cron | ~$13.25 |
| **Cold** (301+) | Stable, slow-moving | Weekly | Cron | Varies |
| **On-demand** | Any | Triggered | User requests fresh docs | ~$0.05 |

---

## Monthly Reindexing Budget

| Component | Calculation | Cost/Month |
|-----------|-----------|------------|
| Hot tier (daily × 30) | $2.65 × 30 | $79.50 |
| Warm tier (every 3 days × 10) | $13.25 × 10 | $132.50 |
| **Total reindexing (300 libs)** | | **~$212/month** |

At moderate scale, this is the largest ongoing cost. Offset by:
- Using cheaper models for stable libraries
- Caching unchanged docs (skip re-condensation if source hasn't changed)
- Batch processing during off-peak

### Optimization: Skip Unchanged Docs

Before re-condensing, compare the git hash of each doc file with the last indexed hash. If unchanged, skip. In practice, most libraries only update a few doc files per release — not all of them. This can reduce reindexing costs by 60–80%.

---

## Monthly Costs by Scale

### 100 Users (Month 1–2)

| Component | Cost/Month |
|-----------|------------|
| Cloudflare Workers + Vectorize | $5 |
| Cloudflare R2 | $0 |
| Neon Postgres | $0 |
| WorkOS | $0 |
| Embedding queries (~3K/month) | $0.01 |
| Domain | $1 |
| **Total infra** | **~$6** |
| **+ Reindexing** | **~$143** |
| **Grand total** | **~$149** |

**Revenue (5 paid × $14): $70**. Not yet profitable. Reindexing is the bottleneck — reduce frequency or fund from savings until 15+ paid users.

### 1,000 Users (Month 3–6)

| Component | Cost/Month |
|-----------|------------|
| Cloudflare Workers + Vectorize | $5 |
| Cloudflare R2 | $0 |
| Neon Postgres | $0 (free tier) |
| WorkOS | $0 (under 1M MAU) |
| Embedding queries (~30K/month) | $0.90 |
| Reindexing | $143 |
| Domain | $1 |
| **Total** | **~$150** |
| **Revenue** (50 paid × $14) | **$700** |
| **Gross margin** | **78.6%** |

### 10,000 Users (Month 8–12)

| Component | Cost/Month |
|-----------|------------|
| Cloudflare Workers + Vectorize | $5 |
| Cloudflare R2 | $0 |
| Neon Postgres | $19 (Launch tier) |
| WorkOS | $0 (under 1M MAU) |
| Stripe fees | ~$50 |
| Embedding queries (~300K/month) | $9 |
| Reindexing (expanded to 1K libs) | $200 |
| Domain | $1 |
| **Total** | **~$284** |
| **Revenue** (500 paid × $14) | **$7,000** |
| **Gross margin** | **96%** |

### 100,000 Users (Year 2)

| Component | Cost/Month |
|-----------|------------|
| Cloudflare Workers + Vectorize | $15 |
| Cloudflare R2 | $1 |
| Neon Postgres | $69 (Scale tier) |
| WorkOS | $0 (under 1M MAU) |
| Stripe fees | $500 |
| Embedding queries (~3M/month) | $90 |
| Reindexing (3K libs) | $500 |
| Domain | $1 |
| **Total** | **~$2,976** |
| **Revenue** (8,000 paid × blended $12 ARPU) | **$96,000** |
| **Gross margin** | **96.9%** |

---

## Why Margins Are So High

1. **Content is pre-computed.** The expensive LLM work (condensation) happens once during indexing, not per query. Serving is vector search + file read — pennies.
2. **Edge-native stack.** Vectorize + R2 + KV all run on Cloudflare's edge. No idle compute, no reserved instances.
3. **No GPU costs.** Condensation uses pay-per-token API (Haiku Batch), not dedicated hardware.
4. **Auto-suspend databases.** Neon Postgres suspends when idle. $0 when not queried.
5. **Free egress.** R2 has zero egress costs. Query volume doesn't increase bandwidth bills.

---

## Break-Even Analysis

| Scenario | Paid Users Needed | Monthly Revenue | Monthly Cost |
|----------|------------------|----------------|-------------|
| **Pre-reindexing** (static 300 libs) | 1 | $14 | $6 |
| **With daily hot reindexing** | 11 | $154 | $149 |
| **With full reindexing schedule** | 16 | $224 | $212 |
| **Comfortable margin** | 30 | $420 | $212 |

**Profitable at 16–30 paying users.** The entire path from $0 to profitability costs under $1,000 out of pocket.

---

## Total Spend to First Revenue

| Phase | What | Cost |
|-------|------|------|
| Phase 0 | e-Residency application | €120 (~$130) |
| Phase 1 | Build MVP + index 300 repos | ~$16 (Anthropic) |
| Phase 2 | Public beta launch | $0 |
| Phase 3 | Incorporate + domain | ~€480–715 (~$520–$775) |
| Phase 4 | Apply for startup credits | $0 |
| Phase 5 | Launch paid product | ~$0 |
| **Total to first revenue** | | **~$660–$915** |

---

## Cost Optimization Tactics

### Short-Term (Launch)

- **Skip daily reindexing initially.** Weekly full reindex ($16/run) is enough for beta. Add daily hot tier when revenue supports it.
- **Use Anthropic's $5 free credits** for the first ~200 repos. Top up $11 for the remaining 100.
- **Voyage AI 200M free tokens** covers embedding for months.
- **Cache unchanged docs.** Compare git hashes before re-condensing. Skip if source unchanged.

### Medium-Term (Growth)

- **Apply for Cloudflare credits** ($5–250K). Covers Workers, R2, Vectorize, KV for months/years.
- **Apply for Google credits** ($2K–$350K). The $10K Anthropic credits via Model Garden cover ~600 full reindex runs.
- **Batch reindexing off-peak.** Anthropic Batch API processes within 24 hours — submit overnight.

### Long-Term (Scale)

- **Use cheaper models for stable libraries.** Libraries that rarely change (lodash, jquery) don't need Haiku quality — use an even cheaper model or skip re-condensation entirely.
- **Community-contributed condensation.** Let library maintainers submit their own condensed docs (like Context7's model). Zero LLM cost for those.
- **Enterprise revenue funds everything.** A single 200-seat enterprise deal ($60K ARR) covers years of reindexing costs.

---

## Comparison: CodeRank vs Context7

| Cost Driver | Context7 (Upstash) | CodeRank |
|-------------|-------------------|----------|
| Infrastructure | Upstash Redis, managed servers | Cloudflare edge (mostly free tier) |
| Serving cost per query | Unknown (Redis + reranking) | ~$0.00003 |
| Indexing model | Community-contributed + LLM enrichment | Haiku Batch (~$0.05/library) |
| Team size | ~5–10 (Upstash employees) | 1 (solo founder) |
| Funding | Upstash revenue ($1M+ ARR) | Bootstrapped ($660–$915 total) |
| Break-even | Unknown | 16–30 paid users |

Context7 has Upstash's infrastructure and team behind it. CodeRank's advantage: Cloudflare's edge + free tiers keep costs near zero, making bootstrapping viable.
