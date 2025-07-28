// MSW server setup for mocking API calls
import { rest } from 'msw';
import { setupServer } from 'msw/node';

// Mock data
import { mockPlayers, mockClubs, mockTournaments } from './mockData';

const baseURL = 'http://localhost:8080';

// Define request handlers
const handlers = [
  // Health check endpoint
  rest.get(`${baseURL}/health`, (req, res, ctx) => {
    return res(
      ctx.status(200),
      ctx.json({
        status: 'ok',
        timestamp: Date.now(),
        version: '1.0.0',
        database: 'connected'
      })
    );
  }),

  // Players endpoints
  rest.get(`${baseURL}/api/v1/players`, (req, res, ctx) => {
    const limit = parseInt(req.url.searchParams.get('limit')) || 10;
    const offset = parseInt(req.url.searchParams.get('offset')) || 0;
    const search = req.url.searchParams.get('search') || '';
    
    let filteredPlayers = mockPlayers;
    if (search) {
      filteredPlayers = mockPlayers.filter(player => 
        player.name.toLowerCase().includes(search.toLowerCase()) ||
        player.first_name.toLowerCase().includes(search.toLowerCase())
      );
    }
    
    const paginatedPlayers = filteredPlayers.slice(offset, offset + limit);
    
    return res(
      ctx.status(200),
      ctx.json({
        success: true,
        data: paginatedPlayers,
        meta: {
          total: filteredPlayers.length,
          limit,
          offset,
          has_more: offset + limit < filteredPlayers.length
        }
      })
    );
  }),

  // Single player endpoint
  rest.get(`${baseURL}/api/v1/players/:id`, (req, res, ctx) => {
    const { id } = req.params;
    const player = mockPlayers.find(p => p.id === id);
    
    if (!player) {
      return res(
        ctx.status(404),
        ctx.json({
          success: false,
          error: 'Player not found',
          code: 'PLAYER_NOT_FOUND'
        })
      );
    }
    
    return res(
      ctx.status(200),
      ctx.json({
        success: true,
        data: player
      })
    );
  }),

  // Clubs endpoints
  rest.get(`${baseURL}/api/v1/clubs`, (req, res, ctx) => {
    const limit = parseInt(req.url.searchParams.get('limit')) || 10;
    const offset = parseInt(req.url.searchParams.get('offset')) || 0;
    
    const paginatedClubs = mockClubs.slice(offset, offset + limit);
    
    return res(
      ctx.status(200),
      ctx.json({
        success: true,
        data: paginatedClubs,
        meta: {
          total: mockClubs.length,
          limit,
          offset,
          has_more: offset + limit < mockClubs.length
        }
      })
    );
  }),

  // Tournaments endpoints
  rest.get(`${baseURL}/api/v1/tournaments`, (req, res, ctx) => {
    const limit = parseInt(req.url.searchParams.get('limit')) || 10;
    const offset = parseInt(req.url.searchParams.get('offset')) || 0;
    
    const paginatedTournaments = mockTournaments.slice(offset, offset + limit);
    
    return res(
      ctx.status(200),
      ctx.json({
        success: true,
        data: paginatedTournaments,
        meta: {
          total: mockTournaments.length,
          limit,
          offset,
          has_more: offset + limit < mockTournaments.length
        }
      })
    );
  }),

  // Error handler for unhandled requests
  rest.get('*', (req, res, ctx) => {
    console.error(`Unhandled request: ${req.method} ${req.url}`);
    return res(
      ctx.status(404),
      ctx.json({
        success: false,
        error: 'Endpoint not found',
        code: 'NOT_FOUND'
      })
    );
  })
];

// Setup MSW server
export const server = setupServer(...handlers);
