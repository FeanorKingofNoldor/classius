#ifndef BOOKENGINE_H
#define BOOKENGINE_H

#include <QObject>
#include <QString>
#include <QStringList>
#include <QUrl>
#include <QFont>
#include <QSize>
#include <QSqlDatabase>
#include <QNetworkAccessManager>
#include <QThread>

class EpubParser;
class PdfParser;
class TextPaginator;
class AnnotationManager;

/**
 * @brief Structure representing reading progress for a book
 */
struct BookProgress {
    QString bookId;
    int currentPage;
    int totalPages;
    int position;        // Character position on current page
    float percentage;    // Overall progress percentage
    qint64 lastRead;     // Timestamp of last reading session
    int timeSpent;       // Total reading time in minutes
};

/**
 * @brief Structure representing book metadata
 */
struct BookMetadata {
    QString id;
    QString title;
    QString author;
    QString language;
    QString publisher;
    QString isbn;
    QString description;
    QString coverPath;
    QString filePath;
    QString format;      // EPUB, PDF, TXT, etc.
    int pageCount;
    qint64 fileSize;
    qint64 addedDate;
};

/**
 * @brief Core engine for book management, parsing, and reading functionality
 * 
 * The BookEngine handles all book-related operations including:
 * - Loading and parsing different book formats (EPUB, PDF, TXT)
 * - Text pagination optimized for e-ink displays
 * - Reading progress tracking and synchronization
 * - Book metadata management
 * - Integration with server API for sync
 */
class BookEngine : public QObject
{
    Q_OBJECT
    
    Q_PROPERTY(QString currentBookId READ currentBookId NOTIFY currentBookChanged)
    Q_PROPERTY(QString currentTitle READ currentTitle NOTIFY currentBookChanged)
    Q_PROPERTY(int currentPage READ currentPage WRITE setCurrentPage NOTIFY pageChanged)
    Q_PROPERTY(int totalPages READ totalPages NOTIFY bookLoaded)
    Q_PROPERTY(float progress READ progress NOTIFY progressChanged)
    Q_PROPERTY(QString currentText READ currentText NOTIFY textChanged)
    Q_PROPERTY(bool isLoading READ isLoading NOTIFY loadingChanged)

public:
    explicit BookEngine(QObject *parent = nullptr);
    ~BookEngine();

    // Book loading and management
    Q_INVOKABLE bool loadBook(const QString &bookId);
    Q_INVOKABLE bool loadBookFromFile(const QString &filePath);
    Q_INVOKABLE QString addBookToLibrary(const QString &filePath);
    Q_INVOKABLE QStringList getLibraryBooks() const;
    Q_INVOKABLE BookMetadata getBookMetadata(const QString &bookId) const;
    
    // Navigation
    Q_INVOKABLE void turnPage(int direction = 1);
    Q_INVOKABLE void goToPage(int pageNumber);
    Q_INVOKABLE void goToPosition(int position);
    Q_INVOKABLE void goToChapter(int chapterIndex);
    
    // Search functionality
    Q_INVOKABLE QStringList searchInBook(const QString &query) const;
    Q_INVOKABLE int findTextPosition(const QString &text) const;
    
    // Progress and bookmarks
    Q_INVOKABLE void saveProgress();
    Q_INVOKABLE BookProgress getProgress(const QString &bookId) const;
    Q_INVOKABLE void addBookmark(const QString &name = QString());
    Q_INVOKABLE QStringList getBookmarks() const;
    
    // Format support
    Q_INVOKABLE bool isFormatSupported(const QString &format) const;
    Q_INVOKABLE QString convertFormat(const QString &inputPath, const QString &outputFormat);
    
    // Display settings
    Q_INVOKABLE void setFont(const QFont &font);
    Q_INVOKABLE void setPageSize(const QSize &size);
    Q_INVOKABLE void setMargins(int left, int top, int right, int bottom);
    Q_INVOKABLE void setLineSpacing(float spacing);
    
    // Getters
    QString currentBookId() const { return m_currentBookId; }
    QString currentTitle() const { return m_currentTitle; }
    int currentPage() const { return m_currentPage; }
    int totalPages() const { return m_totalPages; }
    float progress() const { return m_progress; }
    QString currentText() const { return m_currentPageText; }
    bool isLoading() const { return m_isLoading; }
    
    // Setters
    void setCurrentPage(int page);

signals:
    void currentBookChanged();
    void bookLoaded(const QString &bookId);
    void pageChanged(int page);
    void progressChanged(float progress);
    void textChanged();
    void loadingChanged(bool loading);
    void error(const QString &message);
    void bookAdded(const QString &bookId);

private slots:
    void onPaginationComplete();
    void onSyncComplete();
    void onError(const QString &error);

private:
    // Internal methods
    bool initializeDatabase();
    void setupParsers();
    void paginateCurrentBook();
    void updateProgress();
    void syncWithServer();
    QString extractText();
    QString generateBookId(const QString &filePath);
    void saveBookMetadata(const BookMetadata &metadata);
    BookMetadata loadBookMetadata(const QString &bookId);
    
    // Member variables
    QString m_currentBookId;
    QString m_currentTitle;
    QString m_currentFilePath;
    int m_currentPage;
    int m_totalPages;
    float m_progress;
    QString m_currentPageText;
    bool m_isLoading;
    
    // Display settings
    QFont m_font;
    QSize m_pageSize;
    QMargins m_margins;
    float m_lineSpacing;
    
    // Components
    EpubParser *m_epubParser;
    PdfParser *m_pdfParser;
    TextPaginator *m_paginator;
    AnnotationManager *m_annotationManager;
    QNetworkAccessManager *m_networkManager;
    
    // Database
    QSqlDatabase m_database;
    QString m_databasePath;
    
    // Cache
    QStringList m_currentBookPages;
    QHash<QString, BookMetadata> m_metadataCache;
    
    // Threading
    QThread *m_paginationThread;
};

#endif // BOOKENGINE_H