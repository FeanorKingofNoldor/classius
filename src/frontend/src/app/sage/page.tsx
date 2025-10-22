'use client';

import { useState, useEffect, useRef } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { toast } from 'react-hot-toast';

interface Message {
  id: string;
  role: 'user' | 'sage';
  content: string;
  timestamp: string;
  context?: {
    book_title?: string;
    book_author?: string;
    selected_text?: string;
  };
}

interface SageCapabilities {
  languages: string[];
  specialties: string[];
  features: string[];
}

export default function SagePage() {
  const { user, isAuthenticated } = useAuthStore();
  const router = useRouter();
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const [messages, setMessages] = useState<Message[]>([]);
  const [inputText, setInputText] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [capabilities, setCapabilities] = useState<SageCapabilities | null>(null);
  const [showWelcome, setShowWelcome] = useState(true);

  useEffect(() => {
    if (!isAuthenticated) {
      router.push('/auth/login');
      return;
    }
    fetchCapabilities();
    loadWelcomeMessage();
  }, [isAuthenticated, router]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const fetchCapabilities = async () => {
    try {
      const token = localStorage.getItem('token');
      if (!token) return;

      const response = await fetch('http://localhost:8080/api/sage/capabilities', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const result = await response.json();
        setCapabilities(result.data);
      }
    } catch (err) {
      console.error('Error fetching capabilities:', err);
    }
  };

  const loadWelcomeMessage = () => {
    const welcomeMessage: Message = {
      id: 'welcome',
      role: 'sage',
      content: `Welcome to the AI Sage! I'm here to help you understand classical literature, philosophy, and historical texts.

I can assist you with:
• Explaining difficult passages and concepts
• Providing historical and cultural context  
• Analyzing philosophical arguments
• Translating Latin, Greek, and other classical languages
• Connecting ideas across different texts and authors
• Discussing the relevance of classical ideas today

Feel free to ask me anything about your readings, or start with one of the suggested questions below!`,
      timestamp: new Date().toISOString(),
    };
    setMessages([welcomeMessage]);
  };

  const sendMessage = async () => {
    if (!inputText.trim() || isLoading) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: inputText,
      timestamp: new Date().toISOString(),
    };

    setMessages(prev => [...prev, userMessage]);
    setInputText('');
    setIsLoading(true);
    setShowWelcome(false);

    try {
      const token = localStorage.getItem('token');
      if (!token) throw new Error('No authentication token');

      const response = await fetch('http://localhost:8080/api/sage/ask', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          question: inputText,
          context: 'general',
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to get response from AI Sage');
      }

      const result = await response.json();
      
      const sageMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'sage',
        content: result.data.answer,
        timestamp: new Date().toISOString(),
      };

      setMessages(prev => [...prev, sageMessage]);
    } catch (err) {
      console.error('Error:', err);
      toast.error(err instanceof Error ? err.message : 'Failed to get AI Sage response');
      
      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'sage',
        content: 'I apologize, but I\'m having trouble processing your request right now. Please try again in a moment.',
        timestamp: new Date().toISOString(),
      };
      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  const suggestedQuestions = [
    "What are the main themes in Plato's Republic?",
    "Explain the concept of virtue ethics in Aristotle",
    "What is the historical context of Marcus Aurelius' Meditations?",
    "How does Stoicism differ from Epicureanism?",
    "What are the key ideas in Augustine's Confessions?",
    "Explain the significance of the Iliad in ancient Greek culture",
  ];

  const clearConversation = () => {
    setMessages([]);
    loadWelcomeMessage();
    setShowWelcome(true);
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleTimeString([], { 
      hour: '2-digit', 
      minute: '2-digit' 
    });
  };

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-4">
              <Link href="/dashboard" className="text-indigo-600 hover:text-indigo-800">
                ← Back to Dashboard
              </Link>
              <div className="border-l border-gray-300 pl-4">
                <h1 className="text-xl font-semibold text-gray-900 flex items-center">
                  <span className="text-2xl mr-2">🤖</span>
                  AI Sage
                </h1>
                <p className="text-sm text-gray-600">Classical Education Assistant</p>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <button
                onClick={clearConversation}
                className="px-3 py-2 text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-md"
              >
                New Conversation
              </button>
              <span className="text-gray-700">Welcome, {user?.username || user?.email}</span>
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-4xl mx-auto h-[calc(100vh-4rem)] flex flex-col">
        {/* Messages Container */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {messages.map((message) => (
            <div
              key={message.id}
              className={`flex ${message.role === 'user' ? 'justify-end' : 'justify-start'}`}
            >
              <div
                className={`max-w-3xl px-4 py-3 rounded-lg ${
                  message.role === 'user'
                    ? 'bg-indigo-600 text-white'
                    : 'bg-white border border-gray-200 shadow-sm'
                }`}
              >
                {message.role === 'sage' && (
                  <div className="flex items-center mb-2">
                    <span className="text-xl mr-2">🤖</span>
                    <span className="font-medium text-indigo-600">AI Sage</span>
                    <span className="text-xs text-gray-500 ml-2">
                      {formatTimestamp(message.timestamp)}
                    </span>
                  </div>
                )}
                
                <div className={`prose ${message.role === 'user' ? 'prose-invert' : ''} max-w-none`}>
                  <div className="whitespace-pre-wrap text-sm leading-relaxed">
                    {message.content}
                  </div>
                </div>

                {message.context && (
                  <div className="mt-3 pt-3 border-t border-gray-200 text-xs text-gray-500">
                    <div className="flex items-center space-x-2">
                      <span>📖</span>
                      <span>
                        {message.context.book_title} by {message.context.book_author}
                      </span>
                    </div>
                    {message.context.selected_text && (
                      <div className="mt-1 italic">
                        "{message.context.selected_text.substring(0, 100)}..."
                      </div>
                    )}
                  </div>
                )}

                {message.role === 'user' && (
                  <div className="text-right mt-2">
                    <span className="text-xs text-indigo-200">
                      {formatTimestamp(message.timestamp)}
                    </span>
                  </div>
                )}
              </div>
            </div>
          ))}

          {isLoading && (
            <div className="flex justify-start">
              <div className="bg-white border border-gray-200 shadow-sm rounded-lg px-4 py-3 max-w-3xl">
                <div className="flex items-center mb-2">
                  <span className="text-xl mr-2">🤖</span>
                  <span className="font-medium text-indigo-600">AI Sage</span>
                  <span className="text-xs text-gray-500 ml-2">thinking...</span>
                </div>
                <div className="flex space-x-1">
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{animationDelay: '0.1s'}}></div>
                  <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{animationDelay: '0.2s'}}></div>
                </div>
              </div>
            </div>
          )}

          {/* Suggested Questions */}
          {showWelcome && messages.length === 1 && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mt-6">
              <div className="col-span-full">
                <h3 className="text-lg font-medium text-gray-900 mb-4">
                  Suggested Questions to Get Started:
                </h3>
              </div>
              {suggestedQuestions.map((question, index) => (
                <button
                  key={index}
                  onClick={() => {
                    setInputText(question);
                    setShowWelcome(false);
                  }}
                  className="text-left p-4 bg-white border border-gray-200 rounded-lg hover:border-indigo-300 hover:shadow-sm transition-all"
                >
                  <div className="flex items-start">
                    <span className="text-indigo-600 mr-2">💡</span>
                    <span className="text-sm text-gray-700">{question}</span>
                  </div>
                </button>
              ))}
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        {/* Input Area */}
        <div className="border-t border-gray-200 bg-white p-4">
          <div className="flex space-x-4">
            <div className="flex-1">
              <textarea
                value={inputText}
                onChange={(e) => setInputText(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder="Ask me anything about classical literature, philosophy, or history..."
                className="w-full px-4 py-3 border border-gray-300 rounded-lg resize-none focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                rows={3}
                disabled={isLoading}
              />
              <div className="flex justify-between items-center mt-2">
                <div className="text-xs text-gray-500">
                  Press Enter to send, Shift+Enter for new line
                </div>
                <div className="text-xs text-gray-500">
                  {inputText.length} characters
                </div>
              </div>
            </div>
            <div className="flex flex-col justify-end">
              <button
                onClick={sendMessage}
                disabled={!inputText.trim() || isLoading}
                className="px-6 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed flex items-center"
              >
                {isLoading ? (
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                ) : (
                  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M10.894 2.553a1 1 0 00-1.788 0l-7 14a1 1 0 001.169 1.409l5-1.429A1 1 0 009 15.571V11a1 1 0 112 0v4.571a1 1 0 00.725.962l5 1.428a1 1 0 001.17-1.408l-7-14z" />
                  </svg>
                )}
              </button>
            </div>
          </div>
        </div>

        {/* Capabilities Info */}
        {capabilities && (
          <div className="bg-gray-50 border-t border-gray-200 p-4">
            <details className="group">
              <summary className="flex justify-between items-center cursor-pointer text-sm text-gray-600 hover:text-gray-900">
                <span>ℹ️ AI Sage Capabilities</span>
                <svg className="w-4 h-4 group-open:rotate-180 transition-transform" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clipRule="evenodd" />
                </svg>
              </summary>
              <div className="mt-3 grid grid-cols-1 md:grid-cols-3 gap-4 text-xs">
                {capabilities.languages && (
                  <div>
                    <h4 className="font-medium text-gray-700 mb-1">Languages</h4>
                    <p className="text-gray-500">{capabilities.languages.join(', ')}</p>
                  </div>
                )}
                {capabilities.specialties && (
                  <div>
                    <h4 className="font-medium text-gray-700 mb-1">Specialties</h4>
                    <p className="text-gray-500">{capabilities.specialties.join(', ')}</p>
                  </div>
                )}
                {capabilities.features && (
                  <div>
                    <h4 className="font-medium text-gray-700 mb-1">Features</h4>
                    <p className="text-gray-500">{capabilities.features.join(', ')}</p>
                  </div>
                )}
              </div>
            </details>
          </div>
        )}
      </div>
    </div>
  );
}