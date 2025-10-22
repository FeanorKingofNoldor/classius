'use client';

import React from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  Area,
  AreaChart,
  RadialBarChart,
  RadialBar,
} from 'recharts';

// Monthly Progress Chart
interface MonthlyProgressProps {
  data: Array<{
    month: string;
    pages_read: number;
    books_completed: number;
    reading_time_minutes: number;
  }>;
}

export const MonthlyProgressChart: React.FC<MonthlyProgressProps> = ({ data }) => {
  const formatTime = (minutes: number) => {
    const hours = Math.floor(minutes / 60);
    return hours > 0 ? `${hours}h` : `${minutes}m`;
  };

  return (
    <ResponsiveContainer width="100%" height={300}>
      <LineChart data={data}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="month" />
        <YAxis />
        <Tooltip 
          labelFormatter={(label) => `Month: ${label}`}
          formatter={(value, name) => {
            if (name === 'Reading Time') return [formatTime(value as number), name];
            return [value, name];
          }}
        />
        <Legend />
        <Line 
          type="monotone" 
          dataKey="pages_read" 
          stroke="#8b5cf6" 
          strokeWidth={2}
          name="Pages Read"
          dot={{ fill: '#8b5cf6' }}
        />
        <Line 
          type="monotone" 
          dataKey="books_completed" 
          stroke="#10b981" 
          strokeWidth={2}
          name="Books Completed"
          dot={{ fill: '#10b981' }}
        />
        <Line 
          type="monotone" 
          dataKey="reading_time_minutes" 
          stroke="#f59e0b" 
          strokeWidth={2}
          name="Reading Time"
          dot={{ fill: '#f59e0b' }}
        />
      </LineChart>
    </ResponsiveContainer>
  );
};

// Reading Habits Heatmap-style Chart
interface ReadingHabitsProps {
  weeklyData: Array<{
    day: string;
    pages_read: number;
    sessions: number;
  }>;
  hourlyData: Array<{
    hour: number;
    pages_read: number;
    sessions: number;
  }>;
}

export const ReadingHabitsChart: React.FC<ReadingHabitsProps> = ({ weeklyData, hourlyData }) => {
  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {/* Weekly Reading Pattern */}
      <div className="bg-white p-4 rounded-lg">
        <h4 className="text-lg font-medium text-gray-900 mb-4">Weekly Reading Pattern</h4>
        <ResponsiveContainer width="100%" height={250}>
          <BarChart data={weeklyData} layout="horizontal">
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis type="number" />
            <YAxis dataKey="day" type="category" />
            <Tooltip />
            <Bar dataKey="pages_read" fill="#6366f1" name="Pages Read" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      {/* Hourly Reading Pattern */}
      <div className="bg-white p-4 rounded-lg">
        <h4 className="text-lg font-medium text-gray-900 mb-4">Preferred Reading Hours</h4>
        <ResponsiveContainer width="100%" height={250}>
          <AreaChart data={hourlyData.slice(6, 24)}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis 
              dataKey="hour" 
              tickFormatter={(hour) => `${hour}:00`}
            />
            <YAxis />
            <Tooltip 
              labelFormatter={(hour) => `${hour}:00`}
              formatter={(value) => [value, 'Sessions']}
            />
            <Area 
              type="monotone" 
              dataKey="sessions" 
              stroke="#10b981" 
              fill="#10b981" 
              fillOpacity={0.6}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
};

// Genre Distribution Pie Chart
interface GenreDistributionProps {
  data: Array<{
    genre: string;
    book_count: number;
    pages_read: number;
    completion_rate: number;
  }>;
}

export const GenreDistributionChart: React.FC<GenreDistributionProps> = ({ data }) => {
  const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884d8', '#82ca9d'];
  
  const pieData = data.map((item, index) => ({
    name: item.genre,
    value: item.pages_read,
    color: COLORS[index % COLORS.length]
  }));

  return (
    <div className="bg-white p-6 rounded-lg">
      <h4 className="text-lg font-medium text-gray-900 mb-4">Reading by Genre</h4>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ResponsiveContainer width="100%" height={300}>
          <PieChart>
            <Pie
              data={pieData}
              cx="50%"
              cy="50%"
              labelLine={false}
              label={({ name, percent }) => `${name} (${(percent * 100).toFixed(0)}%)`}
              outerRadius={80}
              fill="#8884d8"
              dataKey="value"
            >
              {pieData.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={entry.color} />
              ))}
            </Pie>
            <Tooltip formatter={(value) => [`${value} pages`, 'Pages Read']} />
          </PieChart>
        </ResponsiveContainer>

        {/* Genre completion rates */}
        <div className="space-y-3">
          {data.map((genre, index) => (
            <div key={genre.genre} className="flex items-center justify-between">
              <div className="flex items-center space-x-3">
                <div 
                  className="w-4 h-4 rounded-full"
                  style={{ backgroundColor: COLORS[index % COLORS.length] }}
                />
                <span className="text-sm font-medium text-gray-900">{genre.genre}</span>
              </div>
              <div className="text-right">
                <div className="text-sm font-medium text-gray-900">{genre.completion_rate}%</div>
                <div className="text-xs text-gray-500">{genre.book_count} books</div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

// Reading Goal Progress
interface ReadingGoalProps {
  yearlyGoal?: {
    target_books: number;
    completed_books: number;
    progress_percentage: number;
    projected_completion: string;
  };
  dailyGoal?: {
    target_pages: number;
    average_achieved: number;
    success_rate: number;
  };
}

export const ReadingGoalChart: React.FC<ReadingGoalProps> = ({ yearlyGoal, dailyGoal }) => {
  const goalData = [];
  
  if (yearlyGoal) {
    goalData.push({
      name: 'Yearly Goal',
      progress: yearlyGoal.progress_percentage,
      completed: yearlyGoal.completed_books,
      target: yearlyGoal.target_books
    });
  }
  
  if (dailyGoal) {
    goalData.push({
      name: 'Daily Goal',
      progress: dailyGoal.success_rate,
      completed: dailyGoal.average_achieved,
      target: dailyGoal.target_pages
    });
  }

  return (
    <div className="bg-white p-6 rounded-lg">
      <h4 className="text-lg font-medium text-gray-900 mb-4">Goal Progress</h4>
      <ResponsiveContainer width="100%" height={200}>
        <RadialBarChart cx="50%" cy="50%" innerRadius="20%" outerRadius="90%" data={goalData}>
          <RadialBar
            minAngle={15}
            label={{ position: 'insideStart', fill: '#fff' }}
            background
            clockWise
            dataKey="progress"
            fill="#6366f1"
          />
          <Legend />
          <Tooltip 
            formatter={(value) => [`${value}%`, 'Progress']}
          />
        </RadialBarChart>
      </ResponsiveContainer>
    </div>
  );
};

// Reading Streak Visualization
interface ReadingStreakProps {
  currentStreak: number;
  longestStreak: number;
  streakHistory?: Array<{
    date: string;
    hasRead: boolean;
    pagesRead: number;
  }>;
}

export const ReadingStreakChart: React.FC<ReadingStreakProps> = ({ 
  currentStreak, 
  longestStreak, 
  streakHistory = [] 
}) => {
  // Generate last 30 days for visualization
  const generateStreakData = () => {
    const data = [];
    const today = new Date();
    
    for (let i = 29; i >= 0; i--) {
      const date = new Date(today);
      date.setDate(date.getDate() - i);
      
      const dayData = streakHistory.find(h => 
        new Date(h.date).toDateString() === date.toDateString()
      );
      
      data.push({
        date: date.getDate(),
        hasRead: dayData?.hasRead || Math.random() > 0.3, // Mock data if not provided
        pagesRead: dayData?.pagesRead || (Math.random() > 0.3 ? Math.floor(Math.random() * 50) + 1 : 0)
      });
    }
    
    return data;
  };

  const streakData = generateStreakData();

  return (
    <div className="bg-white p-6 rounded-lg">
      <h4 className="text-lg font-medium text-gray-900 mb-4">Reading Streak</h4>
      
      {/* Streak Stats */}
      <div className="grid grid-cols-2 gap-4 mb-6">
        <div className="text-center p-4 bg-green-50 rounded-lg">
          <div className="text-2xl font-bold text-green-600">{currentStreak}</div>
          <div className="text-sm text-gray-600">Current Streak</div>
        </div>
        <div className="text-center p-4 bg-orange-50 rounded-lg">
          <div className="text-2xl font-bold text-orange-600">{longestStreak}</div>
          <div className="text-sm text-gray-600">Best Streak</div>
        </div>
      </div>

      {/* Last 30 Days Heatmap-style */}
      <div className="space-y-2">
        <div className="text-sm text-gray-600 mb-2">Last 30 Days</div>
        <div className="grid grid-cols-10 gap-1">
          {streakData.map((day, index) => (
            <div
              key={index}
              className={`aspect-square rounded-sm text-xs flex items-center justify-center ${
                day.hasRead
                  ? day.pagesRead > 30
                    ? 'bg-green-600 text-white'
                    : day.pagesRead > 15
                    ? 'bg-green-400 text-white'
                    : 'bg-green-200 text-gray-700'
                  : 'bg-gray-100 text-gray-400'
              }`}
              title={`Day ${day.date}: ${day.pagesRead} pages`}
            >
              {day.date}
            </div>
          ))}
        </div>
        <div className="flex justify-between text-xs text-gray-500 mt-2">
          <span>Less</span>
          <div className="flex space-x-1">
            <div className="w-3 h-3 bg-gray-100 rounded-sm"></div>
            <div className="w-3 h-3 bg-green-200 rounded-sm"></div>
            <div className="w-3 h-3 bg-green-400 rounded-sm"></div>
            <div className="w-3 h-3 bg-green-600 rounded-sm"></div>
          </div>
          <span>More</span>
        </div>
      </div>
    </div>
  );
};

// Language and Author Analytics
interface ContentAnalyticsProps {
  languages: Array<{
    language: string;
    book_count: number;
    pages_read: number;
    reading_time_minutes: number;
  }>;
  authors: Array<{
    author: string;
    book_count: number;
    pages_read: number;
    favorite_book?: string;
  }>;
}

export const ContentAnalyticsChart: React.FC<ContentAnalyticsProps> = ({ 
  languages, 
  authors 
}) => {
  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {/* Languages Chart */}
      <div className="bg-white p-6 rounded-lg">
        <h4 className="text-lg font-medium text-gray-900 mb-4">Languages Read</h4>
        <ResponsiveContainer width="100%" height={250}>
          <BarChart data={languages}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="language" />
            <YAxis />
            <Tooltip />
            <Bar dataKey="pages_read" fill="#8b5cf6" name="Pages Read" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      {/* Authors Chart */}
      <div className="bg-white p-6 rounded-lg">
        <h4 className="text-lg font-medium text-gray-900 mb-4">Favorite Authors</h4>
        <ResponsiveContainer width="100%" height={250}>
          <BarChart data={authors.slice(0, 6)} layout="horizontal">
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis type="number" />
            <YAxis dataKey="author" type="category" width={100} />
            <Tooltip />
            <Bar dataKey="pages_read" fill="#10b981" name="Pages Read" />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
};

// Annotation Engagement Chart
interface AnnotationEngagementProps {
  totalAnnotations: number;
  highlights: number;
  notes: number;
  bookmarks: number;
  mostAnnotatedBooks: Array<{
    title: string;
    author: string;
    annotation_count: number;
  }>;
}

export const AnnotationEngagementChart: React.FC<AnnotationEngagementProps> = ({
  totalAnnotations,
  highlights,
  notes,
  bookmarks,
  mostAnnotatedBooks
}) => {
  const annotationData = [
    { name: 'Highlights', value: highlights, color: '#fbbf24' },
    { name: 'Notes', value: notes, color: '#10b981' },
    { name: 'Bookmarks', value: bookmarks, color: '#8b5cf6' },
  ];

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {/* Annotation Types Distribution */}
      <div className="bg-white p-6 rounded-lg">
        <h4 className="text-lg font-medium text-gray-900 mb-4">Annotation Types</h4>
        <ResponsiveContainer width="100%" height={250}>
          <PieChart>
            <Pie
              data={annotationData}
              cx="50%"
              cy="50%"
              outerRadius={80}
              fill="#8884d8"
              dataKey="value"
              label={({ name, value }) => `${name}: ${value}`}
            >
              {annotationData.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={entry.color} />
              ))}
            </Pie>
            <Tooltip />
          </PieChart>
        </ResponsiveContainer>
      </div>

      {/* Most Annotated Books */}
      <div className="bg-white p-6 rounded-lg">
        <h4 className="text-lg font-medium text-gray-900 mb-4">Most Annotated Books</h4>
        <ResponsiveContainer width="100%" height={250}>
          <BarChart data={mostAnnotatedBooks.slice(0, 5)}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="title" angle={-45} textAnchor="end" height={100} />
            <YAxis />
            <Tooltip 
              formatter={(value, name) => [value, 'Annotations']}
              labelFormatter={(title) => {
                const book = mostAnnotatedBooks.find(b => b.title === title);
                return book ? `${book.title} by ${book.author}` : title;
              }}
            />
            <Bar dataKey="annotation_count" fill="#6366f1" />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
};