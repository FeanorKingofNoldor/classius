# Book Reader Components

This directory contains comprehensive book reader components for the Classius reading platform.

## Components

### BookReader (Main Component)
The unified reader component that handles all file types and provides a complete reading experience.

**Features:**
- ✅ Supports EPUB, PDF, and text files
- ✅ Fullscreen reading mode
- ✅ Reading progress synchronization
- ✅ Annotation creation (highlights, notes, bookmarks)
- ✅ Customizable themes (light, dark, sepia)
- ✅ Font and display customization
- ✅ Keyboard shortcuts
- ✅ Settings persistence

**Usage:**
```tsx
import { BookReader } from '@/components/readers';

<BookReader
  book={book}
  fullscreen={true}
  onClose={handleClose}
  initialLocation="chapter-1" // or page number for PDFs
/>
```

### EpubReader
Specialized reader for EPUB files using epubjs library.

**Features:**
- ✅ EPUB parsing and rendering
- ✅ Table of contents navigation
- ✅ Chapter navigation
- ✅ Text selection and annotation
- ✅ Progress tracking
- ✅ Keyboard navigation

### PdfReader  
PDF viewer using react-pdf and PDF.js.

**Features:**
- ✅ PDF rendering with zoom controls
- ✅ Page navigation
- ✅ Text selection
- ✅ Annotation support
- ✅ Keyboard shortcuts
- ✅ Progress tracking

### TextReader
Simple text file reader with formatting options.

**Features:**
- ✅ Text rendering with custom styling
- ✅ Multiple reading modes (normal, column, focus)
- ✅ Text selection and annotation
- ✅ Bookmark management
- ✅ Progress tracking
- ✅ Customizable fonts and spacing

## Dependencies

Add these to your package.json:

```json
{
  "dependencies": {
    "epubjs": "^0.3.93",
    "pdfjs-dist": "^4.8.69",
    "react-pdf": "^9.1.1"
  },
  "devDependencies": {
    "@types/react-pdf": "^9.1.0"
  }
}
```

## Installation

```bash
npm install epubjs pdfjs-dist react-pdf @types/react-pdf
```

## Reading Settings

The BookReader component supports persistent reading settings:

- **Theme**: light, dark, sepia
- **Font Size**: 12px - 32px
- **Font Family**: serif, sans-serif, monospace  
- **Line Height**: 1.0 - 2.5 (text files)
- **Column Width**: 50-120 characters (text files)

Settings are automatically saved to localStorage.

## Keyboard Shortcuts

- **F11**: Toggle fullscreen
- **Ctrl+S**: Toggle settings panel
- **Alt+1**: Light theme
- **Alt+2**: Dark theme  
- **Alt+3**: Sepia theme
- **Escape**: Close reader
- **Arrow keys**: Navigate pages/content

## Progress Tracking

All readers automatically sync reading progress with the backend:

- Position/page tracking
- Reading time
- Last read timestamp
- Cross-device synchronization

## Annotation Support

Create annotations while reading:

- **Highlights**: Visual text highlighting
- **Notes**: Text with user comments
- **Bookmarks**: Position markers

Annotations are automatically saved to the backend via the annotations API.

## Integration Example

```tsx
// In your book detail page
import { BookReader } from '@/components/readers';

function BookDetailPage() {
  const [showReader, setShowReader] = useState(false);
  const [book, setBook] = useState<Book>();

  if (showReader) {
    return (
      <BookReader
        book={book}
        fullscreen={true}
        onClose={() => setShowReader(false)}
      />
    );
  }

  return (
    <div>
      <button onClick={() => setShowReader(true)}>
        Read Book
      </button>
    </div>
  );
}
```

## Browser Compatibility

- Chrome/Chromium: Full support
- Firefox: Full support  
- Safari: Full support (with minor PDF.js limitations)
- Edge: Full support

## Performance Notes

- EPUB files are processed client-side for optimal performance
- PDF rendering is lazy-loaded per page
- Text files support large documents via virtual scrolling
- All readers implement proper cleanup on unmount