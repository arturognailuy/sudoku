# Difficulty Measurement Report

- Manifest: `mixed-baseline-pilot-v2`
- Hash: `1cc1f09fea0d3696b6377a081217e8651b018c647a570538377a72db514a2533`
- Progress: 30/30
- Complete: true
- Reproducible classifications: 30/30
- Expected matches: 8/8

## Outcomes

- solved: 25
- strategy-unsolved: 5

## Difficulties

- easy: 14
- evil: 2
- expert: 3
- hard: 1
- medium: 8
- unassigned: 2

## Source Groups

- external: 9 observed, 7 solved, 2 strategy-unsolved (22.2%, 95% CI 6.3–54.7%)
- generated: 10 observed, 10 solved, 0 strategy-unsolved (0.0%, 95% CI 0.0–27.8%)
- imported: 6 observed, 6 solved, 0 strategy-unsolved (0.0%, 95% CI 0.0–39.0%)
- pathological: 5 observed, 2 solved, 3 strategy-unsolved (60.0%, 95% CI 23.1–88.2%)

## Split Groups

- exploratory: 19 observed, 16 solved, 3 strategy-unsolved (15.8%, 95% CI 5.5–37.6%)
- held-out: 11 observed, 9 solved, 2 strategy-unsolved (18.2%, 95% CI 5.1–47.7%)

## Tier Distributions

| Tier | n | Score min/median/p95/max | Moves min/median/p95/max | Clues min/median/p95/max |
|---|---:|---|---|---|
| easy | 14 | 0/196/412/412 | 0/49/57/57 | 24/30/81/81 |
| evil | 2 | 748/748/830/830 | 11/11/54/54 | 26/26/29/29 |
| expert | 3 | 248/774/892/892 | 4/59/59/59 | 23/24/26/26 |
| hard | 1 | 936/936/936/936 | 60/60/60/60 | 25/25/25/25 |
| medium | 8 | 342/482/804/804 | 49/56/67/67 | 17/22/32/32 |
| unassigned | 2 | 0/0/0/0 | 0/0/0/0 | 0/0/3/3 |

## Neighboring-Tier Score-Range Overlap

- easy / medium: overlap 342–412
- expert / evil: overlap 748–830
- hard / expert: no observed range overlap
- medium / hard: no observed range overlap

## External Rating Agreement

### norvig-file-group

- easy50: easy=3
- hardest: evil=1, expert=1, medium=1
- top95: medium=3

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
