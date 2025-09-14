# Kader-Planung Migration Guide

This guide helps you migrate from the previous version of kader-planung to the current version with integrated somatogram percentiles and simplified CLI interface.

## Overview of Changes

### ✅ **New Features**
- **Somatogram Percentile Column**: Added `somatogram_percentile` column with Germany-wide percentile rankings
- **Simplified CLI**: Removed obsolete parameters for cleaner user experience
- **Always Germany-wide Data**: Complete German player dataset (~50,000 players) processed for accurate percentiles
- **Improved Performance**: Optimized data collection and processing pipeline

### ⚠️ **Breaking Changes**
- **Removed Parameters**: Several obsolete CLI parameters have been completely removed
- **CSV Format**: Output now always includes `somatogram_percentile` column
- **Processing Behavior**: Always processes complete German dataset regardless of filters

## Changed Default Values

The following parameter has a new default value:

### `--min-sample-size`
- **Before**: Default was `100`
- **After**: Default is now `10`
- **Reason**: Lower threshold provides percentile data for more age groups while maintaining reasonable statistical reliability
- **Migration**: If you specifically need the old behavior, explicitly set `--min-sample-size=100` in your commands

## Removed Parameters

The following CLI parameters have been **completely removed**:

### `--include-statistics`
- **Before**: `--include-statistics=true/false`
- **After**: Statistics are now **always included** by default
- **Migration**: Simply remove this parameter from your commands

### `--mode`
- **Before**: `--mode=detailed|efficient|statistical|hybrid`
- **After**: Application now **always uses hybrid mode** for optimal performance
- **Migration**: Remove this parameter from your commands

### `--output-format`
- **Before**: `--output-format=json|excel|csv-statistical|csv`
- **After**: Output is **always CSV format** optimized for German Excel compatibility
- **Migration**: Remove this parameter - all output will be CSV

## Command Migration Examples

### Before (Old Commands)
```bash
# Old way - with removed parameters
.\bin\kader-planung.exe --mode=hybrid --include-statistics --output-format=csv --club-prefix="C0"

# Old way - different modes
.\bin\kader-planung.exe --mode=detailed --include-statistics=true --output-format=csv

# Old way - statistical mode
.\bin\kader-planung.exe --mode=statistical --output-format=csv-statistical
```

### After (New Commands)
```bash
# New way - much simpler!
.\bin\kader-planung.exe --club-prefix="C0"

# All previous functionality is included by default
.\bin\kader-planung.exe

# With additional options
.\bin\kader-planung.exe --club-prefix="C0" --concurrency=10 --verbose
```

## Configuration File Migration

### Environment Configuration (.env)

**Before**:
```env
# These parameters are no longer supported or have changed defaults
KADER_PLANUNG_MODE=hybrid
KADER_PLANUNG_INCLUDE_STATISTICS=true
KADER_PLANUNG_OUTPUT_FORMAT=csv
KADER_PLANUNG_MIN_SAMPLE_SIZE=100
```

**After**:
```env
# Remove the obsolete parameters entirely
# Keep only the supported parameters:
KADER_PLANUNG_ENABLED=true
KADER_PLANUNG_BINARY_PATH=kader-planung/bin/kader-planung.exe
KADER_PLANUNG_OUTPUT_DIR=internal/static/demo/kader-planung
KADER_PLANUNG_API_BASE_URL=http://localhost:8080
KADER_PLANUNG_CLUB_PREFIX=
KADER_PLANUNG_TIMEOUT=30
KADER_PLANUNG_CONCURRENCY=16
KADER_PLANUNG_MIN_SAMPLE_SIZE=10  # Changed from 100 to 10 for better data coverage
KADER_PLANUNG_VERBOSE=false
```

## New CSV Output Format

The CSV output now includes an additional column `somatogram_percentile`:

### New Column Position
The new column is positioned after `success_rate_last_12_months`:

```csv
...,games_last_12_months,success_rate_last_12_months,somatogram_percentile,dwz_age_relation
...,25,36.0,67.2,150
...,18,DATA_NOT_AVAILABLE,89.0,175
```

### Column Values
- **Valid percentiles**: `0.0` to `100.0` (with one decimal place)
- **Missing data**: `DATA_NOT_AVAILABLE` when insufficient sample size

## Behavioral Changes

### 1. Data Collection Scope

**Before**: Data collection could be limited by parameters and filters

**After**: **Always collects ALL German players** (~50,000 players from ~4,300 clubs) for accurate Germany-wide percentiles

```bash
# Both commands process the SAME complete German dataset
.\bin\kader-planung.exe --club-prefix="C0"    # Filters output only
.\bin\kader-planung.exe                       # All clubs in output
```

### 2. Filter Behavior

**Before**: Filters could affect data collection and percentile accuracy

**After**: `--club-prefix` **only affects final CSV output**, not data collection or percentile calculation

### 3. Processing Time

**Before**: Processing time varied based on selected clubs

**After**: Processing time is more consistent (~2-5 minutes additional) due to complete German dataset collection

## Migration Checklist

### ✅ **Step 1: Update CLI Commands**
- [ ] Remove `--include-statistics` from all commands
- [ ] Remove `--mode` from all commands
- [ ] Remove `--output-format` from all commands
- [ ] Test simplified commands work as expected

### ✅ **Step 2: Update Configuration Files**
- [ ] Remove obsolete environment variables from `.env` files
- [ ] Update any automation scripts or CI/CD pipelines
- [ ] Update Docker configurations if applicable

### ✅ **Step 3: Update Data Processing**
- [ ] Update any CSV parsing code to handle new `somatogram_percentile` column
- [ ] Verify column positions if using positional CSV parsing
- [ ] Test with new percentile data format (floating-point values)

### ✅ **Step 4: Performance Considerations**
- [ ] Adjust expectations for processing time (+30-60 seconds)
- [ ] Ensure sufficient memory available (~200-300MB additional)
- [ ] Update any timeout configurations for longer processing

### ✅ **Step 5: Validation**
- [ ] Run test commands to verify output format
- [ ] Validate percentile values are within expected ranges (0.0-100.0)
- [ ] Confirm all existing functionality still works

## Troubleshooting

### Migration Issues

**1. Command Not Found / Unknown Flag**
```bash
Error: unknown flag: --mode
```
**Solution**: Remove the obsolete parameter from your command

**2. Unexpected Processing Time**
```
Processing is taking much longer than before
```
**Solution**: This is expected - the application now processes all German players for accurate percentiles

**3. CSV Parse Errors**
```
Error: unexpected column count in CSV
```
**Solution**: Update your CSV parsing to handle the new `somatogram_percentile` column

**4. Percentile Values Format**
```
Expected integer percentiles, got floating-point
```
**Solution**: Update your code to handle floating-point percentile values (e.g., `67.2`, `89.0`)

### Getting Help

If you encounter issues during migration:

1. **Check logs**: Use `--verbose` flag for detailed logging
2. **Verify parameters**: Run `.\bin\kader-planung.exe --help` to see current parameters
3. **Test with small datasets**: Use specific `--club-prefix` for testing
4. **Validate output**: Check that CSV includes the new `somatogram_percentile` column

### Rollback Plan

If you need to temporarily rollback:

1. **Identify previous version**: Check git history for last working version
2. **Rebuild old version**: Use git checkout and rebuild if necessary
3. **Document issues**: Note specific problems for troubleshooting
4. **Plan migration**: Work through migration steps systematically

## Benefits of Migration

After successful migration, you'll have:

- ✅ **Simpler CLI**: Fewer parameters to remember and manage
- ✅ **Richer Data**: Germany-wide percentile rankings for all players
- ✅ **Better Accuracy**: Percentiles calculated from complete German dataset
- ✅ **Consistent Performance**: Predictable processing behavior
- ✅ **Future-Proof**: Modern codebase ready for additional enhancements

## Support

For migration assistance:
- Review this guide thoroughly before starting
- Test commands in a development environment first
- Use `--verbose` logging to diagnose issues
- Document any problems encountered for future reference