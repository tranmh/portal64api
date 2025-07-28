// Add TextEncoder/TextDecoder polyfills for Node.js compatibility
const { TextEncoder, TextDecoder } = require('util');
global.TextEncoder = TextEncoder;
global.TextDecoder = TextDecoder;

// Setup jest-fetch-mock
const fetchMock = require('jest-fetch-mock');
fetchMock.enableMocks();

// Mock console methods to reduce noise in tests
global.console = {
  ...console,
  // Suppress console.log in tests unless specifically needed
  log: jest.fn(),
  debug: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
  error: jest.fn()
};
