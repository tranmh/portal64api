# 🔍 IMPORT PROCESS DEBUGGING GUIDE

## 📊 **CURRENT SITUATION SUMMARY**

| **Metric** | **Production** | **Local** | **Difference** |
|------------|---------------|-----------|----------------|
| Total person records | 521,365 | 492,846 | **-28,519 (-5.5%)** |
| "Lo*" surnames | 2,862 | 2,690 | -172 |
| "Leona*" first names | 745 | 630 | -115 |
| Target records | 5 | 4 | **-1 missing** |

**✅ Character encoding is CORRECT** - No corruption detected  
**❌ Data import process has 5.5% data loss** - Root cause unknown

## 🛠️ **DEBUGGING TOOLS CREATED**

### 1. **investigate_import_process.sh** - MAIN INVESTIGATION SCRIPT
**Purpose:** Complete 6-phase investigation of import process
```bash
./investigate_import_process.sh
```
**Output:** Comprehensive investigation report in `logs/investigation_YYYYMMDD_HHMMSS/`

### 2. **debug_import_process.sh** - DATABASE STATE ANALYSIS
**Purpose:** Analyze current database state and compare with production
```bash
./debug_import_process.sh
```
**Features:** Record counting, encoding validation, file analysis, production comparison

### 3. **import_monitor.sh** - REAL-TIME IMPORT MONITORING
**Purpose:** Monitor import process execution with detailed logging
```bash
# Monitor complete import process (default)
./import_monitor.sh monitor

# Show real-time status dashboard
./import_monitor.sh status  

# Monitor specific SQL file import
./import_monitor.sh sql /path/to/file.sql database_name
```
**Output:** Detailed import logs with SQL statement counting and progress tracking

### 4. **investigate_db_migration.sh** - PRODUCTION COMPARISON
**Purpose:** Compare local database with production (requires credentials)
```bash
./investigate_db_migration.sh -u production_user -p production_password
```
**Features:** Side-by-side database analysis, character encoding comparison

### 5. **quick_encoding_check.sh** - ENCODING VALIDATION
**Purpose:** Quick character encoding diagnostic
```bash
./quick_encoding_check.sh
```
**Features:** Database charset analysis, non-ASCII character detection

## ⚙️ **CONFIGURATION UPDATES**

### ✅ **Character Encoding Fixed**
Updated `.env` configuration to match production:
```env
MVDSB_CHARSET=latin1
PORTAL64_BDW_CHARSET=latin1
```

### 🔧 **Import Configuration Added**
Complete import configuration template added to `.env`:
- SCP connection settings (need production credentials)
- ZIP extraction passwords (need actual passwords)
- Enhanced debugging enabled
- Timeout and retry settings configured

## 🎯 **INVESTIGATION EXECUTION PLAN**

### **STEP 1: Basic Analysis**
```bash
# Quick database state check
./debug_import_process.sh

# Encoding validation
./quick_encoding_check.sh
```

### **STEP 2: Full Investigation**
```bash
# Complete 6-phase investigation
./investigate_import_process.sh
```
This will:
1. Check prerequisites
2. Analyze current database state  
3. Monitor import process (with user confirmation)
4. Perform post-import analysis
5. Identify root causes
6. Generate recommendations

### **STEP 3: Production Comparison** (Optional)
```bash
# Compare with production (requires credentials)
./investigate_db_migration.sh -u your_prod_user -p your_prod_password
```

### **STEP 4: Real-Time Monitoring** (During import)
```bash
# Monitor import in real-time
./import_monitor.sh status
```

## 📋 **REQUIRED ACTIONS BEFORE INVESTIGATION**

### 1. **Update .env Configuration**
```env
# Add your production credentials
IMPORT_SCP_USERNAME=your_production_username
IMPORT_SCP_PASSWORD=your_production_password

# Add actual ZIP passwords  
IMPORT_ZIP_PASSWORD_MVDSB=actual_mvdsb_password
IMPORT_ZIP_PASSWORD_PORTAL64_BDW=actual_portal64_password
```

### 2. **Ensure Server is Running**
```bash
# Start Portal64 API server
./bin/portal64api.exe

# Verify server is running
curl http://localhost:8080/api/v1/admin/import/status
```

### 3. **Make Scripts Executable** (Linux/Mac)
```bash
chmod +x *.sh
```

## 🔍 **EXPECTED INVESTIGATION OUTCOMES**

### **Likely Root Causes:**
1. **Incomplete SCP Download** - Files partially downloaded from portal.svw.info
2. **SQL Import Timeout** - Large imports hitting timeout limits
3. **Database Constraint Failures** - Duplicate keys or constraint violations during import
4. **Memory/Resource Limits** - Import process running out of memory
5. **ZIP Extraction Issues** - Corrupted or incomplete ZIP files

### **Investigation Will Identify:**
- Exact point where data loss occurs (download, extract, or import phase)
- SQL errors or warnings during import
- File size differences between production and local
- Performance bottlenecks or timeout issues
- Database constraint violations

## 📊 **ENHANCED LOGGING & METRICS**

The enhanced database importer provides detailed metrics:
- File analysis (size, line count, INSERT statement count)
- Pre/post import record counts
- SQL execution progress and errors
- Data quality validation (encoding issues, empty names)
- Import duration and performance metrics

## 🚀 **NEXT STEPS AFTER INVESTIGATION**

1. **Review Investigation Report** - Check `logs/investigation_*/investigation.log`
2. **Identify Root Cause** - Based on detailed logging and analysis
3. **Implement Fix** - Address specific issue found (credentials, timeouts, etc.)
4. **Re-test Import** - Verify fix resolves data loss issue
5. **Validate Data Completeness** - Ensure 521,365 records are imported successfully

## 📞 **TROUBLESHOOTING**

### **If Scripts Don't Run:**
- Windows: Use Git Bash or WSL
- Check file permissions
- Verify MySQL client is accessible

### **If MySQL Connection Fails:**
- Check XAMPP MySQL is running
- Verify credentials in .env
- Test connection: `mysql -u root -h localhost`

### **If API Server Not Responding:**
- Start server: `./bin/portal64api.exe`
- Check port 8080 availability
- Review server logs

## 🎯 **SUCCESS CRITERIA**

Investigation is successful when:
- ✅ Root cause of data loss identified
- ✅ Missing 28,519 records explained
- ✅ Fix implemented and tested
- ✅ Import achieves 521,365 total records (0% loss)
- ✅ Character encoding remains correct (latin1_german2_ci)

---

**Start Investigation:** `./investigate_import_process.sh`
