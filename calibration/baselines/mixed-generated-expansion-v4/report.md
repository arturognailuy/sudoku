# Difficulty Measurement Report

- Manifest: `mixed-generated-expansion-v4`
- Hash: `d4642c0a92271bee7e1614f083392bbfdef0b18c1a5f1fe4a1b046f4cb967b70`
- Progress: 67/67
- Complete: true
- Reproducible classifications: 67/67
- Expected matches: 8/8

## Outcomes

- solved: 54
- strategy-unsolved: 13

## Difficulties

- easy: 29
- evil: 11
- expert: 7
- hard: 6
- medium: 12
- unassigned: 2

## Source Groups

- external: 31 observed, 21 solved, 10 strategy-unsolved (32.3%, 95% CI 18.6–49.9%)
- generated: 25 observed, 25 solved, 0 strategy-unsolved (0.0%, 95% CI 0.0–13.3%)
- imported: 6 observed, 6 solved, 0 strategy-unsolved (0.0%, 95% CI 0.0–39.0%)
- pathological: 5 observed, 2 solved, 3 strategy-unsolved (60.0%, 95% CI 23.1–88.2%)

## Split Groups

- exploratory: 45 observed, 36 solved, 9 strategy-unsolved (20.0%, 95% CI 10.9–33.8%)
- held-out: 22 observed, 18 solved, 4 strategy-unsolved (18.2%, 95% CI 7.3–38.5%)

## Tier Distributions

| Tier | n | Score min/median/p95/max | Moves min/median/p95/max | Clues min/median/p95/max |
|---|---:|---|---|---|
| easy | 29 | 0/196/360/412 | 0/48/57/59 | 22/32/80/81 |
| evil | 11 | 378/1012/2518/2518 | 6/25/67/67 | 17/25/29/29 |
| expert | 7 | 248/1128/2530/2530 | 4/59/78/78 | 17/23/26/26 |
| hard | 6 | 466/764/1784/1784 | 55/60/74/74 | 17/25/28/28 |
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
