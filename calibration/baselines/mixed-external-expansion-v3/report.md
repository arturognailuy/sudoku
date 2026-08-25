# Difficulty Measurement Report

- Manifest: `mixed-external-expansion-v3`
- Hash: `5d26ab6413e56bb3f59ce8637b82916c719217e88cd6bce9ac70968f99ad59b0`
- Progress: 52/52
- Complete: true
- Reproducible classifications: 52/52
- Expected matches: 8/8

## Outcomes

- solved: 39
- strategy-unsolved: 13

## Difficulties

- easy: 20
- evil: 7
- expert: 7
- hard: 5
- medium: 11
- unassigned: 2

## Source Groups

- external: 31 observed, 21 solved, 10 strategy-unsolved (32.3%, 95% CI 18.6–49.9%)
- generated: 10 observed, 10 solved, 0 strategy-unsolved (0.0%, 95% CI 0.0–27.8%)
- imported: 6 observed, 6 solved, 0 strategy-unsolved (0.0%, 95% CI 0.0–39.0%)
- pathological: 5 observed, 2 solved, 3 strategy-unsolved (60.0%, 95% CI 23.1–88.2%)

## Split Groups

- exploratory: 35 observed, 26 solved, 9 strategy-unsolved (25.7%, 95% CI 14.2–42.1%)
- held-out: 17 observed, 13 solved, 4 strategy-unsolved (23.5%, 95% CI 9.6–47.3%)

## Tier Distributions

| Tier | n | Score min/median/p95/max | Moves min/median/p95/max | Clues min/median/p95/max |
|---|---:|---|---|---|
| easy | 20 | 0/196/360/412 | 0/49/57/59 | 22/30/80/81 |
| evil | 7 | 378/914/1184/1184 | 6/13/54/54 | 17/26/29/29 |
| expert | 7 | 248/1128/2530/2530 | 4/59/78/78 | 17/23/26/26 |
| hard | 5 | 466/936/1784/1784 | 55/60/74/74 | 17/25/28/28 |
| medium | 11 | 318/500/804/804 | 49/57/67/67 | 17/24/32/32 |
| unassigned | 2 | 0/0/0/0 | 0/0/0/0 | 0/0/3/3 |

## Neighboring-Tier Score-Range Overlap

- easy / medium: overlap 318–412
- expert / evil: overlap 378–1184
- hard / expert: overlap 466–1784
- medium / hard: overlap 466–804

## External Rating Agreement

### norvig-file-group

- easy50: easy=7, hard=2, medium=1
- hardest: easy=2, evil=4, expert=2, hard=1, medium=2
- top95: evil=2, expert=3, hard=1, medium=4

## Generation Measurements

- easy: hits 2/2 (100.0%, 95% CI 34.2–100.0%); observed {easy=2}; latency ms 1/1/1/1; rounds 1/1/1/1
- evil: hits 0/2 (0.0%, 95% CI 0.0–65.8%); observed {easy=1, expert=1}; latency ms 2107/2107/20289/20289; rounds 1/1/3/3
- expert: hits 0/2 (0.0%, 95% CI 0.0–65.8%); observed {easy=1, medium=1}; latency ms 4156/4156/5250/5250; rounds 1/1/1/1
- hard: hits 0/2 (0.0%, 95% CI 0.0–65.8%); observed {easy=1, medium=1}; latency ms 868/868/1584/1584; rounds 1/1/1/1
- medium: hits 0/2 (0.0%, 95% CI 0.0–65.8%); observed {easy=2}; latency ms 9/9/22/22; rounds 1/1/1/1

## Method and Limits

- Every puzzle was classified twice; only exact outcome, score, maximum-technique, move-count, and trace agreement was recorded.
- Numeric summaries use deterministic nearest-rank percentiles. Rate intervals are Wilson 95% confidence intervals.
- External source scales remain separate and are not treated as equivalent human-difficulty labels.
- Generated samples test generator behavior, not external validity. Random seeds are absent because the public generator boundary does not expose them.
- Pilot strata are deliberately small. The report identifies expansion targets and does not justify policy changes by itself.
