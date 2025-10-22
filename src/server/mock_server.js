const http = require('http');
const url = require('url');

// Mock data for analytics
const mockStatsData = {
  overview: {
    total_books: 25,
    books_read: 18,
    books_in_progress: 4,
    total_pages_read: 4387,
    total_reading_time_minutes: 2640,
    average_reading_speed_pages_per_hour: 45,
    current_reading_streak_days: 12,
    longest_reading_streak_days: 28
  },
  reading_habits: {
    daily_averages: {
      pages_per_day: 23,
      minutes_per_day: 45,
      books_per_month: 2.5
    },
    monthly_progress: [
      { month: "Jan 2024", pages_read: 320, books_completed: 2, reading_time_minutes: 480 },
      { month: "Feb 2024", pages_read: 280, books_completed: 1, reading_time_minutes: 420 },
      { month: "Mar 2024", pages_read: 450, books_completed: 3, reading_time_minutes: 675 },
      { month: "Apr 2024", pages_read: 380, books_completed: 2, reading_time_minutes: 570 },
      { month: "May 2024", pages_read: 520, books_completed: 4, reading_time_minutes: 780 },
      { month: "Jun 2024", pages_read: 290, books_completed: 2, reading_time_minutes: 435 }
    ],
    reading_by_day_of_week: [
      { day: "Monday", pages_read: 45, sessions: 8 },
      { day: "Tuesday", pages_read: 38, sessions: 6 },
      { day: "Wednesday", pages_read: 52, sessions: 9 },
      { day: "Thursday", pages_read: 41, sessions: 7 },
      { day: "Friday", pages_read: 35, sessions: 5 },
      { day: "Saturday", pages_read: 67, sessions: 12 },
      { day: "Sunday", pages_read: 58, sessions: 10 }
    ],
    reading_by_hour: Array.from({length: 24}, (_, i) => ({
      hour: i,
      pages_read: i < 6 || i > 22 ? Math.floor(Math.random() * 5) : Math.floor(Math.random() * 25) + 5,
      sessions: i < 6 || i > 22 ? Math.floor(Math.random() * 2) : Math.floor(Math.random() * 8) + 1
    }))
  },
  content_analytics: {
    genres: [
      { genre: "Philosophy", book_count: 8, pages_read: 1650, completion_rate: 75 },
      { genre: "Classical Literature", book_count: 6, pages_read: 1200, completion_rate: 83 },
      { genre: "History", book_count: 4, pages_read: 890, completion_rate: 90 },
      { genre: "Theology", book_count: 3, pages_read: 567, completion_rate: 67 },
      { genre: "Poetry", book_count: 4, pages_read: 780, completion_rate: 95 }
    ],
    languages: [
      { language: "English", book_count: 15, pages_read: 2890, reading_time_minutes: 1734 },
      { language: "Latin", book_count: 5, pages_read: 890, reading_time_minutes: 534 },
      { language: "Greek", book_count: 3, pages_read: 450, reading_time_minutes: 270 },
      { language: "German", book_count: 2, pages_read: 157, reading_time_minutes: 102 }
    ],
    authors: [
      { author: "Plato", book_count: 3, pages_read: 567, favorite_book: "The Republic" },
      { author: "Aristotle", book_count: 2, pages_read: 445, favorite_book: "Nicomachean Ethics" },
      { author: "Marcus Aurelius", book_count: 1, pages_read: 234, favorite_book: "Meditations" },
      { author: "Augustine", book_count: 2, pages_read: 678, favorite_book: "Confessions" },
      { author: "Thomas Aquinas", book_count: 1, pages_read: 389, favorite_book: "Summa Theologica" }
    ],
    file_formats: [
      { format: "PDF", count: 12, total_size_mb: 145.7 },
      { format: "EPUB", count: 8, total_size_mb: 67.3 },
      { format: "TXT", count: 5, total_size_mb: 12.1 }
    ]
  },
  goals_and_progress: {
    yearly_goal: {
      target_books: 24,
      completed_books: 18,
      progress_percentage: 75,
      projected_completion: "November 2024"
    },
    daily_goal: {
      target_pages: 25,
      average_achieved: 23,
      success_rate: 78
    },
    reading_milestones: [
      { milestone: "Read 10 books", achieved_date: "2024-04-15", progress: 100 },
      { milestone: "Read 1000 pages", achieved_date: "2024-03-22", progress: 100 },
      { milestone: "7-day reading streak", achieved_date: "2024-05-10", progress: 100 },
      { milestone: "Read 20 books", progress: 90 },
      { milestone: "Read 5000 pages", progress: 87 }
    ]
  },
  annotations_and_engagement: {
    total_annotations: 234,
    highlights: 156,
    notes: 67,
    bookmarks: 11,
    most_annotated_books: [
      { title: "The Republic", author: "Plato", annotation_count: 45 },
      { title: "Meditations", author: "Marcus Aurelius", annotation_count: 38 },
      { title: "Confessions", author: "Augustine", annotation_count: 32 },
      { title: "Nicomachean Ethics", author: "Aristotle", annotation_count: 28 },
      { title: "The Aeneid", author: "Virgil", annotation_count: 22 }
    ]
  }
};

const server = http.createServer((req, res) => {
  // Enable CORS
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Origin, Content-Type, Authorization, X-Requested-With, Accept');
  
  if (req.method === 'OPTIONS') {
    res.writeHead(200);
    res.end();
    return;
  }

  const parsedUrl = url.parse(req.url, true);
  const path = parsedUrl.pathname;
  
  console.log(`${req.method} ${path}`);

  // Handle registration
  if (path === '/api/v1/auth/register' && req.method === 'POST') {
    let body = '';
    req.on('data', chunk => { body += chunk; });
    req.on('end', () => {
      const userData = JSON.parse(body);
      const mockUser = {
        id: 'user-' + Date.now(),
        username: userData.username,
        email: userData.email,
        full_name: userData.full_name || userData.username,
        avatar_url: null,
        subscription_tier: 'free',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      };
      
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({
        success: true,
        message: 'Registration successful',
        data: {
          user: mockUser,
          access_token: 'mock-access-token-' + Date.now(),
          refresh_token: 'mock-refresh-token-' + Date.now(),
          expires_in: 86400
        }
      }));
    });
    return;
  }

  // Handle login
  if (path === '/api/v1/auth/login' && req.method === 'POST') {
    let body = '';
    req.on('data', chunk => { body += chunk; });
    req.on('end', () => {
      const credentials = JSON.parse(body);
      const mockUser = {
        id: 'user-12345',
        username: 'demo_user',
        email: credentials.email,
        full_name: 'Demo User',
        avatar_url: null,
        subscription_tier: 'free',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: new Date().toISOString()
      };
      
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({
        success: true,
        message: 'Login successful',
        data: {
          user: mockUser,
          access_token: 'mock-access-token-' + Date.now(),
          refresh_token: 'mock-refresh-token-' + Date.now(),
          expires_in: 86400
        }
      }));
    });
    return;
  }

  // Handle logout
  if (path === '/api/v1/auth/logout' && req.method === 'POST') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({
      success: true,
      message: 'Logged out successfully'
    }));
    return;
  }

  if (path === '/api/stats/books') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({
      success: true,
      data: mockStatsData
    }));
  } else if (path === '/api/health') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({
      status: 'ok',
      service: 'classius-mock-server',
      version: '0.1.0',
      time: new Date().toISOString()
    }));
  } else {
    res.writeHead(404, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({
      success: false,
      error: 'Not found'
    }));
  }
});

const PORT = 8082;
server.listen(PORT, () => {
  console.log(`🚀 Mock Classius API server running on port ${PORT}`);
  console.log(`📊 Analytics endpoint: http://localhost:${PORT}/api/stats/books`);
  console.log(`❤️  Health endpoint: http://localhost:${PORT}/api/health`);
});
