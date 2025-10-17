# Stress Test Milestone Schedule

**Test Started**: 06:40:40  
**Expected End**: 07:01:40 (21 minutes)

## Milestone Checkpoints

| Milestone | Clock Time | Elapsed | VUs | Significance |
|-----------|------------|---------|-----|--------------|
| ✅ Stage 1 | 06:44:40 | 4m | 50 | Baseline established |
| 🎯 Stage 2 | 06:47:40 | 7m | 100 | Load test level |
| 🎯 Stage 3 | **06:50:40** | **10m** | **200** | **Stress begins** ⚠️ |
| 🎯 Stage 4 | 06:53:40 | 13m | 300 | High stress |
| 🎯 Stage 5 | **06:56:40** | **16m** | **400** | **Very high stress** ⚠️ |
| 🎯 Stage 6 | 06:59:40 | 19m | 500 | Maximum stress |
| 🏁 Complete | **~07:02** | **~21m** | **0** | **Final results** ⚠️ |

## Check Commands

```bash
# Milestone 1: Stage 3 @ 200 VUs (06:50:40)
./tests/load/milestone_check.sh "Stage 3 - 200 VUs" 200

# Milestone 2: Stage 5 @ 400 VUs (06:56:40)  
./tests/load/milestone_check.sh "Stage 5 - 400 VUs" 400

# Milestone 3: Completion (07:02)
ps -p 132195 || echo "✅ Test Complete - Ready for analysis"
./tests/load/analyze_stress.sh
```

## What to Look For

### Stage 3 (200 VUs) - Expected ~06:50
- **Expect**: p95 < 10ms (still good performance)
- **Watch**: First signs of stress
- **Alert**: If errors appear

### Stage 5 (400 VUs) - Expected ~06:56
- **Expect**: p95 10-100ms (degradation visible)
- **Watch**: Error rate climbing
- **Alert**: If p95 > 1000ms or errors > 5%

### Completion (~07:02)
- **Analyze**: Full performance curve
- **Identify**: Breaking point
- **Document**: Capacity limits

---

**Status**: Waiting for milestones (current: Stage 2 @ 100 VUs)
