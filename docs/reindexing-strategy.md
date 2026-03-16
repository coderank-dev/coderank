# Reindexing Strategy & Budget Automation

> **Source:** ROADMAP__15_COSTS.md
> **Updated:** 2026-03-08

---

## Overview

Not all libraries need the same reindexing frequency. A tiered approach optimizes costs while maintaining freshness:

| Tier | Libraries | Frequency | Trigger | Cost/Run |
|------|-----------|-----------|---------|----------|
| **Hot** (top 50) | React, Next.js, Prisma, etc. | Daily | Cron + webhook on new release | ~$2.65 |
| **Warm** (51–300) | Popular, actively maintained | Every 3 days | Cron | ~$13.25 |
| **Cold** (301+) | Stable, slow-moving | Weekly | Cron | Varies |
| **On-demand** | Any | Triggered | User requests fresh docs | ~$0.05 |

---

## Release Detection Implementation

### Detection via GitHub API

```go
// Check for new releases via GitHub API
// GET /repos/{owner}/{repo}/releases/latest
// Compare tag_name with last indexed version
// If different → trigger reindex

// For repos without formal releases:
// GET /repos/{owner}/{repo}/commits?since={last_indexed_at}&per_page=1
// If commits in /docs → trigger reindex
```

### Pseudocode: Release Detection Function

```go
package indexing

import (
	"context"
	"time"
	"github.com/google/go-github/v57/github"
)

type ReleaseDetector struct {
	client *github.Client
}

// CheckForNewRelease determines if a library has been updated since last index
func (rd *ReleaseDetector) CheckForNewRelease(
	ctx context.Context,
	owner, repo string,
	lastIndexedAt time.Time,
) (hasNewRelease bool, latestTag string, err error) {

	// Strategy 1: Check latest formal release
	release, _, err := rd.client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err == nil && release != nil {
		if release.TagName != "" {
			// TODO: Store and compare with db.lastIndexedVersion
			latestTag = *release.TagName
			// If different from stored version → hasNewRelease = true
			return true, latestTag, nil
		}
	}

	// Strategy 2: Check recent commits to doc files
	opts := &github.CommitsListOptions{
		Since: lastIndexedAt,
		ListOptions: github.ListOptions{PerPage: 1},
	}
	commits, _, err := rd.client.Repositories.ListCommits(ctx, owner, repo, opts)
	if err == nil && len(commits) > 0 {
		for _, commit := range commits {
			for _, file := range commit.Files {
				if strings.HasPrefix(*file.Filename, "docs/") {
					return true, commit.SHA, nil
				}
			}
		}
	}

	// No updates detected
	return false, "", nil
}
```

> **Implementation:** See pipeline-ops UOW_41 (Freshness Enforcement) and UOW_52 (Pipeline Event Webhooks) for the production implementation of this logic.

---

## Monthly Reindexing Budget

### Full Schedule (300 Libraries)

| Component | Calculation | Cost/Month |
|-----------|-------------|------------|
| Hot tier (daily × 30) | $2.65 × 30 | $79.50 |
| Warm tier (every 3 days × 10) | $13.25 × 10 | $132.50 |
| **Total reindexing (300 libs)** | | **~$212/month** |

At moderate scale, this is the largest ongoing cost. Offset by:
- Using cheaper models for stable libraries
- Caching unchanged docs (skip re-condensation if source hasn't changed)
- Batch processing during off-peak

### Cost Breakdown by Tier

#### Hot Tier (Daily Reindexing)
- **Libraries:** Top 50 (React, Next.js, Prisma, etc.)
- **Frequency:** Daily (30 runs/month)
- **Cost per run:** ~$2.65
- **Monthly cost:** $79.50
- **Rationale:** High-velocity repos release frequently; users expect current docs

#### Warm Tier (Every 3 Days)
- **Libraries:** 51–300 (popular, actively maintained)
- **Frequency:** Every 3 days (~10 runs/month)
- **Cost per run:** ~$13.25
- **Monthly cost:** $132.50
- **Rationale:** Moderate update frequency; balanced freshness vs. cost

#### Cold Tier (Weekly)
- **Libraries:** 301+ (stable, slow-moving)
- **Frequency:** Weekly (~4 runs/month)
- **Cost per run:** Variable (partial coverage)
- **Monthly cost:** ~$0 (grouped with warm or skipped)
- **Rationale:** Minimal changes; weekly updates sufficient

#### On-Demand Reindexing
- **Cost per request:** ~$0.05
- **Trigger:** User requests fresh docs for specific library
- **Budget:** Not included in base calculation (add as needed)

---

## Optimization: Skip Unchanged Docs

Before re-condensing documentation, compare content hashes to skip unchanged files.

### Algorithm

```
FOR EACH library in reindex_batch:
  FETCH last_indexed_docs_hash FROM database

  FOR EACH doc file in library:
    COMPUTE current_hash = SHA256(file_content)
    IF current_hash == last_indexed_docs_hash:
      SKIP re-condensation for this file
    ELSE:
      TRIGGER re-condensation + embedding

  UPDATE database WITH new_hash
```

### Impact

- **Cost reduction:** 60–80% (most libraries update only a subset of docs per release)
- **Example:** React releases ~3 docs per month out of 200 total → 98.5% skip rate
- **Monthly savings at 300-library scale:** ~$127–170

### Implementation Notes

- Store git commit SHAs or file hashes in database for each indexed doc
- Before condensation, compare current source hash with stored hash
- Skip API calls and LLM tokens for unchanged content
- Fall back to full reindex if hash storage fails

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

**Revenue (5 paid × $14): $70.** Reindexing is the bottleneck — reduce frequency or fund from savings until 15+ paid users.

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

## Reindexing Optimization Tactics

### At Launch
- **Skip daily hot tier initially.** Weekly full reindex ($16/run) covers beta phase.
- **Enable on-demand reindexing only.** Users can request fresh docs manually.
- **Cost:** ~$70/month (4 scheduled runs)

### During Growth (Post-Revenue)
- **Activate hot tier (daily) at 11+ paying users.** Revenue covers $79.50/month cost.
- **Add warm tier (every 3 days) at 20+ paying users.** Revenue covers full $212/month budget.
- **Implement hash-based skipping.** Reduce effective cost by 60–80%.

### At Scale
- **Use cheaper models for cold tier.** Libraries like lodash, jquery don't need Haiku quality.
- **Community-contributed condensation.** Accept pre-written docs from library maintainers (zero LLM cost).
- **Batch during off-peak.** Submit reindexing jobs to Anthropic Batch API at night (24-hour turnaround).
- **Enterprise revenue funds reindexing.** One $60K/year deal covers years of updates.
