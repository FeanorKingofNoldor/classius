#include "AnnotationManager.h"
#include <QSqlQuery>
#include <QSqlError>
#include <QStandardPaths>
#include <QDir>
#include <QDebug>
#include <QJsonDocument>
#include <QJsonObject>
#include <QJsonArray>
#include <QFile>
#include <QUuid>
#include <QCoreApplication>

// Constants
const QString AnnotationManager::DB_NAME = "annotations.db";
const int AnnotationManager::CACHE_LIMIT = 1000;

// Annotation struct implementation
QVariantMap Annotation::toVariant() const {
    QVariantMap map;
    map["id"] = id;
    map["bookId"] = bookId;
    map["type"] = type;
    map["pageNumber"] = pageNumber;
    map["startPosition"] = startPosition;
    map["endPosition"] = endPosition;
    map["selectedText"] = selectedText;
    map["content"] = content;
    map["color"] = color;
    map["tags"] = tags;
    map["isPrivate"] = isPrivate;
    map["createdAt"] = createdAt;
    map["updatedAt"] = updatedAt;
    return map;
}

Annotation Annotation::fromVariant(const QVariantMap& map) {
    Annotation annotation;
    annotation.id = map["id"].toString();
    annotation.bookId = map["bookId"].toString();
    annotation.type = map["type"].toString();
    annotation.pageNumber = map["pageNumber"].toInt();
    annotation.startPosition = map["startPosition"].toInt();
    annotation.endPosition = map["endPosition"].toInt();
    annotation.selectedText = map["selectedText"].toString();
    annotation.content = map["content"].toString();
    annotation.color = map["color"].toString();
    annotation.tags = map["tags"].toStringList();
    annotation.isPrivate = map["isPrivate"].toBool();
    annotation.createdAt = map["createdAt"].toDateTime();
    annotation.updatedAt = map["updatedAt"].toDateTime();
    return annotation;
}

AnnotationManager::AnnotationManager(QObject *parent)
    : QObject(parent)
    , m_connected(false)
    , m_hasUnsyncedChanges(false)
{
    // Initialize with local SQLite database
    connectToDatabase();
}

AnnotationManager::~AnnotationManager() {
    if (m_database.isOpen()) {
        m_database.close();
    }
}

bool AnnotationManager::connectToDatabase(const QString& connectionString) {
    QString dbPath;
    
    if (!connectionString.isEmpty()) {
        m_connectionString = connectionString;
        // Parse connection string for remote database
        // For now, we'll use local SQLite
    }
    
    // Use local SQLite database
    QString dataDir = QStandardPaths::writableLocation(QStandardPaths::AppDataLocation);
    QDir().mkpath(dataDir);
    dbPath = QDir(dataDir).filePath(DB_NAME);
    
    m_database = QSqlDatabase::addDatabase("QSQLITE", "annotations");
    m_database.setDatabaseName(dbPath);
    
    if (!m_database.open()) {
        qWarning() << "Failed to open annotations database:" << m_database.lastError().text();
        return false;
    }
    
    if (!initializeDatabase()) {
        qWarning() << "Failed to initialize annotations database";
        return false;
    }
    
    m_connected = true;
    emit connectionChanged(true);
    
    // Load existing annotations
    loadAnnotationsFromDatabase();
    
    qDebug() << "AnnotationManager: Connected to database at" << dbPath;
    return true;
}

bool AnnotationManager::initializeDatabase() {
    if (!m_database.isOpen()) {
        return false;
    }
    
    return createTables();
}

bool AnnotationManager::createTables() {
    QSqlQuery query(m_database);
    
    // Create annotations table
    QString createTableSql = R"(
        CREATE TABLE IF NOT EXISTS annotations (
            id TEXT PRIMARY KEY,
            book_id TEXT NOT NULL,
            type TEXT NOT NULL,
            page_number INTEGER,
            start_position INTEGER,
            end_position INTEGER,
            selected_text TEXT,
            content TEXT,
            color TEXT,
            tags TEXT,
            is_private INTEGER DEFAULT 1,
            created_at TEXT,
            updated_at TEXT,
            synced INTEGER DEFAULT 0
        )
    )";
    
    if (!query.exec(createTableSql)) {
        qWarning() << "Failed to create annotations table:" << query.lastError().text();
        return false;
    }
    
    // Create indexes
    query.exec("CREATE INDEX IF NOT EXISTS idx_annotations_book_id ON annotations(book_id)");
    query.exec("CREATE INDEX IF NOT EXISTS idx_annotations_type ON annotations(type)");
    query.exec("CREATE INDEX IF NOT EXISTS idx_annotations_page ON annotations(book_id, page_number)");
    
    return true;
}

QString AnnotationManager::generateId() const {
    return QUuid::createUuid().toString(QUuid::WithoutBraces);
}

QString AnnotationManager::createHighlight(const QString& bookId, int pageNumber,
                                         int startPos, int endPos, 
                                         const QString& selectedText,
                                         const QString& color) {
    Annotation annotation;
    annotation.id = generateId();
    annotation.bookId = bookId;
    annotation.type = "highlight";
    annotation.pageNumber = pageNumber;
    annotation.startPosition = startPos;
    annotation.endPosition = endPos;
    annotation.selectedText = selectedText;
    annotation.color = color;
    annotation.isPrivate = true;
    annotation.createdAt = QDateTime::currentDateTime();
    annotation.updatedAt = annotation.createdAt;
    
    if (saveAnnotation(annotation)) {
        addToCache(annotation);
        emit annotationCreated(annotation.id, annotation.toVariant());
        emit annotationCountChanged(m_annotations.size());
        m_hasUnsyncedChanges = true;
        return annotation.id;
    }
    
    return QString();
}

QString AnnotationManager::createNote(const QString& bookId, int pageNumber,
                                    int startPos, int endPos,
                                    const QString& selectedText,
                                    const QString& noteContent) {
    Annotation annotation;
    annotation.id = generateId();
    annotation.bookId = bookId;
    annotation.type = "note";
    annotation.pageNumber = pageNumber;
    annotation.startPosition = startPos;
    annotation.endPosition = endPos;
    annotation.selectedText = selectedText;
    annotation.content = noteContent;
    annotation.isPrivate = true;
    annotation.createdAt = QDateTime::currentDateTime();
    annotation.updatedAt = annotation.createdAt;
    
    if (saveAnnotation(annotation)) {
        addToCache(annotation);
        emit annotationCreated(annotation.id, annotation.toVariant());
        emit annotationCountChanged(m_annotations.size());
        m_hasUnsyncedChanges = true;
        return annotation.id;
    }
    
    return QString();
}

QString AnnotationManager::createBookmark(const QString& bookId, int pageNumber,
                                       const QString& name) {
    Annotation annotation;
    annotation.id = generateId();
    annotation.bookId = bookId;
    annotation.type = "bookmark";
    annotation.pageNumber = pageNumber;
    annotation.content = name.isEmpty() ? QString("Page %1").arg(pageNumber) : name;
    annotation.isPrivate = true;
    annotation.createdAt = QDateTime::currentDateTime();
    annotation.updatedAt = annotation.createdAt;
    
    if (saveAnnotation(annotation)) {
        addToCache(annotation);
        emit annotationCreated(annotation.id, annotation.toVariant());
        emit annotationCountChanged(m_annotations.size());
        m_hasUnsyncedChanges = true;
        return annotation.id;
    }
    
    return QString();
}

bool AnnotationManager::saveAnnotation(const Annotation& annotation) {
    if (!m_database.isOpen()) {
        emit error("Database not connected");
        return false;
    }
    
    QSqlQuery query(m_database);
    query.prepare(R"(
        INSERT OR REPLACE INTO annotations 
        (id, book_id, type, page_number, start_position, end_position, 
         selected_text, content, color, tags, is_private, created_at, updated_at, synced)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
    )");
    
    query.addBindValue(annotation.id);
    query.addBindValue(annotation.bookId);
    query.addBindValue(annotation.type);
    query.addBindValue(annotation.pageNumber);
    query.addBindValue(annotation.startPosition);
    query.addBindValue(annotation.endPosition);
    query.addBindValue(annotation.selectedText);
    query.addBindValue(annotation.content);
    query.addBindValue(annotation.color);
    query.addBindValue(annotation.tags.join(","));
    query.addBindValue(annotation.isPrivate ? 1 : 0);
    query.addBindValue(annotation.createdAt.toString(Qt::ISODate));
    query.addBindValue(annotation.updatedAt.toString(Qt::ISODate));
    
    if (!query.exec()) {
        qWarning() << "Failed to save annotation:" << query.lastError().text();
        emit error("Failed to save annotation: " + query.lastError().text());
        return false;
    }
    
    return true;
}

bool AnnotationManager::loadAnnotationsFromDatabase() {
    if (!m_database.isOpen()) {
        return false;
    }
    
    QSqlQuery query("SELECT * FROM annotations ORDER BY created_at DESC", m_database);
    
    m_annotations.clear();
    m_annotationIndex.clear();
    
    int index = 0;
    while (query.next()) {
        Annotation annotation;
        annotation.id = query.value("id").toString();
        annotation.bookId = query.value("book_id").toString();
        annotation.type = query.value("type").toString();
        annotation.pageNumber = query.value("page_number").toInt();
        annotation.startPosition = query.value("start_position").toInt();
        annotation.endPosition = query.value("end_position").toInt();
        annotation.selectedText = query.value("selected_text").toString();
        annotation.content = query.value("content").toString();
        annotation.color = query.value("color").toString();
        
        QString tagsStr = query.value("tags").toString();
        annotation.tags = tagsStr.isEmpty() ? QStringList() : tagsStr.split(",");
        
        annotation.isPrivate = query.value("is_private").toInt() == 1;
        annotation.createdAt = QDateTime::fromString(query.value("created_at").toString(), Qt::ISODate);
        annotation.updatedAt = QDateTime::fromString(query.value("updated_at").toString(), Qt::ISODate);
        
        m_annotations.append(annotation);
        m_annotationIndex[annotation.id] = index++;
    }
    
    emit annotationCountChanged(m_annotations.size());
    qDebug() << "Loaded" << m_annotations.size() << "annotations from database";
    return true;
}

QVariantList AnnotationManager::getAnnotationsForBook(const QString& bookId) const {
    QVariantList result;
    
    for (const Annotation& annotation : m_annotations) {
        if (annotation.bookId == bookId) {
            result.append(annotation.toVariant());
        }
    }
    
    return result;
}

QVariantList AnnotationManager::getAnnotationsForPage(const QString& bookId, int pageNumber) const {
    QVariantList result;
    
    for (const Annotation& annotation : m_annotations) {
        if (annotation.bookId == bookId && annotation.pageNumber == pageNumber) {
            result.append(annotation.toVariant());
        }
    }
    
    return result;
}

QVariantMap AnnotationManager::getAnnotation(const QString& annotationId) const {
    auto it = m_annotationIndex.find(annotationId);
    if (it != m_annotationIndex.end() && it.value() < m_annotations.size()) {
        return m_annotations[it.value()].toVariant();
    }
    return QVariantMap();
}

bool AnnotationManager::deleteAnnotation(const QString& annotationId) {
    if (!m_database.isOpen()) {
        emit error("Database not connected");
        return false;
    }
    
    QSqlQuery query(m_database);
    query.prepare("DELETE FROM annotations WHERE id = ?");
    query.addBindValue(annotationId);
    
    if (!query.exec()) {
        emit error("Failed to delete annotation: " + query.lastError().text());
        return false;
    }
    
    removeFromCache(annotationId);
    emit annotationDeleted(annotationId);
    emit annotationCountChanged(m_annotations.size());
    m_hasUnsyncedChanges = true;
    
    return true;
}

void AnnotationManager::addToCache(const Annotation& annotation) {
    m_annotationIndex[annotation.id] = m_annotations.size();
    m_annotations.append(annotation);
    
    // Simple cache limit management
    if (m_annotations.size() > CACHE_LIMIT) {
        // Remove oldest entries (simple strategy)
        auto oldestId = m_annotations.first().id;
        m_annotations.removeFirst();
        m_annotationIndex.remove(oldestId);
        
        // Rebuild index
        for (int i = 0; i < m_annotations.size(); ++i) {
            m_annotationIndex[m_annotations[i].id] = i;
        }
    }
}

void AnnotationManager::removeFromCache(const QString& annotationId) {
    auto it = m_annotationIndex.find(annotationId);
    if (it != m_annotationIndex.end()) {
        int index = it.value();
        m_annotations.removeAt(index);
        m_annotationIndex.remove(annotationId);
        
        // Update indices
        for (int i = index; i < m_annotations.size(); ++i) {
            m_annotationIndex[m_annotations[i].id] = i;
        }
    }
}

void AnnotationManager::onBookChanged(const QString& bookId) {
    m_currentBookId = bookId;
    // Could preload annotations for current book here
}

void AnnotationManager::clearCache() {
    m_annotations.clear();
    m_annotationIndex.clear();
    emit annotationCountChanged(0);
}

int AnnotationManager::getHighlightCount(const QString& bookId) const {
    int count = 0;
    for (const Annotation& annotation : m_annotations) {
        if (annotation.bookId == bookId && annotation.type == "highlight") {
            count++;
        }
    }
    return count;
}

int AnnotationManager::getNoteCount(const QString& bookId) const {
    int count = 0;
    for (const Annotation& annotation : m_annotations) {
        if (annotation.bookId == bookId && annotation.type == "note") {
            count++;
        }
    }
    return count;
}

int AnnotationManager::getBookmarkCount(const QString& bookId) const {
    int count = 0;
    for (const Annotation& annotation : m_annotations) {
        if (annotation.bookId == bookId && annotation.type == "bookmark") {
            count++;
        }
    }
    return count;
}

void AnnotationManager::syncAnnotations() {
    // Placeholder for server sync functionality
    emit syncStatusChanged(false);
    qDebug() << "Annotation sync completed (local only for now)";
}

bool AnnotationManager::hasUnsyncedChanges() const {
    return m_hasUnsyncedChanges;
}

bool AnnotationManager::updateAnnotation(const QString& annotationId, const QVariantMap& updates) {
    auto it = m_annotationIndex.find(annotationId);
    if (it == m_annotationIndex.end() || it.value() >= m_annotations.size()) {
        emit error("Annotation not found: " + annotationId);
        return false;
    }
    
    Annotation& annotation = m_annotations[it.value()];
    
    // Update fields if provided
    if (updates.contains("content")) {
        annotation.content = updates["content"].toString();
    }
    if (updates.contains("color")) {
        annotation.color = updates["color"].toString();
    }
    if (updates.contains("tags")) {
        annotation.tags = updates["tags"].toStringList();
    }
    if (updates.contains("isPrivate")) {
        annotation.isPrivate = updates["isPrivate"].toBool();
    }
    
    annotation.updatedAt = QDateTime::currentDateTime();
    
    if (saveAnnotation(annotation)) {
        emit annotationUpdated(annotationId, annotation.toVariant());
        m_hasUnsyncedChanges = true;
        return true;
    }
    
    return false;
}

QVariantList AnnotationManager::getHighlights(const QString& bookId) const {
    QVariantList result;
    
    for (const Annotation& annotation : m_annotations) {
        if (annotation.bookId == bookId && annotation.type == "highlight") {
            result.append(annotation.toVariant());
        }
    }
    
    return result;
}

QVariantList AnnotationManager::getNotes(const QString& bookId) const {
    QVariantList result;
    
    for (const Annotation& annotation : m_annotations) {
        if (annotation.bookId == bookId && annotation.type == "note") {
            result.append(annotation.toVariant());
        }
    }
    
    return result;
}

QVariantList AnnotationManager::getBookmarks(const QString& bookId) const {
    QVariantList result;
    
    for (const Annotation& annotation : m_annotations) {
        if (annotation.bookId == bookId && annotation.type == "bookmark") {
            result.append(annotation.toVariant());
        }
    }
    
    return result;
}

QVariantList AnnotationManager::searchAnnotations(const QString& query) const {
    QVariantList result;
    QString lowerQuery = query.toLower();
    
    for (const Annotation& annotation : m_annotations) {
        bool matches = false;
        
        // Search in selected text
        if (annotation.selectedText.toLower().contains(lowerQuery)) {
            matches = true;
        }
        // Search in note content
        else if (annotation.content.toLower().contains(lowerQuery)) {
            matches = true;
        }
        // Search in tags
        else {
            for (const QString& tag : annotation.tags) {
                if (tag.toLower().contains(lowerQuery)) {
                    matches = true;
                    break;
                }
            }
        }
        
        if (matches) {
            result.append(annotation.toVariant());
        }
    }
    
    return result;
}

bool AnnotationManager::exportAnnotations(const QString& bookId, const QString& filePath) const {
    QJsonArray annotationsArray;
    
    for (const Annotation& annotation : m_annotations) {
        if (bookId.isEmpty() || annotation.bookId == bookId) {
            QJsonObject obj;
            obj["id"] = annotation.id;
            obj["bookId"] = annotation.bookId;
            obj["type"] = annotation.type;
            obj["pageNumber"] = annotation.pageNumber;
            obj["startPosition"] = annotation.startPosition;
            obj["endPosition"] = annotation.endPosition;
            obj["selectedText"] = annotation.selectedText;
            obj["content"] = annotation.content;
            obj["color"] = annotation.color;
            obj["tags"] = QJsonArray::fromStringList(annotation.tags);
            obj["isPrivate"] = annotation.isPrivate;
            obj["createdAt"] = annotation.createdAt.toString(Qt::ISODate);
            obj["updatedAt"] = annotation.updatedAt.toString(Qt::ISODate);
            annotationsArray.append(obj);
        }
    }
    
    QJsonDocument doc(annotationsArray);
    
    QFile file(filePath);
    if (!file.open(QIODevice::WriteOnly)) {
        return false;
    }
    
    file.write(doc.toJson());
    return true;
}

bool AnnotationManager::importAnnotations(const QString& filePath) {
    QFile file(filePath);
    if (!file.open(QIODevice::ReadOnly)) {
        emit error("Could not open file: " + filePath);
        return false;
    }
    
    QByteArray data = file.readAll();
    QJsonParseError error;
    QJsonDocument doc = QJsonDocument::fromJson(data, &error);
    
    if (error.error != QJsonParseError::NoError) {
        emit this->error("JSON parse error: " + error.errorString());
        return false;
    }
    
    if (!doc.isArray()) {
        emit this->error("Invalid annotation file format");
        return false;
    }
    
    QJsonArray annotationsArray = doc.array();
    int imported = 0;
    
    for (const QJsonValue& value : annotationsArray) {
        if (!value.isObject()) continue;
        
        QJsonObject obj = value.toObject();
        Annotation annotation;
        
        annotation.id = obj["id"].toString();
        annotation.bookId = obj["bookId"].toString();
        annotation.type = obj["type"].toString();
        annotation.pageNumber = obj["pageNumber"].toInt();
        annotation.startPosition = obj["startPosition"].toInt();
        annotation.endPosition = obj["endPosition"].toInt();
        annotation.selectedText = obj["selectedText"].toString();
        annotation.content = obj["content"].toString();
        annotation.color = obj["color"].toString();
        
        QJsonArray tagsArray = obj["tags"].toArray();
        for (const QJsonValue& tagValue : tagsArray) {
            annotation.tags.append(tagValue.toString());
        }
        
        annotation.isPrivate = obj["isPrivate"].toBool();
        annotation.createdAt = QDateTime::fromString(obj["createdAt"].toString(), Qt::ISODate);
        annotation.updatedAt = QDateTime::fromString(obj["updatedAt"].toString(), Qt::ISODate);
        
        if (saveAnnotation(annotation)) {
            addToCache(annotation);
            imported++;
        }
    }
    
    if (imported > 0) {
        emit annotationCountChanged(m_annotations.size());
        m_hasUnsyncedChanges = true;
    }
    
    qDebug() << "Imported" << imported << "annotations";
    return true;
}

// MOC will be included automatically by the build system
