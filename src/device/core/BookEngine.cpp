#include "BookEngine.h"
#include "TextPaginator.h"
#include "AnnotationManager.h"

#include <QDebug>
#include <QDir>
#include <QFile>
#include <QFileInfo>
#include <QStandardPaths>
#include <QSqlQuery>
#include <QSqlError>
#include <QCryptographicHash>
#include <QJsonDocument>
#include <QJsonObject>
#include <QNetworkRequest>
#include <QNetworkReply>
#include <QTextDocument>
#include <QRegularExpression>
#include <QDateTime>

BookEngine::BookEngine(QObject *parent)
    : QObject(parent)
    , m_currentPage(0)
    , m_totalPages(0)
    , m_progress(0.0f)
    , m_isLoading(false)
    , m_font("Garamond", 12)
    , m_pageSize(600, 800)
    , m_margins(40, 40, 40, 40)
    , m_lineSpacing(1.2f)
    , m_epubParser(nullptr)
    , m_pdfParser(nullptr)
    , m_paginator(nullptr)
    , m_annotationManager(nullptr)
    , m_networkManager(nullptr)
    , m_paginationThread(nullptr)
{
    // Initialize components
    setupParsers();
    initializeDatabase();
    
    m_paginator = new TextPaginator(this);
    m_annotationManager = new AnnotationManager(this);
    m_networkManager = new QNetworkAccessManager(this);
    
    // Connect signals
    connect(m_paginator, &TextPaginator::paginationComplete, 
            this, &BookEngine::onPaginationComplete);
    
    qDebug() << "BookEngine initialized";
}

BookEngine::~BookEngine()
{
    if (m_database.isOpen()) {
        saveProgress();
        m_database.close();
    }
}

bool BookEngine::initializeDatabase()
{
    QString dataPath = QStandardPaths::writableLocation(QStandardPaths::AppDataLocation);
    QDir().mkpath(dataPath);
    m_databasePath = dataPath + "/classius.db";
    
    m_database = QSqlDatabase::addDatabase("QSQLITE", "BookEngine");
    m_database.setDatabaseName(m_databasePath);
    
    if (!m_database.open()) {
        qCritical() << "Failed to open database:" << m_database.lastError().text();
        return false;
    }
    
    // Create tables
    QSqlQuery query(m_database);
    
    // Books metadata table
    query.exec(R"(
        CREATE TABLE IF NOT EXISTS books (
            id TEXT PRIMARY KEY,
            title TEXT NOT NULL,
            author TEXT,
            language TEXT,
            publisher TEXT,
            isbn TEXT,
            description TEXT,
            cover_path TEXT,
            file_path TEXT NOT NULL,
            format TEXT NOT NULL,
            page_count INTEGER DEFAULT 0,
            file_size INTEGER,
            added_date INTEGER
        )
    )");
    
    // Reading progress table
    query.exec(R"(
        CREATE TABLE IF NOT EXISTS reading_progress (
            book_id TEXT PRIMARY KEY,
            current_page INTEGER DEFAULT 0,
            total_pages INTEGER DEFAULT 0,
            position INTEGER DEFAULT 0,
            percentage REAL DEFAULT 0.0,
            last_read INTEGER,
            time_spent INTEGER DEFAULT 0,
            FOREIGN KEY (book_id) REFERENCES books (id)
        )
    )");
    
    // Bookmarks table
    query.exec(R"(
        CREATE TABLE IF NOT EXISTS bookmarks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            book_id TEXT NOT NULL,
            name TEXT,
            page INTEGER NOT NULL,
            position INTEGER DEFAULT 0,
            created_date INTEGER,
            FOREIGN KEY (book_id) REFERENCES books (id)
        )
    )");
    
    if (query.lastError().type() != QSqlError::NoError) {
        qCritical() << "Failed to create database tables:" << query.lastError().text();
        return false;
    }
    
    qDebug() << "Database initialized at:" << m_databasePath;
    return true;
}

void BookEngine::setupParsers()
{
    // For MVP, we'll implement basic text parsing
    // TODO: Add proper EPUB and PDF parsers later
}

bool BookEngine::loadBook(const QString &bookId)
{
    if (bookId.isEmpty()) {
        emit error("Invalid book ID");
        return false;
    }
    
    m_isLoading = true;
    emit loadingChanged(true);
    
    // Load book metadata from database
    BookMetadata metadata = loadBookMetadata(bookId);
    if (metadata.id.isEmpty()) {
        m_isLoading = false;
        emit loadingChanged(false);
        emit error("Book not found: " + bookId);
        return false;
    }
    
    // Check if file still exists
    if (!QFile::exists(metadata.filePath)) {
        m_isLoading = false;
        emit loadingChanged(false);
        emit error("Book file not found: " + metadata.filePath);
        return false;
    }
    
    m_currentBookId = bookId;
    m_currentTitle = metadata.title;
    m_currentFilePath = metadata.filePath;
    
    // Extract text content
    QString content = extractText();
    if (content.isEmpty()) {
        m_isLoading = false;
        emit loadingChanged(false);
        emit error("Failed to extract text from book");
        return false;
    }
    
    // Start pagination
    m_paginator->paginateText(content, m_font, m_pageSize, m_margins, m_lineSpacing);
    
    return true;
}

bool BookEngine::loadBookFromFile(const QString &filePath)
{
    if (!QFile::exists(filePath)) {
        emit error("File not found: " + filePath);
        return false;
    }
    
    // First add to library
    QString bookId = addBookToLibrary(filePath);
    if (bookId.isEmpty()) {
        return false;
    }
    
    // Then load it
    return loadBook(bookId);
}

QString BookEngine::addBookToLibrary(const QString &filePath)
{
    QFileInfo fileInfo(filePath);
    if (!fileInfo.exists()) {
        emit error("File not found: " + filePath);
        return QString();
    }
    
    QString bookId = generateBookId(filePath);
    
    // Check if already in library
    QSqlQuery query(m_database);
    query.prepare("SELECT id FROM books WHERE id = ?");
    query.bindValue(0, bookId);
    query.exec();
    
    if (query.next()) {
        qDebug() << "Book already in library:" << bookId;
        return bookId;
    }
    
    // Extract metadata
    BookMetadata metadata;
    metadata.id = bookId;
    metadata.filePath = filePath;
    metadata.format = fileInfo.suffix().toUpper();
    metadata.fileSize = fileInfo.size();
    metadata.addedDate = QDateTime::currentSecsSinceEpoch();
    
    // Try to extract title from filename
    metadata.title = fileInfo.baseName();
    
    // For MVP, use simple heuristics for metadata
    // TODO: Implement proper metadata extraction
    if (metadata.title.contains(" - ")) {
        QStringList parts = metadata.title.split(" - ");
        if (parts.size() >= 2) {
            metadata.author = parts[0].trimmed();
            metadata.title = parts[1].trimmed();
        }
    }
    
    // Save to database
    saveBookMetadata(metadata);
    
    emit bookAdded(bookId);
    qDebug() << "Added book to library:" << metadata.title;
    
    return bookId;
}

QString BookEngine::extractText()
{
    QFile file(m_currentFilePath);
    if (!file.open(QIODevice::ReadOnly | QIODevice::Text)) {
        qCritical() << "Failed to open file:" << m_currentFilePath;
        return QString();
    }
    
    QString content;
    QFileInfo fileInfo(m_currentFilePath);
    QString format = fileInfo.suffix().toLower();
    
    if (format == "txt") {
        // Plain text file
        QTextStream stream(&file);
        // stream.setEncoding(QStringConverter::Utf8); // TODO: Fix Qt6 encoding
        content = stream.readAll();
    }
    else if (format == "epub") {
        // For MVP, treat as text (TODO: implement proper EPUB parsing)
        QTextStream stream(&file);
        content = stream.readAll();
        // Basic cleanup for EPUB-like content
        content = content.remove(QRegularExpression("<[^>]*>"));  // Remove HTML tags
    }
    else {
        // Fallback: try to read as text
        QTextStream stream(&file);
        content = stream.readAll();
    }
    
    file.close();
    
    // Basic text cleanup
    content = content.replace('\r', '\n');  // Normalize line endings
    content = content.replace(QRegularExpression("\n{3,}"), "\n\n");  // Reduce excessive newlines
    
    qDebug() << "Extracted" << content.length() << "characters from" << m_currentFilePath;
    return content;
}

void BookEngine::onPaginationComplete()
{
    m_currentBookPages = m_paginator->getPages();
    m_totalPages = m_currentBookPages.size();
    
    // Load previous progress
    BookProgress progress = getProgress(m_currentBookId);
    if (progress.bookId == m_currentBookId && progress.currentPage <= m_totalPages) {
        m_currentPage = progress.currentPage;
    } else {
        m_currentPage = 1;  // Start from first page
    }
    
    // Update display
    if (m_currentPage > 0 && m_currentPage <= m_totalPages) {
        m_currentPageText = m_currentBookPages[m_currentPage - 1];
    }
    
    updateProgress();
    
    m_isLoading = false;
    emit loadingChanged(false);
    emit bookLoaded(m_currentBookId);
    emit textChanged();
    
    qDebug() << "Book loaded successfully:" << m_currentTitle << "(" << m_totalPages << "pages)";
}

void BookEngine::turnPage(int direction)
{
    if (m_totalPages == 0) return;
    
    int newPage = m_currentPage + direction;
    if (newPage >= 1 && newPage <= m_totalPages) {
        setCurrentPage(newPage);
    }
}

void BookEngine::setCurrentPage(int page)
{
    if (page < 1 || page > m_totalPages || page == m_currentPage) {
        return;
    }
    
    m_currentPage = page;
    
    if (!m_currentBookPages.isEmpty() && page <= m_currentBookPages.size()) {
        m_currentPageText = m_currentBookPages[page - 1];
        emit textChanged();
    }
    
    updateProgress();
    emit pageChanged(m_currentPage);
    
    // Auto-save progress every page turn
    saveProgress();
}

void BookEngine::updateProgress()
{
    if (m_totalPages > 0) {
        m_progress = static_cast<float>(m_currentPage) / m_totalPages;
        emit progressChanged(m_progress);
    }
}

void BookEngine::saveProgress()
{
    if (m_currentBookId.isEmpty()) return;
    
    QSqlQuery query(m_database);
    query.prepare(R"(
        INSERT OR REPLACE INTO reading_progress 
        (book_id, current_page, total_pages, percentage, last_read, time_spent)
        VALUES (?, ?, ?, ?, ?, COALESCE((SELECT time_spent FROM reading_progress WHERE book_id = ?), 0) + 1)
    )");
    
    query.bindValue(0, m_currentBookId);
    query.bindValue(1, m_currentPage);
    query.bindValue(2, m_totalPages);
    query.bindValue(3, m_progress);
    query.bindValue(4, QDateTime::currentSecsSinceEpoch());
    query.bindValue(5, m_currentBookId);  // For the time_spent calculation
    
    if (!query.exec()) {
        qWarning() << "Failed to save progress:" << query.lastError().text();
    }
}

BookProgress BookEngine::getProgress(const QString &bookId) const
{
    BookProgress progress;
    
    QSqlQuery query(m_database);
    query.prepare("SELECT * FROM reading_progress WHERE book_id = ?");
    query.bindValue(0, bookId);
    query.exec();
    
    if (query.next()) {
        progress.bookId = query.value("book_id").toString();
        progress.currentPage = query.value("current_page").toInt();
        progress.totalPages = query.value("total_pages").toInt();
        progress.percentage = query.value("percentage").toFloat();
        progress.lastRead = query.value("last_read").toLongLong();
        progress.timeSpent = query.value("time_spent").toInt();
    }
    
    return progress;
}

QString BookEngine::generateBookId(const QString &filePath)
{
    // Generate unique ID based on file path and modification time
    QFileInfo fileInfo(filePath);
    QString data = filePath + QString::number(fileInfo.lastModified().toSecsSinceEpoch());
    
    QCryptographicHash hash(QCryptographicHash::Sha256);
    hash.addData(data.toUtf8());
    
    return hash.result().toHex().left(16);  // Use first 16 characters
}

void BookEngine::saveBookMetadata(const BookMetadata &metadata)
{
    QSqlQuery query(m_database);
    query.prepare(R"(
        INSERT OR REPLACE INTO books 
        (id, title, author, language, publisher, isbn, description, 
         cover_path, file_path, format, page_count, file_size, added_date)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    )");
    
    query.bindValue(0, metadata.id);
    query.bindValue(1, metadata.title);
    query.bindValue(2, metadata.author);
    query.bindValue(3, metadata.language);
    query.bindValue(4, metadata.publisher);
    query.bindValue(5, metadata.isbn);
    query.bindValue(6, metadata.description);
    query.bindValue(7, metadata.coverPath);
    query.bindValue(8, metadata.filePath);
    query.bindValue(9, metadata.format);
    query.bindValue(10, metadata.pageCount);
    query.bindValue(11, metadata.fileSize);
    query.bindValue(12, metadata.addedDate);
    
    if (!query.exec()) {
        qWarning() << "Failed to save book metadata:" << query.lastError().text();
    }
}

BookMetadata BookEngine::loadBookMetadata(const QString &bookId)
{
    BookMetadata metadata;
    
    QSqlQuery query(m_database);
    query.prepare("SELECT * FROM books WHERE id = ?");
    query.bindValue(0, bookId);
    query.exec();
    
    if (query.next()) {
        metadata.id = query.value("id").toString();
        metadata.title = query.value("title").toString();
        metadata.author = query.value("author").toString();
        metadata.language = query.value("language").toString();
        metadata.publisher = query.value("publisher").toString();
        metadata.isbn = query.value("isbn").toString();
        metadata.description = query.value("description").toString();
        metadata.coverPath = query.value("cover_path").toString();
        metadata.filePath = query.value("file_path").toString();
        metadata.format = query.value("format").toString();
        metadata.pageCount = query.value("page_count").toInt();
        metadata.fileSize = query.value("file_size").toLongLong();
        metadata.addedDate = query.value("added_date").toLongLong();
    }
    
    return metadata;
}

QStringList BookEngine::getLibraryBooks() const
{
    QStringList bookIds;
    
    QSqlQuery query(m_database);
    query.exec("SELECT id FROM books ORDER BY added_date DESC");
    
    while (query.next()) {
        bookIds.append(query.value("id").toString());
    }
    
    return bookIds;
}

BookMetadata BookEngine::getBookMetadata(const QString &bookId) const
{
    return const_cast<BookEngine*>(this)->loadBookMetadata(bookId);
}

void BookEngine::onSyncComplete()
{
    // TODO: Implement server synchronization
    qDebug() << "Progress synced with server";
}

void BookEngine::onError(const QString &error)
{
    qWarning() << "BookEngine error:" << error;
    emit this->error(error);
}

// Placeholder implementations for MVP
void BookEngine::goToPage(int pageNumber) { setCurrentPage(pageNumber); }
void BookEngine::goToPosition(int position) { /* TODO */ }
void BookEngine::goToChapter(int chapterIndex) { /* TODO */ }
QStringList BookEngine::searchInBook(const QString &query) const { return QStringList(); }
int BookEngine::findTextPosition(const QString &text) const { return -1; }
void BookEngine::addBookmark(const QString &name) { /* TODO */ }
QStringList BookEngine::getBookmarks() const { return QStringList(); }
bool BookEngine::isFormatSupported(const QString &format) const { 
    return format.toLower() == "txt" || format.toLower() == "epub"; 
}
QString BookEngine::convertFormat(const QString &inputPath, const QString &outputFormat) { return QString(); }
void BookEngine::setFont(const QFont &font) { m_font = font; }
void BookEngine::setPageSize(const QSize &size) { m_pageSize = size; }
void BookEngine::setMargins(int left, int top, int right, int bottom) { 
    m_margins = QMargins(left, top, right, bottom); 
}
void BookEngine::setLineSpacing(float spacing) { m_lineSpacing = spacing; }