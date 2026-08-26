# Difficulty Measurement Report

- Manifest: `mixed-generator-alignment-v6`
- Hash: `505b47c42428d0167c5f53f53ba82420674cd4885347059949b1193ee875fa9e`
- Progress: 101/101
- Complete: true
- Reproducible classifications: 101/101
- Expected matches: 15/17

## Outcomes

- solved: 84
- strategy-unsolved: 17

## Difficulties

- easy: 43
- evil: 19
- expert: 10
- hard: 12
- medium: 15
- unassigned: 2

## Source Groups

- external: 31 observed, 21 solved, 10 strategy-unsolved (32.3%, 95% CI 18.6–49.9%)
- generated: 50 observed, 47 solved, 3 strategy-unsolved (6.0%, 95% CI 2.1–16.2%)
- imported: 15 observed, 14 solved, 1 strategy-unsolved (6.7%, 95% CI 1.2–29.8%)
- pathological: 5 observed, 2 solved, 3 strategy-unsolved (60.0%, 95% CI 23.1–88.2%)

## Split Groups

- exploratory: 71 observed, 59 solved, 12 strategy-unsolved (16.9%, 95% CI 9.9–27.3%)
- held-out: 30 observed, 25 solved, 5 strategy-unsolved (16.7%, 95% CI 7.3–33.6%)

## Tier Distributions

| Tier | n | Score min/median/p95/max | Moves min/median/p95/max | Clues min/median/p95/max |
|---|---:|---|---|---|
| easy | 43 | 0/180/360/458 | 0/45/57/59 | 22/36/58/81 |
| evil | 19 | 378/1080/3402/3402 | 6/25/69/69 | 17/25/29/29 |
| expert | 10 | 248/952/2530/2530 | 4/50/78/78 | 17/24/52/52 |
| hard | 12 | 466/832/1784/1784 | 44/60/74/74 | 17/25/39/39 |
| medium | 15 | 318/480/804/804 | 49/56/67/67 | 17/25/32/32 |
| unassigned | 2 | 0/0/0/0 | 0/0/0/0 | 0/0/3/3 |

## Neighboring-Tier Score-Range Overlap

- easy / medium: overlap 318–458
- expert / evil: overlap 378–2530
- hard / expert: overlap 466–1784
- medium / hard: overlap 466–804

## Optional External Source Comparison

### norvig-file-group

- easy50: easy=7, hard=2, medium=1
- hardest: easy=2, evil=4, expert=2, hard=1, medium=2
- top95: evil=2, expert=3, hard=1, medium=4

## Generation Measurements

- easy: hits 10/10 (100.0%, 95% CI 72.2–100.0%); observed {easy=10}; latency ms 0/0/7/7; rounds 1/1/1/1
- evil: hits 4/10 (40.0%, 95% CI 16.8–68.7%); observed {easy=2, evil=5, expert=1, hard=1, medium=1}; latency ms 353/4758/20289/20289; rounds 1/1/3/3
- expert: hits 0/10 (0.0%, 95% CI 0.0–27.8%); observed {easy=3, evil=4, hard=1, medium=2}; latency ms 2236/4156/15988/15988; rounds 1/1/3/3
- hard: hits 1/10 (10.0%, 95% CI 1.8–40.4%); observed {easy=5, evil=1, hard=1, medium=3}; latency ms 32/386/2816/2816; rounds 1/1/3/3
- medium: hits 0/10 (0.0%, 95% CI 0.0–27.8%); observed {easy=10}; latency ms 6/33/2034/2034; rounds 1/1/1/1

## Method and Limits

- Every puzzle was classified twice; only exact outcome, score, maximum-technique, move-count, and trace agreement was recorded.
- Numeric summaries use deterministic nearest-rank percentiles. Rate intervals are Wilson 95% confidence intervals.
- External source scales remain separate and do not validate or redefine canonical strategy grades.
- Generated samples test generator behavior, not external source agreement. Random seeds are absent because the public generator boundary does not expose them.
- Pilot strata are deliberately small. The report identifies expansion targets and does not justify policy changes by itself.
