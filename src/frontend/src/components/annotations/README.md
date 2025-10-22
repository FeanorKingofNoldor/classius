# Advanced Annotation System

A comprehensive annotation system for the Classius reading platform that provides inline annotation creation, visualization, and management.

## Components

### AnnotationToolbar
An intelligent floating toolbar that appears when users select text while reading.

**Features:**
- ✅ Context-aware annotation creation (highlights, notes, bookmarks)
- ✅ Color picker with 8 predefined highlight colors
- ✅ Inline note creation with rich text support
- ✅ Tag management system
- ✅ Duplicate detection (prevents re-annotating same text)
- ✅ Keyboard shortcuts (Ctrl+Enter to save notes)
- ✅ Theme support (light, dark, sepia)
- ✅ Click-outside-to-close functionality

**Usage:**
```tsx
<AnnotationToolbar
  selectedText="The selected text"
  position={{ x: 100, y: 50 }}
  onCreateAnnotation={handleCreate}
  onClose={handleClose}
  existingAnnotations={annotations}
  theme="light"
/>
```

### AnnotationOverlay
Visual indicators that display existing annotations within the reading content.

**Features:**
- ✅ Highlight backgrounds for highlight-type annotations
- ✅ Interactive annotation indicators (icons)
- ✅ Hover tooltips with annotation previews  
- ✅ Detailed popup with full annotation content
- ✅ Inline editing capabilities
- ✅ Privacy toggle (public/private)
- ✅ Delete functionality with confirmation
- ✅ Responsive positioning

**Usage:**
```tsx
<AnnotationOverlay
  annotation={annotation}
  position={{ x: 100, y: 50, width: 200, height: 20 }}
  onUpdate={handleUpdate}
  onDelete={handleDelete}
  theme="light"
/>
```

### AnnotationPanel
Comprehensive annotation management interface.

**Features:**
- ✅ Advanced filtering (type, color, tags, date range, privacy)
- ✅ Full-text search across annotations
- ✅ Multiple sorting options (date, type, page)
- ✅ Bulk operations (delete, update tags, update color)
- ✅ Export functionality (JSON, CSV formats)
- ✅ Selection management with "select all"
- ✅ Real-time statistics display
- ✅ Navigation to annotation locations
- ✅ Responsive design with scrollable lists

**Usage:**
```tsx
<AnnotationPanel
  bookId="book-uuid"
  onNavigateToAnnotation={handleNavigate}
  theme="light"
/>
```

### useAnnotations Hook
Powerful React hook for annotation management.

**Features:**
- ✅ Complete CRUD operations
- ✅ Real-time filtering and search
- ✅ Bulk operations support
- ✅ Export functionality
- ✅ Statistics calculation
- ✅ Automatic loading and caching
- ✅ Error handling with toast notifications

**Usage:**
```tsx
const {
  annotations,
  filteredAnnotations,
  loading,
  createAnnotation,
  updateAnnotation,
  deleteAnnotation,
  setFilter,
  exportAnnotations,
  getStats
} = useAnnotations(bookId);
```

## Installation

The annotation system requires the following dependencies:

```json
{
  "dependencies": {
    "react": "^19.1.0",
    "react-dom": "^19.1.0",
    "react-hot-toast": "^2.6.0"
  }
}
```

## Integration Example

```tsx
import { EnhancedBookReader } from '@/components/readers/EnhancedBookReader';
import { Book } from '@/lib/api';

function ReadingPage({ book }: { book: Book }) {
  return (
    <EnhancedBookReader
      book={book}
      fullscreen={true}
      onClose={() => router.back()}
    />
  );
}
```

## Annotation Types

### Highlights
- Visual text highlighting with customizable colors
- 8 predefined color options (Yellow, Green, Blue, Purple, Pink, Orange, Red, Gray)
- Opacity-based overlay rendering
- Quick creation from selection toolbar

### Notes
- Rich text annotations with selected text context
- Tag support for organization
- Privacy controls (public/private)
- Inline editing capabilities
- Full-text search support

### Bookmarks
- Position markers without selected text
- Quick navigation aids
- Minimal visual footprint
- Bulk management support

## Color System

The annotation system uses a consistent color palette:

```tsx
const COLORS = [
  { name: 'Yellow', value: '#fef3c7', textColor: '#92400e' },
  { name: 'Green', value: '#d1fae5', textColor: '#065f46' },
  { name: 'Blue', value: '#dbeafe', textColor: '#1e40af' },
  // ... more colors
];
```

## Filtering & Search

### Available Filters
- **Type**: All, Highlights, Notes, Bookmarks
- **Color**: Any of the 8 predefined colors
- **Tags**: Multiple tag selection
- **Date Range**: From/to date filtering
- **Privacy**: Public, Private, or All

### Search Functionality
- Full-text search across selected text and note content
- Tag-based search
- Real-time filtering as you type
- Search result highlighting

## Bulk Operations

### Supported Operations
- **Bulk Delete**: Remove multiple annotations
- **Bulk Tag Update**: Apply tags to multiple annotations
- **Bulk Color Update**: Change highlight color for multiple items
- **Bulk Privacy Toggle**: Make annotations public/private

### Selection Management
- Select all/none functionality
- Individual selection checkboxes
- Visual indication of selected items
- Persistent selection across filtering

## Export Features

### Supported Formats
- **JSON**: Complete annotation data with metadata
- **CSV**: Spreadsheet-friendly format for analysis
- **PDF**: Formatted report (planned)

### Export Data Includes
- Annotation content and selected text
- Type, color, and tag information
- Creation and modification timestamps
- Privacy settings and user metadata
- Book context and location information

## Performance Considerations

- **Lazy Loading**: Annotations loaded on demand
- **Virtual Scrolling**: Handles large annotation lists efficiently
- **Debounced Search**: Prevents excessive API calls
- **Optimistic Updates**: Immediate UI feedback for user actions
- **Caching**: Hook-based caching for repeated requests

## Accessibility Features

- **Keyboard Navigation**: Full keyboard support for all interactions
- **Screen Reader Support**: ARIA labels and semantic HTML
- **High Contrast**: Theme support for accessibility
- **Focus Management**: Proper focus handling for popups and modals
- **Tooltip Delays**: Appropriate timing for hover interactions

## Keyboard Shortcuts

### Global Shortcuts
- **Escape**: Close active popups/toolbars
- **Ctrl+S**: Toggle settings/annotation panel
- **Tab/Shift+Tab**: Navigate through interactive elements

### Annotation Toolbar
- **Ctrl+Enter**: Quick save for notes
- **Escape**: Close toolbar
- **Enter**: Add tags

### Annotation Panel
- **Ctrl+A**: Select all annotations
- **Delete**: Delete selected annotations
- **Ctrl+F**: Focus search input

## Theme Support

The annotation system supports three themes:

```tsx
type Theme = 'light' | dark' | 'sepia';

// Theme classes automatically applied
const themes = {
  light: 'bg-white border-gray-300 text-gray-900',
  dark: 'bg-gray-800 border-gray-600 text-gray-100',
  sepia: 'bg-yellow-50 border-yellow-300 text-yellow-900',
};
```

## Best Practices

### Component Integration
- Always provide `onClose` handlers for modals/popups
- Use proper loading states during async operations
- Implement error boundaries for annotation components
- Cache annotation data at the appropriate level

### User Experience
- Provide visual feedback for all user actions
- Use toast notifications for success/error states
- Implement proper loading indicators
- Maintain scroll position during operations

### Performance
- Debounce search and filter operations
- Use React.memo for expensive list renders
- Implement proper cleanup in useEffect hooks
- Avoid unnecessary re-renders with useMemo/useCallback

## Browser Compatibility

- **Chrome/Chromium**: Full support
- **Firefox**: Full support
- **Safari**: Full support (minor text selection limitations)
- **Edge**: Full support

## Future Enhancements

Planned features for future releases:

- **Collaborative Annotations**: Real-time shared annotations
- **AI-Powered Suggestions**: Smart annotation recommendations
- **Annotation Analytics**: Usage patterns and insights
- **Advanced Export Options**: PDF with highlights, EPUB annotations
- **Annotation Threads**: Discussion support for shared annotations
- **Voice Annotations**: Audio note recording
- **OCR Integration**: Annotations for scanned texts
- **Annotation Sync**: Cross-device synchronization