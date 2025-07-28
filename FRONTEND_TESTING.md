# Frontend Testing Setup for Portal64 API

This document explains how to set up and run the frontend tests for your HTML demo pages.

## Installation

1. **Install Node.js** (if not already installed):
   - Download from https://nodejs.org/ 
   - Choose LTS version (recommended)

2. **Install test dependencies**:
   ```bash
   cd C:\Users\tranm\work\svw.info\portal64api
   npm install
   ```

3. **Add Babel preset** (needed for ES6 module support):
   ```bash
   npm install --save-dev @babel/preset-env
   ```

## Running Tests

### Basic Commands

```bash
# Run all tests once
npm test

# Run tests in watch mode (re-runs on file changes)
npm run test:watch

# Run tests with coverage report
npm run test:coverage

# Run specific test file
npm test -- tests/frontend/unit/api/api.test.js

# Run tests matching a pattern
npm test -- --testNamePattern="Health Check"
```

### Test Output

The tests will provide:
- ✅ Pass/fail status for each test
- 📊 Coverage report showing tested code percentage
- 🐛 Detailed error messages for failing tests
- ⏱️ Performance metrics

## Test Structure

```
tests/frontend/
├── __mocks__/
│   ├── server.js          # MSW API mock server
│   ├── mockData.js        # Test data fixtures
│   └── fileMock.js        # Static file mocks
├── setup/
│   ├── jest.setup.js      # Main Jest configuration
│   └── jest.globals.js    # Global test environment setup
└── unit/
    ├── api/
    │   └── api.test.js     # API client tests
    └── pages/
        └── index.test.js   # Dashboard page tests
```

## What Gets Tested

### API Client Tests (`api.test.js`)
- ✅ API client initialization
- ✅ HTTP request/response handling
- ✅ Error handling and network failures
- ✅ Player, club, and tournament search
- ✅ Individual record retrieval
- ✅ Parameter validation

### Page Tests (`index.test.js`)
- ✅ HTML structure and DOM elements
- ✅ Health check functionality
- ✅ Quick test buttons
- ✅ Navigation links
- ✅ Error display and user feedback
- ✅ CSS and JavaScript loading

## Creating New Tests

### For New HTML Pages

1. Create test file: `tests/frontend/unit/pages/[pagename].test.js`
2. Follow this pattern:

```javascript
describe('Page Name', () => {
  let dom, document, window;

  beforeEach(() => {
    // Load HTML file
    const htmlPath = path.resolve(__dirname, '../../../../demo/pagename.html');
    const htmlContent = fs.readFileSync(htmlPath, 'utf8');
    
    // Create JSDOM instance
    dom = new JSDOM(htmlContent, {
      url: 'http://localhost:3000',
      pretendToBeVisual: true
    });
    
    document = dom.window.document;
    window = dom.window;
  });

  test('should have correct page structure', () => {
    // Test DOM elements exist
    expect(document.querySelector('h1')).toBeTruthy();
  });
});
```

### For New JavaScript Functions

1. Create test file: `tests/frontend/unit/[module]/[function].test.js`
2. Load and test the function:

```javascript
// Load the actual JS file
const jsPath = path.resolve(__dirname, '../../../../demo/js/module.js');
const jsCode = fs.readFileSync(jsPath, 'utf8');
eval(jsCode);

describe('Function Name', () => {
  test('should work correctly', () => {
    const result = functionName(input);
    expect(result).toBe(expectedOutput);
  });
});
```

## Debugging Tests

### Common Issues

1. **Module not found errors**:
   - Check file paths in test files
   - Ensure HTML/JS files exist in demo folder

2. **Fetch is not defined**:
   - Make sure `jest-fetch-mock` is installed
   - Check Jest setup files are configured correctly

3. **JSDOM errors**:
   - Verify HTML is valid
   - Check for missing global variables

### Debug Commands

```bash
# Run single test with debug info
npm test -- --verbose tests/frontend/unit/api/api.test.js

# Debug specific test case
npm test -- --testNamePattern="should perform health check"

# Run tests with console output visible
npm test -- --silent=false
```

## Coverage Reports

After running `npm run test:coverage`, check:
```
coverage/
├── lcov-report/
│   └── index.html     # Visual coverage report
└── lcov.info          # Raw coverage data
```

Open `coverage/lcov-report/index.html` in browser to see:
- 📊 Line coverage percentage
- 📈 Function coverage
- 📉 Branch coverage
- 🔍 Uncovered code highlighting

## Integration with Development

### Pre-commit Testing
Add to your workflow:
```bash
# Before committing changes
npm run test:ci
```

### Continuous Integration
For automated testing, the `npm run test:ci` command:
- Runs all tests once (no watch mode)
- Generates coverage reports
- Exits with proper status codes
- Suitable for CI/CD pipelines

## Best Practices

### Writing Good Tests

1. **Test user behavior, not implementation**:
   ```javascript
   // Good: Test what user sees
   expect(document.querySelector('.error-message')).toBeVisible();
   
   // Bad: Test internal implementation
   expect(someInternalVariable).toBe(true);
   ```

2. **Use descriptive test names**:
   ```javascript
   // Good
   test('should display error message when API call fails', () => {});
   
   // Bad  
   test('error test', () => {});
   ```

3. **Mock external dependencies**:
   - Always mock API calls
   - Mock browser APIs (localStorage, etc.)
   - Use consistent test data

4. **Test edge cases**:
   - Empty responses
   - Network failures
   - Invalid input data
   - Error conditions

### Maintaining Tests

1. **Keep tests simple and focused**
2. **Update tests when functionality changes**
3. **Remove obsolete tests**
4. **Ensure tests run quickly (< 5 seconds total)**

## Next Steps

1. **Run the basic setup**:
   ```bash
   npm install
   npm test
   ```

2. **Add tests for remaining pages**:
   - players.html
   - clubs.html  
   - tournaments.html
   - api-docs.html

3. **Add integration tests** for complete user workflows

4. **Set up automated testing** in your development process

## Troubleshooting

If you encounter issues:

1. **Check Node.js version**: `node --version` (should be 16+)
2. **Clear npm cache**: `npm cache clean --force`
3. **Reinstall dependencies**: `rm -rf node_modules && npm install`
4. **Verify file paths** in test files match your actual structure

For questions or issues, check the test output carefully - Jest provides detailed error messages that usually point to the exact problem.
