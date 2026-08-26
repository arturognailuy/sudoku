# Difficulty Measurement Report

- Manifest: `mixed-imported-expansion-v5`
- Hash: `3114979512be4590e77af9ab407530f590b58dbbe00c33146ce2dfb370e6b910`
- Progress: 76/76
- Complete: true
- Reproducible classifications: 76/76
- Expected matches: 15/17

## Outcomes

- solved: 62
- strategy-unsolved: 14

## Difficulties

- easy: 29
- evil: 13
- expert: 10
- hard: 10
- medium: 12
- unassigned: 2

## Source Groups

- external: 31 observed, 21 solved, 10 strategy-unsolved (32.3%, 95% CI 18.6–49.9%)
- generated: 25 observed, 25 solved, 0 strategy-unsolved (0.0%, 95% CI 0.0–13.3%)
- imported: 15 observed, 14 solved, 1 strategy-unsolved (6.7%, 95% CI 1.2–29.8%)
- pathological: 5 observed, 2 solved, 3 strategy-unsolved (60.0%, 95% CI 23.1–88.2%)

## Split Groups

- exploratory: 51 observed, 42 solved, 9 strategy-unsolved (17.6%, 95% CI 9.6–30.3%)
- held-out: 25 observed, 20 solved, 5 strategy-unsolved (20.0%, 95% CI 8.9–39.1%)

## Tier Distributions

| Tier | n | Score min/median/p95/max | Moves min/median/p95/max | Clues min/median/p95/max |
|---|---:|---|---|---|
| easy | 29 | 0/196/360/412 | 0/48/57/59 | 22/32/80/81 |
| evil | 13 | 378/914/2518/2518 | 6/25/67/67 | 17/25/29/29 |
| expert | 10 | 248/952/2530/2530 | 4/50/78/78 | 17/24/52/52 |
| hard | 10 | 466/764/1784/1784 | 44/59/74/74 | 17/25/39/39 |
| medium | 12 | 318/482/804/804 | 49/57/67/67 | 17/24/32/32 |
| unassigned | 2 | 0/0/0/0 | 0/0/0/0 | 0/0/3/3 |

## Neighboring-Tier Score-Range Overlap

- easy / medium: overlap 318–412
- expert / evil: overlap 378–2518
- hard / expert: overlap 466–1784
- medium / hard: overlap 466–804

## External Rating Agreement

### norvig-file-group

- easy50: easy=7, hard=2, medium=1
- hardest: easy=2, evil=4, expert=2, hard=1, medium=2
- top95: evil=2, expert=3, hard=1, medium=4

## Generation Measurements

- easy: hits 5/5 (100.0%, 95% CI 56.6–100.0%); observed {easy=5}; latency ms 0/1/1/1; rounds 1/1/1/1
- evil: hits 2/5 (40.0%, 95% CI 11.8–76.9%); observed {easy=1, evil=2, expert=1, hard=1}; latency ms 2107/5925/20289/20289; rounds 1/2/3/3
- expert: hits 0/5 (0.0%, 95% CI 0.0–43.4%); observed {easy=1, evil=2, medium=2}; latency ms 4156/7632/15988/15988; rounds 1/1/3/3
- hard: hits 0/5 (0.0%, 95% CI 0.0–43.4%); observed {easy=4, medium=1}; latency ms 32/868/1964/1964; rounds 1/1/1/1
- medium: hits 0/5 (0.0%, 95% CI 0.0–43.4%); observed {easy=5}; latency ms 6/22/115/115; rounds 1/1/1/1

## Method and Limits

- Every puzzle was classified twice; only exact outcome, score, maximum-technique, move-count, and trace agreement was recorded.
- Numeric summaries use deterministic nearest-rank percentiles. Rate intervals are Wilson 95% confidence intervals.
- External source scales remain separate and are not treated as equivalent human-difficulty labels.
- Generated samples test generator behavior, not external validity. Random seeds are absent because the public generator boundary does not expose them.
- Pilot strata are deliberately small. The report identifies expansion targets and does not justify policy changes by itself.
