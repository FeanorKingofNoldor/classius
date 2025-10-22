#ifndef ANNOTATIONMANAGER_H
#define ANNOTATIONMANAGER_H

#include <QObject>
#include <QString>
#include <QDateTime>
#include <QList>
#include <QVariantMap>
#include <QColor>
#include <QSqlDatabase>
#include <QUuid>

// Annotation data structure
struct Annotation {
    QString id;
    QString bookId;
    QString type;  // "highlight", "note", "bookmark"
    int pageNumber;
    int startPosition;
    int endPosition;
    QString selectedText;
    QString content;  // Note content
    QString color;    // Highlight color
    QStringList tags;
    bool isPrivate;
    QDateTime createdAt;
    QDateTime updatedAt;
    
    // Convert to QVariantMap for QML
    QVariantMap toVariant() const;
    static Annotation fromVariant(const QVariantMap& map);
};

class AnnotationManager : public QObject
{
    Q_OBJECT
    Q_PROPERTY(bool isConnected READ isConnected NOTIFY connectionChanged)
    Q_PROPERTY(int annotationCount READ annotationCount NOTIFY annotationCountChanged)

public:
    explicit AnnotationManager(QObject *parent = nullptr);
    ~AnnotationManager();
    
    // Database connection
    Q_INVOKABLE bool connectToDatabase(const QString& connectionString = "");
    bool isConnected() const { return m_connected; }
    
    // Annotation management
    Q_INVOKABLE QString createHighlight(const QString& bookId, int pageNumber, 
                                       int startPos, int endPos, 
                                       const QString& selectedText,
                                       const QString& color = "#ffff00");
    
    Q_INVOKABLE QString createNote(const QString& bookId, int pageNumber,
                                  int startPos, int endPos,
                                  const QString& selectedText,
                                  const QString& noteContent);
    
    Q_INVOKABLE QString createBookmark(const QString& bookId, int pageNumber,
                                      const QString& name = "");
    
    Q_INVOKABLE bool updateAnnotation(const QString& annotationId,
                                     const QVariantMap& updates);
    
    Q_INVOKABLE bool deleteAnnotation(const QString& annotationId);
    
    // Retrieval
    Q_INVOKABLE QVariantList getAnnotationsForBook(const QString& bookId) const;
    Q_INVOKABLE QVariantList getAnnotationsForPage(const QString& bookId, int pageNumber) const;
    Q_INVOKABLE QVariantMap getAnnotation(const QString& annotationId) const;
    
    // Search and filtering
    Q_INVOKABLE QVariantList searchAnnotations(const QString& query) const;
    Q_INVOKABLE QVariantList getHighlights(const QString& bookId) const;
    Q_INVOKABLE QVariantList getNotes(const QString& bookId) const;
    Q_INVOKABLE QVariantList getBookmarks(const QString& bookId) const;
    
    // Statistics
    int annotationCount() const { return m_annotations.size(); }
    Q_INVOKABLE int getHighlightCount(const QString& bookId) const;
    Q_INVOKABLE int getNoteCount(const QString& bookId) const;
    Q_INVOKABLE int getBookmarkCount(const QString& bookId) const;
    
    // Export/Import
    Q_INVOKABLE bool exportAnnotations(const QString& bookId, const QString& filePath) const;
    Q_INVOKABLE bool importAnnotations(const QString& filePath);
    
    // Sync status
    Q_INVOKABLE void syncAnnotations();
    Q_INVOKABLE bool hasUnsyncedChanges() const;

public slots:
    void onBookChanged(const QString& bookId);
    void clearCache();

signals:
    void connectionChanged(bool connected);
    void annotationCreated(const QString& annotationId, const QVariantMap& annotation);
    void annotationUpdated(const QString& annotationId, const QVariantMap& annotation);
    void annotationDeleted(const QString& annotationId);
    void annotationCountChanged(int count);
    void syncStatusChanged(bool syncing);
    void error(const QString& message);

private:
    bool initializeDatabase();
    bool createTables();
    QString generateId() const;
    bool saveAnnotation(const Annotation& annotation);
    bool loadAnnotationsFromDatabase();
    void addToCache(const Annotation& annotation);
    void removeFromCache(const QString& annotationId);
    void updateCache(const Annotation& annotation);
    
    // Database
    QSqlDatabase m_database;
    bool m_connected;
    QString m_connectionString;
    
    // Cache for quick access
    QList<Annotation> m_annotations;
    QHash<QString, int> m_annotationIndex;  // ID -> index mapping
    
    // Current state
    QString m_currentBookId;
    bool m_hasUnsyncedChanges;
    
    // Constants
    static const QString DB_NAME;
    static const int CACHE_LIMIT;
};

#endif // ANNOTATIONMANAGER_H
