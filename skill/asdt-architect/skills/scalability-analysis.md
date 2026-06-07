# Scalability Analysis Guidelines

## Purpose

Identify bottlenecks, caching opportunities, and scaling options. Applied during Step 5 of the Architect workflow.

## Bottleneck Identification

### N+1 Query Problem

Occurs when fetching a list of N items triggers N additional queries (one per item). Detection pattern:
- Any loop that calls a data layer function: suspect N+1.
- ORM `for item in collection: item.related_data` patterns.

Mitigation: eager loading, batch loading (DataLoader pattern), or joining at the query level.

### Missing Indexes

For every query in the system design:
1. Identify the WHERE / JOIN / ORDER BY columns.
2. Verify that a composite index covers the selectivity order (most selective column first).
3. Flag missing indexes as Medium or High risk depending on expected table size.

Rule of thumb: any table expected to exceed 10k rows needs index analysis before shipping.

### Synchronous Chain Bottlenecks

A synchronous call chain `A → B → C → D` means latency stacks. Identify:
- Total expected p95 latency of the chain.
- Which calls can be made async or parallelized.
- Which calls are on the critical path (cannot be deferred).

Flag chains longer than 3 hops with p95 latency > 500ms as High risk.

## Horizontal vs Vertical Scaling Options

| Strategy | What it is | When to use |
|----------|------------|-------------|
| **Vertical** | Increase CPU/RAM of existing instance | Quick win; limited ceiling; no code change |
| **Horizontal** | Add more instances behind a load balancer | Requires stateless design; scales indefinitely |
| **Read replicas** | Separate read/write DB endpoints | Read-heavy workloads; eventual consistency acceptable |
| **Sharding** | Partition data across multiple DB instances | Very large datasets; complex operationally |
| **Event-driven** | Decouple producers/consumers via queue | Bursty workloads; tolerate eventual consistency |

Document the chosen strategy in `system-design.yaml` under `service_boundaries`.

## Caching Strategies

| Layer | Use case | TTL guidance |
|-------|----------|-------------|
| **In-process / memory** | Frequently read, rarely changed config or lookup data | Session-lifetime or explicit invalidation |
| **Distributed cache (Redis)** | Shared state across instances; user session data; rate-limit counters | Seconds to hours depending on staleness tolerance |
| **HTTP cache (ETag / Cache-Control)** | Public or semi-public read endpoints | Minutes to days |
| **CDN** | Static assets, public content | Hours to months |

Cache invalidation strategy must be documented before introducing any cache. Options:
- **TTL-based**: simple; accepts eventual consistency window.
- **Write-through**: cache updated on every write; strong consistency; write latency cost.
- **Cache-aside**: load on miss; explicit invalidation on write; most common for application caches.

## Async Patterns

When a synchronous operation is too slow for the critical path:
1. **Fire-and-forget**: send to queue; return 202 Accepted; client polls or subscribes for result.
2. **Background job**: schedule for deferred execution; return job ID to caller.
3. **Event sourcing**: emit domain event; downstream consumers react independently.

Async introduces eventual consistency. Document which guarantees are acceptable for the feature.

## Load Estimation

For any new endpoint or background job, estimate:
- Expected requests per second at p50 and p99 traffic.
- Data volume per operation (read bytes, write bytes).
- Growth curve: 30-day and 1-year projections.

If estimates are unavailable, flag as `open_items[]` and default to designing for 10x current load.

## Scalability Checklist

Before finalizing `system-design.yaml`:
- [ ] N+1 patterns identified and mitigated
- [ ] Index strategy documented for all queries
- [ ] Synchronous chain latency estimated
- [ ] Scaling strategy stated (vertical / horizontal / event-driven)
- [ ] Cache strategy defined if applicable
- [ ] Async pattern chosen if any operations exceed 200ms on critical path
- [ ] Load estimates present or flagged as open_items
