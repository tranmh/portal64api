// Jest setup file - runs after the test framework is set up
import '@testing-library/jest-dom';

// Setup MSW (Mock Service Worker) for API mocking
import { server } from '../__mocks__/server';

// Establish API mocking before all tests
beforeAll(() => {
  server.listen({
    onUnhandledRequest: 'error'
  });
});

// Reset any request handlers that are declared during tests
afterEach(() => {
  server.resetHandlers();
});

// Clean up after the tests are finished
afterAll(() => {
  server.close();
});

// Mock window.location and other browser APIs
Object.defineProperty(window, 'location', {
  value: {
    href: 'http://localhost:3000',
    origin: 'http://localhost:3000',
    protocol: 'http:',
    host: 'localhost:3000',
    hostname: 'localhost',
    port: '3000',
    pathname: '/',
    search: '',
    hash: '',
    assign: jest.fn(),
    replace: jest.fn(),
    reload: jest.fn()
  },
  writable: true
});

// Mock localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn()
};
global.localStorage = localStorageMock;

// Mock sessionStorage
const sessionStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn()
};
global.sessionStorage = sessionStorageMock;

// Extend expect with custom matchers if needed
expect.extend({
  toBeVisible(element) {
    const pass = element.style.display !== 'none' && 
                 element.hidden !== true &&
                 element.style.visibility !== 'hidden';
    
    return {
      message: () =>
        pass
          ? `expected element not to be visible`
          : `expected element to be visible`,
      pass,
    };
  }
});
