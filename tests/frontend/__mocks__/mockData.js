// Mock data for testing
export const mockPlayers = [
  {
    id: 'C0101-1014',
    name: 'Müller, Hans',
    first_name: 'Hans',
    club_id: 'C0101',
    club_name: 'Post-SV Ulm',
    rating: 1856,
    rating_type: 'DWZ',
    last_tournament: '2024-03-15',
    games_played: 45,
    status: 'active'
  },
  {
    id: 'C0102-2034',
    name: 'Schmidt, Anna',
    first_name: 'Anna',
    club_id: 'C0102',
    club_name: 'SC Friedrichshafen',
    rating: 1642,
    rating_type: 'DWZ',
    last_tournament: '2024-02-28',
    games_played: 32,
    status: 'active'
  },
  {
    id: 'C0103-3056',
    name: 'Weber, Michael',
    first_name: 'Michael',
    club_id: 'C0103',
    club_name: 'Schachfreunde Stuttgart',
    rating: 1789,
    rating_type: 'DWZ',
    last_tournament: '2024-03-20',
    games_played: 58,
    status: 'active'
  },
  {
    id: 'C0104-4078',
    name: 'Fischer, Lisa',  
    first_name: 'Lisa',
    club_id: 'C0104',
    club_name: 'TSV Beilstein',
    rating: 1534,
    rating_type: 'DWZ',
    last_tournament: '2024-01-15',
    games_played: 28,
    status: 'active'
  },
  {
    id: 'C0105-5090',
    name: 'Wagner, Thomas',
    first_name: 'Thomas',
    club_id: 'C0105',
    club_name: 'Schachclub Esslingen',
    rating: 1923,
    rating_type: 'DWZ',
    last_tournament: '2024-03-10',
    games_played: 67,
    status: 'active'
  }
];

export const mockClubs = [
  {
    id: 'C0101',
    name: 'Post-SV Ulm',
    short_name: 'Post Ulm',
    city: 'Ulm',
    region: 'Württemberg',
    district: 'Alb-Donau',
    founded: '1965-03-15',
    member_count: 45,
    active_players: 38,
    contact_email: 'info@post-sv-ulm.de',
    website: 'https://post-sv-ulm.de',
    status: 'active'
  },
  {
    id: 'C0102',
    name: 'SC Friedrichshafen',
    short_name: 'SC FN',
    city: 'Friedrichshafen',
    region: 'Württemberg',
    district: 'Bodensee',
    founded: '1952-08-20',
    member_count: 62,
    active_players: 54,
    contact_email: 'kontakt@sc-friedrichshafen.de',
    website: 'https://sc-friedrichshafen.de',
    status: 'active'
  },
  {
    id: 'C0103',
    name: 'Schachfreunde Stuttgart',
    short_name: 'SF Stuttgart',
    city: 'Stuttgart',
    region: 'Württemberg',
    district: 'Stuttgart',
    founded: '1923-11-12',
    member_count: 89,
    active_players: 76,
    contact_email: 'info@schachfreunde-stuttgart.de',
    website: 'https://schachfreunde-stuttgart.de',
    status: 'active'
  }
];

export const mockTournaments = [
  {
    id: 'C529-K00-HT1',
    name: 'Württembergische Einzelmeisterschaft 2024',
    short_name: 'WEM 2024',
    organizer: 'Schachverband Württemberg',
    location: 'Stuttgart',
    start_date: '2024-04-15',
    end_date: '2024-04-17',
    tournament_type: 'Swiss System',
    rounds: 9,
    time_control: '90+30',
    entry_fee: 45,
    max_participants: 120,
    current_participants: 87,
    registration_deadline: '2024-04-01',
    status: 'upcoming',
    rating_relevant: true
  },
  {
    id: 'C530-K01-HT2',
    name: 'Ulmer Stadtmeisterschaft',
    short_name: 'USM 2024',
    organizer: 'Post-SV Ulm',
    location: 'Ulm',
    start_date: '2024-03-22',
    end_date: '2024-03-24',
    tournament_type: 'Round Robin',
    rounds: 7,
    time_control: '120+30',
    entry_fee: 25,
    max_participants: 24,
    current_participants: 24,
    registration_deadline: '2024-03-15',
    status: 'completed',
    rating_relevant: true
  },
  {
    id: 'C531-K02-HT3',
    name: 'Schnellschach Grand Prix Stuttgart',
    short_name: 'GP Stuttgart',
    organizer: 'Schachfreunde Stuttgart',
    location: 'Stuttgart',
    start_date: '2024-05-05',
    end_date: '2024-05-05',
    tournament_type: 'Swiss System',
    rounds: 7,
    time_control: '15+10',
    entry_fee: 15,
    max_participants: 60,
    current_participants: 23,
    registration_deadline: '2024-05-01',
    status: 'upcoming',
    rating_relevant: false
  }
];

// Mock API responses for different scenarios
export const mockApiResponses = {
  healthCheck: {
    success: {
      status: 'ok',
      timestamp: Date.now(),
      version: '1.0.0',
      database: 'connected'
    },
    error: {
      status: 'error',
      message: 'Database connection failed',
      timestamp: Date.now()
    }
  },
  
  playersSearch: {
    success: {
      success: true,
      data: mockPlayers.slice(0, 5),
      meta: {
        total: mockPlayers.length,
        limit: 5,
        offset: 0,
        has_more: true
      }
    },
    empty: {
      success: true,
      data: [],
      meta: {
        total: 0,
        limit: 10,
        offset: 0,
        has_more: false
      }
    },
    error: {
      success: false,
      error: 'Search query too short',
      code: 'INVALID_QUERY'
    }
  },

  serverError: {
    success: false,
    error: 'Internal server error',
    code: 'INTERNAL_ERROR'
  }
};
