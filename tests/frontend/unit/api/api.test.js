/**
 * @fileoverview Tests for the Portal64API client class
 * Tests API interactions, error handling, and response parsing
 */

// Mock the API client by loading the actual file
const fs = require('fs');
const path = require('path');

// Load the actual API client code
const apiClientPath = path.resolve(__dirname, '../../../../demo/js/api.js');
const apiClientCode = fs.readFileSync(apiClientPath, 'utf8');

// Execute the API client code in our test environment
eval(apiClientCode);

describe('Portal64API Client', () => {
  let apiClient;

  beforeEach(() => {
    // Reset fetch mock before each test
    fetch.mockClear();
    
    // Create fresh API client instance
    apiClient = new Portal64API();
  });

  describe('Constructor', () => {
    test('should initialize with correct base URL', () => {
      expect(apiClient.baseURL).toBe('http://localhost:8080');
    });

    test('should set default headers', () => {
      expect(apiClient.defaultHeaders).toEqual({
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      });
    });
  });

  describe('Health Check', () => {
    test('should perform health check successfully', async () => {
      const mockResponse = {
        status: 'ok',
        timestamp: Date.now(),
        version: '1.0.0'
      };

      fetch.mockResponseOnce(JSON.stringify(mockResponse));

      const result = await apiClient.healthCheck();

      expect(fetch).toHaveBeenCalledWith(
        'http://localhost:8080/health',
        expect.objectContaining({
          method: 'GET',
          headers: apiClient.defaultHeaders
        })
      );
      expect(result).toEqual(mockResponse);
    });

    test('should handle health check failure', async () => {
      fetch.mockRejectOnce(new Error('Network error'));

      await expect(apiClient.healthCheck()).rejects.toThrow('Network error');
    });
  });

  describe('Player Search', () => {
    test('should search players with default parameters', async () => {
      const mockResponse = {
        success: true,
        data: [
          { id: 'C0101-1014', name: 'Test Player', rating: 1500 }
        ],
        meta: {
          total: 1,
          limit: 10,
          offset: 0
        }
      };

      fetch.mockResponseOnce(JSON.stringify(mockResponse));

      const result = await apiClient.searchPlayers();

      expect(fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/players?limit=10&offset=0',
        expect.objectContaining({
          method: 'GET'
        })
      );
      expect(result).toEqual(mockResponse);
    });

    test('should search players with custom parameters', async () => {
      const searchParams = {
        search: 'Test Player',
        limit: 5,
        offset: 10,
        sort: 'rating'
      };

      const mockResponse = {
        success: true,
        data: [],
        meta: {
          total: 0,
          limit: 5,
          offset: 10
        }
      };

      fetch.mockResponseOnce(JSON.stringify(mockResponse));

      const result = await apiClient.searchPlayers(searchParams);

      expect(fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/players?search=Test%20Player&limit=5&offset=10&sort=rating',
        expect.objectContaining({
          method: 'GET'
        })
      );
      expect(result).toEqual(mockResponse);
    });

    test('should handle API error response', async () => {
      const errorResponse = {
        success: false,
        error: 'Invalid search parameters',
        code: 'INVALID_PARAMS'
      };

      fetch.mockResponseOnce(
        JSON.stringify(errorResponse),
        { status: 400 }
      );

      await expect(apiClient.searchPlayers({ search: '' }))
        .rejects.toThrow('Invalid search parameters');
    });
  });

  describe('Get Player by ID', () => {
    test('should get player by valid ID', async () => {
      const playerId = 'C0101-1014';
      const mockPlayer = {
        id: playerId,
        name: 'Test Player',
        rating: 1500,
        club_name: 'Test Club'
      };

      const mockResponse = {
        success: true,
        data: mockPlayer
      };

      fetch.mockResponseOnce(JSON.stringify(mockResponse));

      const result = await apiClient.getPlayer(playerId);

      expect(fetch).toHaveBeenCalledWith(
        `http://localhost:8080/api/v1/players/${playerId}`,
        expect.objectContaining({
          method: 'GET'
        })
      );
      expect(result).toEqual(mockResponse);
    });

    test('should handle player not found', async () => {
      const playerId = 'INVALID-ID';
      const errorResponse = {
        success: false,
        error: 'Player not found',
        code: 'PLAYER_NOT_FOUND'
      };

      fetch.mockResponseOnce(
        JSON.stringify(errorResponse),
        { status: 404 }
      );

      await expect(apiClient.getPlayer(playerId))
        .rejects.toThrow('Player not found');
    });
  });

  describe('Request Method', () => {
    test('should handle successful requests', async () => {
      const mockData = { test: 'data' };
      fetch.mockResponseOnce(JSON.stringify(mockData));

      const result = await apiClient.request('/test-endpoint');

      expect(result).toEqual(mockData);
    });

    test('should handle network errors', async () => {
      fetch.mockRejectOnce(new Error('Network failure'));

      await expect(apiClient.request('/test-endpoint'))
        .rejects.toThrow('Network failure');
    });

    test('should handle HTTP error status codes', async () => {
      const errorData = {
        error: 'Bad Request',
        details: 'Invalid parameters'
      };

      fetch.mockResponseOnce(
        JSON.stringify(errorData),
        { status: 400 }
      );

      await expect(apiClient.request('/test-endpoint'))
        .rejects.toThrow('Bad Request');
    });

    test('should use custom options', async () => {
      const mockData = { result: 'success' };
      fetch.mockResponseOnce(JSON.stringify(mockData));

      const customOptions = {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json', 
          'Authorization': 'Bearer token'
        },
        body: JSON.stringify({ test: 'data' })
      };

      await apiClient.request('/test-endpoint', customOptions);

      expect(fetch).toHaveBeenCalledWith(
        'http://localhost:8080/test-endpoint',
        expect.objectContaining({
          method: 'POST',
          headers: expect.objectContaining({
            'Authorization': 'Bearer token'
          }),
          body: JSON.stringify({ test: 'data' })
        })
      );
    });
  });
});
