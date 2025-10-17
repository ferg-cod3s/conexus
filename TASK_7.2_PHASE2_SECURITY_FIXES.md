# Task 7.2 Phase 2 - Input Validation Security Fixes

## Status: Validation Package Complete ✅

**Validation package location:** `internal/validation/input.go` (203 lines)  
**Test coverage:** `internal/validation/input_test.go` (378 lines) - ALL TESTS PASSING ✅

### Security Features Implemented:
1. **PathValidator** - Full filesystem validation with `os.Root` (Go 1.24)
2. **ValidatePath()** - Complete validation with stat checks
3. **SanitizePath()** - Basic sanitization without I/O
4. **IsPathSafe()** - Lightweight pattern validation
5. **ValidateConfigPath()** - Absolute path requirement for config files
6. **MustValidatePath()** - Panic-on-invalid variant

## Remaining Work: Apply Validation to 9 Vulnerable Files

### Priority 0 (Critical Path Traversal) - 7 Locations

#### 1. internal/indexer/indexer_impl.go (Line 70 - relPath usage)
**Issue:** `relPath` from `filepath.Rel()` used without validation before chunking  
**Fix:** Add `validation.IsPathSafe(relPath)` check before processing
```go
// After line 62 (after getting relPath):
if !validation.IsPathSafe(relPath) {
    return fmt.Errorf("unsafe path detected: %s", relPath)
}
```

#### 2. internal/indexer/indexer_impl.go (Lines 144, 155, 157 - fullPath construction)
**Issue:** `filepath.Join(opts.RootPath, relPath)` without validating `relPath` first  
**Fix:** Add validation before `os.Stat(fullPath)`
```go
// After line 137 (in loop over changedPaths):
if !validation.IsPathSafe(relPath) {
    return nil, nil, fmt.Errorf("unsafe path in changes: %s", relPath)
}
```

#### 3. internal/indexer/merkle.go (Line 103 - return data)
**Issue:** File content returned without path validation during tree building  
**Context:** This is in `Hash()` method - paths come from Walker  
**Fix:** Walker already validates (see #6), but add defensive check:
```go
// In Walk callback, validate relPath before processing
```

#### 4. internal/indexer/walker.go (Line 138 - match function)
**Issue:** `relPath` parameter in pattern matching not validated  
**Fix:** Add validation in Walk() before calling callback
```go
// In Walk() method, before calling WalkFunc:
if !validation.IsPathSafe(relPath) {
    continue // Skip unsafe paths
}
```

#### 5. internal/config/config.go (Line 87 - config file loading)
**Issue:** `configFile` from env var loaded without validation  
**Fix:** Use `ValidateConfigPath()` before loading
```go
// Replace line 102:
if err := validation.ValidateConfigPath(configFile); err != nil {
    return nil, fmt.Errorf("invalid config file path: %w", err)
}
```

#### 6. internal/orchestrator/state/persistence.go (Lines 93, 172 - ReadDir/Remove)
**Issue:** Directory entries read/removed without validation  
**Fix:** Validate paths before operations
```go
// In ListSessions() after line 98:
sessionPath := filepath.Join(p.baseDir, name)
if !validation.IsPathSafe(name) {
    continue
}

// In ClearAll() after line 170:
if !validation.IsPathSafe(entry.Name()) {
    continue
}
```

#### 7. internal/orchestrator/state/persistence.go (Lines 95, 116, 174 - filepath.Join)
**Issue:** Paths joined without validation  
**Fix:** Already covered by #6 above

### Priority 1 (Command Injection) - 1 Location

#### 8. internal/process/manager.go (Line 59 - exec.CommandContext)
**Issue:** Process name from `req.ProcessName` passed to exec without validation  
**Fix:** Validate against whitelist or reject special characters
```go
// After line 58, before exec.CommandContext:
if strings.ContainsAny(req.ProcessName, ";|&$`<>") {
    return fmt.Errorf("invalid characters in process name")
}
// OR use whitelist:
validProcesses := map[string]bool{"go": true, "git": true}
if !validProcesses[req.ProcessName] {
    return fmt.Errorf("process not in whitelist: %s", req.ProcessName)
}
```

## Implementation Strategy

### Phase A: Add import to each file
```go
import "github.com/ferg-cod3s/conexus/internal/validation"
```

### Phase B: Add validation checks (see specific fixes above)

### Phase C: Update tests to expect new validation errors

### Phase D: Run full test suite
```bash
go test ./internal/indexer ./internal/config ./internal/orchestrator/state ./internal/process
```

## Success Criteria
- ✅ All validation tests pass (DONE)
- ⬜ All 9 vulnerable locations fixed
- ⬜ All existing tests still pass (or updated appropriately)
- ⬜ No new path traversal vulnerabilities introduced
- ⬜ Command injection in process manager blocked

