# CLASSIUS TECHNICAL ARCHITECTURE
## How We Build It: Complete Implementation Plan

---

## SYSTEM OVERVIEW

### High-Level Architecture
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLASSIUS      │    │   HOME SERVER   │    │   EXTERNAL      │
│   DEVICE        │◄──►│   BACKEND       │◄──►│   APIS          │
│                 │    │                 │    │                 │
│ • Qt/QML UI     │    │ • Go API        │    │ • OpenAI GPT-4  │
│ • C++ Drivers   │    │ • PostgreSQL    │    │ • Midjourney    │
│ • Python AI     │    │ • Redis Cache   │    │ • Stripe        │
│ • Audio DSP     │    │ • WebRTC        │    │ • CloudFlare    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

---

## DEVELOPMENT STACK

### Device Side (Modified reMarkable OS)
- **OS Base:** Yocto Linux with custom kernel modules
- **UI Framework:** Qt 6.x with QML for responsive layouts
- **System Language:** C++ for performance-critical components
- **Scripting:** Python 3.11+ for AI integration and text processing
- **Audio:** PortAudio + custom DSP libraries for whistle detection
- **Graphics:** OpenGL ES optimized for e-ink refresh rates

### Server Side (Your Home Server)
- **API Backend:** Go 1.21+ with Gin web framework
- **Database:** PostgreSQL 15+ with full-text search extensions
- **Cache:** Redis 7+ for session management and AI response caching
- **Real-time:** WebSocket connections for live features
- **AI Processing:** Python FastAPI microservices
- **File Storage:** MinIO for book files and user content

---

## PHASE 1: MVP IMPLEMENTATION

### Step 1: Development Environment Setup

**Local Dev Environment:**
```bash
# Repository structure
classius/
├── device/           # Device-side code
│   ├── ui/          # Qt/QML interfaces
│   ├── drivers/     # Hardware abstraction
│   ├── audio/       # Whistle processing
│   └── ai/          # Sage integration
├── server/          # Backend services
│   ├── api/         # Go REST API
│   ├── db/          # Database schemas
│   ├── ai/          # Python AI services
│   └── docker/      # Container configs
├── shared/          # Common protocols
└── docs/           # Technical documentation

# Development tools
make install-dev     # Install Qt, Go, Python, PostgreSQL
make setup-device    # Flash development firmware
make run-server      # Start local backend
make run-device      # Launch device simulator
```

**Cross-Compilation Pipeline:**
```bash
# Device builds (ARM target)
export CROSS_COMPILE=arm-linux-gnueabihf-
make build-device TARGET=arm64
make flash-device    # Deploy to actual hardware

# Server builds (x86_64 for your home server)  
make build-server TARGET=linux-amd64
make deploy-server HOST=your-home-server.local
```

---

### Step 2: Core Reading System (Block 1)

**Qt/QML Reading Interface:**
```cpp
// device/ui/ReadingView.h
class ReadingView : public QQuickItem {
    Q_OBJECT
    Q_PROPERTY(QString bookTitle READ bookTitle NOTIFY bookChanged)
    Q_PROPERTY(int currentPage READ currentPage WRITE setCurrentPage)
    Q_PROPERTY(QString currentText READ currentText NOTIFY textChanged)

public:
    explicit ReadingView(QQuickItem *parent = nullptr);
    
    // Book management
    Q_INVOKABLE void openBook(const QString &bookId);
    Q_INVOKABLE void turnPage(int direction);
    Q_INVOKABLE void addHighlight(int startPos, int endPos, const QString &color);
    
    // E-ink optimization
    void setRefreshMode(EInkRefreshMode mode);
    void optimizeForEInk();

private:
    BookEngine *m_bookEngine;
    TextRenderer *m_textRenderer;
    AnnotationManager *m_annotations;
};
```

**QML User Interface:**
```qml
// device/ui/ReadingPage.qml
import QtQuick 2.15
import QtQuick.Controls 2.15
import Classius 1.0

Page {
    id: readingPage
    
    ReadingView {
        id: bookView
        anchors.fill: parent
        
        onTextChanged: {
            // Update progress, save position
            progressBar.value = bookView.progress
            saveReadingPosition()
        }
        
        MouseArea {
            anchors.fill: parent
            onClicked: {
                if (mouse.x < width/3) bookView.turnPage(-1)
                else if (mouse.x > 2*width/3) bookView.turnPage(1)
                else toggleUI()
            }
        }
    }
    
    // Floating UI elements
    ProgressBar {
        id: progressBar
        anchors.bottom: parent.bottom
        width: parent.width
        visible: showUI
    }
    
    SageButton {
        id: sageButton
        anchors.top: parent.top
        anchors.right: parent.right
        onClicked: activateSage()
    }
}
```

**Book Processing Engine:**
```cpp
// device/core/BookEngine.cpp
class BookEngine : public QObject {
public:
    // Format support
    bool loadEpub(const QString &filePath);
    bool loadPdf(const QString &filePath);
    bool convertToInternalFormat(const QString &input);
    
    // Text rendering
    QStringList paginateText(const QString &text, const QFont &font, const QSize &pageSize);
    QString getPageText(int pageNumber);
    
    // Progress tracking
    void saveProgress(const QString &bookId, int page, int position);
    BookProgress loadProgress(const QString &bookId);
    
private:
    EpubParser *m_epubParser;
    PdfParser *m_pdfParser;
    TextPaginator *m_paginator;
    QSqlDatabase m_database;
};
```

---

### Step 3: Backend API (Go Server)

**Main API Server:**
```go
// server/api/main.go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/classius/server/handlers"
    "github.com/classius/server/middleware"
    "github.com/classius/server/database"
)

func main() {
    // Database connection
    db := database.Connect()
    defer db.Close()
    
    // Router setup
    r := gin.Default()
    r.Use(middleware.CORS())
    r.Use(middleware.Authentication())
    
    // API routes
    api := r.Group("/api/v1")
    {
        // Books
        api.GET("/books", handlers.GetBooks)
        api.POST("/books/upload", handlers.UploadBook)
        api.GET("/books/:id", handlers.GetBook)
        api.GET("/books/:id/content", handlers.GetBookContent)
        
        // User data
        api.GET("/user/profile", handlers.GetUserProfile)
        api.POST("/user/progress", handlers.SaveProgress)
        api.GET("/user/annotations", handlers.GetAnnotations)
        
        // AI Sage
        api.POST("/sage/question", handlers.AskSage)
        api.GET("/sage/history", handlers.GetSageHistory)
    }
    
    r.Run(":8080")
}
```

**Book Management Handlers:**
```go
// server/handlers/books.go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/classius/server/models"
    "github.com/classius/server/services"
)

func GetBooks(c *gin.Context) {
    userID := c.GetString("user_id")
    
    books, err := services.GetUserLibrary(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"books": books})
}

func UploadBook(c *gin.Context) {
    file, header, err := c.Request.FormFile("book")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
        return
    }
    defer file.Close()
    
    // Process book (convert format, extract metadata)
    book, err := services.ProcessUploadedBook(file, header.Filename)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"book": book})
}
```

**Database Models:**
```go
// server/models/book.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type Book struct {
    ID          uint      `json:"id" gorm:"primarykey"`
    Title       string    `json:"title" gorm:"not null"`
    Author      string    `json:"author"`
    ISBN        string    `json:"isbn"`
    Language    string    `json:"language"`
    FilePath    string    `json:"-"` // Not exposed to API
    FileFormat  string    `json:"format"`
    PageCount   int       `json:"page_count"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type UserProgress struct {
    ID         uint      `json:"id" gorm:"primarykey"`
    UserID     uint      `json:"user_id" gorm:"not null"`
    BookID     uint      `json:"book_id" gorm:"not null"`
    CurrentPage int      `json:"current_page"`
    Position   int       `json:"position"`
    Progress   float32   `json:"progress"` // Percentage
    LastRead   time.Time `json:"last_read"`
}

type Annotation struct {
    ID         uint      `json:"id" gorm:"primarykey"`
    UserID     uint      `json:"user_id" gorm:"not null"`
    BookID     uint      `json:"book_id" gorm:"not null"`
    Page       int       `json:"page"`
    StartPos   int       `json:"start_pos"`
    EndPos     int       `json:"end_pos"`
    Type       string    `json:"type"` // highlight, note, bookmark
    Content    string    `json:"content"`
    Color      string    `json:"color,omitempty"`
    CreatedAt  time.Time `json:"created_at"`
}
```

---

### Step 4: AI Sage Integration

**Python AI Service:**
```python
# server/ai/sage_service.py
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import openai
import json
from typing import Optional

app = FastAPI()

class SageRequest(BaseModel):
    question: str
    book_id: Optional[str] = None
    current_passage: Optional[str] = None
    user_context: Optional[dict] = None

class SageResponse(BaseModel):
    answer: str
    confidence: float
    related_passages: list[str] = []
    follow_up_questions: list[str] = []

@app.post("/ask", response_model=SageResponse)
async def ask_sage(request: SageRequest):
    try:
        # Build context for GPT-4
        context = build_context(request)
        
        # Make API call to OpenAI
        response = await openai.ChatCompletion.acreate(
            model="gpt-4",
            messages=[
                {"role": "system", "content": SAGE_SYSTEM_PROMPT},
                {"role": "user", "content": context}
            ],
            max_tokens=500,
            temperature=0.7
        )
        
        answer = response.choices[0].message.content
        
        # Parse and enhance response
        return SageResponse(
            answer=answer,
            confidence=calculate_confidence(answer),
            related_passages=find_related_passages(request.book_id, request.question),
            follow_up_questions=generate_follow_ups(answer)
        )
        
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

def build_context(request: SageRequest) -> str:
    """Build comprehensive context for AI"""
    context_parts = [f"Question: {request.question}"]
    
    if request.current_passage:
        context_parts.append(f"Current passage: {request.current_passage}")
    
    if request.user_context:
        reading_history = request.user_context.get("reading_history", [])
        if reading_history:
            context_parts.append(f"User has read: {', '.join(reading_history[-5:])}")
    
    return "\n\n".join(context_parts)

SAGE_SYSTEM_PROMPT = """
You are The Sage, an AI tutor for classical education. Your role is to:

1. Explain difficult passages in clear, accessible language
2. Provide historical and philosophical context
3. Make connections to other texts the user has read
4. Ask Socratic questions to guide thinking
5. Encourage deeper exploration of ideas

Always:
- Be encouraging and supportive
- Admit when you're uncertain
- Suggest further reading when appropriate
- Use examples to clarify abstract concepts
- Maintain academic rigor while being approachable

Format responses for e-ink display (avoid complex formatting).
"""
```

**Device-Side Sage Integration:**
```cpp
// device/ai/SageClient.cpp
class SageClient : public QObject {
    Q_OBJECT

public:
    Q_INVOKABLE void askQuestion(const QString &question);
    Q_INVOKABLE void explainPassage(const QString &passage);

signals:
    void responseReceived(const QString &answer);
    void errorOccurred(const QString &error);

private slots:
    void handleNetworkResponse();

private:
    QNetworkAccessManager *m_networkManager;
    QString m_serverUrl;
    QString m_currentBookId;
    
    void sendRequest(const QJsonObject &request);
    QString buildContext();
};

void SageClient::askQuestion(const QString &question) {
    QJsonObject request;
    request["question"] = question;
    request["book_id"] = m_currentBookId;
    request["current_passage"] = getCurrentPassage();
    request["user_context"] = buildUserContext();
    
    sendRequest(request);
}
```

---

### Step 5: Whistle Control System

**Audio Processing (C++):**
```cpp
// device/audio/WhistleDetector.h
class WhistleDetector : public QObject {
    Q_OBJECT

public:
    explicit WhistleDetector(QObject *parent = nullptr);
    
    void startListening();
    void stopListening();
    void trainUserWhistle(const QString &patternName);
    
    struct WhistlePattern {
        QString name;
        std::vector<float> frequencies;
        std::vector<float> timings;
        float confidence_threshold;
    };

signals:
    void whistleDetected(const QString &command);
    void trainingComplete(bool success);

private slots:
    void processAudioBuffer();

private:
    QAudioInput *m_audioInput;
    QIODevice *m_audioDevice;
    
    // DSP components
    FFTProcessor *m_fftProcessor;
    PatternMatcher *m_patternMatcher;
    NoiseGate *m_noiseGate;
    
    // User patterns
    QVector<WhistlePattern> m_userPatterns;
    
    bool isWhistleFrequency(float frequency);
    float calculatePatternMatch(const std::vector<float> &detected, 
                               const WhistlePattern &pattern);
    void saveUserPatterns();
    void loadUserPatterns();
};
```

**FFT Audio Analysis:**
```cpp
// device/audio/FFTProcessor.cpp
class FFTProcessor {
public:
    FFTProcessor(int sampleRate, int bufferSize);
    
    struct FrequencyData {
        std::vector<float> frequencies;
        std::vector<float> magnitudes;
        float dominantFreq;
        float confidence;
    };
    
    FrequencyData processBuffer(const float* audioBuffer, int samples);
    
private:
    kiss_fft_cfg m_fftConfig;
    std::vector<kiss_fft_cpx> m_fftInput;
    std::vector<kiss_fft_cpx> m_fftOutput;
    
    int m_sampleRate;
    int m_bufferSize;
    
    void applyWindow(float* buffer, int size);
    float findDominantFrequency(const std::vector<float>& magnitudes);
};

FrequencyData FFTProcessor::processBuffer(const float* audioBuffer, int samples) {
    // Apply Hamming window to reduce spectral leakage
    std::vector<float> windowed(samples);
    std::copy(audioBuffer, audioBuffer + samples, windowed.begin());
    applyWindow(windowed.data(), samples);
    
    // Convert to complex numbers for FFT
    for (int i = 0; i < samples; ++i) {
        m_fftInput[i].r = windowed[i];
        m_fftInput[i].i = 0.0f;
    }
    
    // Perform FFT
    kiss_fft(m_fftConfig, m_fftInput.data(), m_fftOutput.data());
    
    // Extract magnitude spectrum
    FrequencyData result;
    result.frequencies.resize(samples / 2);
    result.magnitudes.resize(samples / 2);
    
    for (int i = 0; i < samples / 2; ++i) {
        float real = m_fftOutput[i].r;
        float imag = m_fftOutput[i].i;
        result.magnitudes[i] = sqrt(real * real + imag * imag);
        result.frequencies[i] = (float)i * m_sampleRate / samples;
    }
    
    result.dominantFreq = findDominantFrequency(result.magnitudes);
    result.confidence = calculateConfidence(result.magnitudes);
    
    return result;
}
```

---

## DEVELOPMENT WORKFLOW

### Daily Development Process

**1. Code Organization:**
```bash
# Start development session
make dev-setup        # Start databases, services
make device-sim        # Launch device simulator
make hot-reload        # Enable live code updates

# Work on specific components
cd device/ui && qmake && make  # UI changes
cd server/api && go run main.go # Backend changes  
cd server/ai && python -m uvicorn sage_service:app --reload # AI updates

# Testing
make test-device       # Run device tests
make test-server       # Run server tests
make test-integration  # End-to-end tests
```

**2. Git Workflow:**
```bash
# Feature branches
git checkout -b feature/whistle-detection
git commit -m "Add FFT-based whistle processing"
git push origin feature/whistle-detection

# Code review process
hub pull-request -b main -h feature/whistle-detection
# Review, approve, merge

# Deployment
git checkout main
make build-release
make deploy-staging    # Test on staging device
make deploy-production # Push to production devices
```

**3. Testing Strategy:**
```bash
# Unit tests
device/tests/test_whistle_detector.cpp
server/tests/handlers_test.go
server/ai/tests/test_sage.py

# Integration tests  
tests/integration/test_device_server_sync.py
tests/integration/test_sage_conversation.py

# Hardware tests (on actual device)
tests/hardware/test_eink_refresh.cpp
tests/hardware/test_stylus_input.cpp
tests/hardware/test_audio_processing.cpp
```

---

## DEPLOYMENT PIPELINE

### Continuous Integration
```yaml
# .github/workflows/ci.yml
name: Classius CI/CD

on: [push, pull_request]

jobs:
  device-build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Qt
        uses: jurplel/install-qt-action@v3
      - name: Setup ARM toolchain
        run: sudo apt-get install gcc-arm-linux-gnueabihf
      - name: Build device firmware
        run: |
          cd device
          qmake CONFIG+=arm-cross
          make
      - name: Run device tests
        run: make test-device

  server-build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Build server
        run: |
          cd server
          go build ./...
      - name: Run server tests
        run: make test-server

  ai-service:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'
      - name: Install dependencies
        run: pip install -r server/ai/requirements.txt
      - name: Run AI tests
        run: python -m pytest server/ai/tests/
```

### Device Deployment
```bash
# Over-the-air updates
make build-firmware-update
make sign-firmware     # Cryptographically sign
make push-ota-update   # Deploy to devices

# USB deployment (development)
make flash-device USB_DEVICE=/dev/ttyUSB0

# Factory programming
make create-factory-image
make program-flash-memory
```

This architecture gives us a clear, implementable path from concept to working product. Each component can be developed and tested independently, then integrated into the full system.

Want me to dive deeper into any specific component or move on to the database schema design?