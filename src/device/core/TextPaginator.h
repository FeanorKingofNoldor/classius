#ifndef TEXTPAGINATOR_H
#define TEXTPAGINATOR_H

#include <QObject>
#include <QString>
#include <QStringList>
#include <QFont>
#include <QSize>
#include <QMargins>
#include <QTextDocument>
#include <QTextLayout>
#include <QFontMetrics>

/**
 * @brief Handles text pagination optimized for e-ink displays
 * 
 * The TextPaginator takes raw text and breaks it into pages that fit
 * perfectly on the device screen. It considers:
 * - Font size and family
 * - Screen dimensions
 * - Margins and spacing
 * - Line breaks and paragraph formatting
 * - E-ink refresh optimization
 */
class TextPaginator : public QObject
{
    Q_OBJECT

public:
    explicit TextPaginator(QObject *parent = nullptr);
    
    /**
     * @brief Paginate text with the given formatting parameters
     * @param text The full text to paginate
     * @param font Font to use for rendering
     * @param pageSize Available page size (width x height)
     * @param margins Page margins (left, top, right, bottom)
     * @param lineSpacing Line spacing multiplier (e.g., 1.2 for 120%)
     */
    void paginateText(const QString &text, 
                     const QFont &font,
                     const QSize &pageSize,
                     const QMargins &margins,
                     float lineSpacing = 1.2f);
    
    /**
     * @brief Get the paginated pages
     * @return List of pages, each containing text that fits on one screen
     */
    QStringList getPages() const { return m_pages; }
    
    /**
     * @brief Get total number of pages
     */
    int getTotalPages() const { return m_pages.size(); }
    
    /**
     * @brief Get specific page content
     * @param pageIndex Zero-based page index
     * @return Page content or empty string if invalid index
     */
    QString getPage(int pageIndex) const;
    
    /**
     * @brief Check if pagination is currently in progress
     */
    bool isPaginating() const { return m_isPaginating; }

signals:
    /**
     * @brief Emitted when pagination is complete
     */
    void paginationComplete();
    
    /**
     * @brief Emitted during pagination to show progress
     * @param percentage Progress from 0.0 to 1.0
     */
    void paginationProgress(float percentage);

private slots:
    void performPagination();

private:
    // Core pagination methods
    void calculateTextMetrics();
    int calculateLinesPerPage();
    void breakIntoPages();
    QString wrapTextToWidth(const QString &text, int maxWidth);
    int getTextWidth(const QString &text) const;
    int getLineHeight() const;
    
    // Text processing helpers
    QStringList splitIntoParagraphs(const QString &text);
    QStringList wrapParagraphToLines(const QString &paragraph);
    QString processLine(const QString &line);
    bool isPageBreakNeeded(const QString &currentPage, const QString &nextLine);
    
    // Layout optimization for e-ink
    void optimizeForEInk();
    void adjustForReadability();
    
    // Member variables
    QString m_sourceText;
    QFont m_font;
    QSize m_pageSize;
    QMargins m_margins;
    float m_lineSpacing;
    
    // Calculated dimensions
    QSize m_textArea;           // Available text area (pageSize - margins)
    QFontMetrics m_fontMetrics;
    int m_lineHeight;
    int m_linesPerPage;
    int m_charactersPerLine;
    
    // Results
    QStringList m_pages;
    bool m_isPaginating;
};

#endif // TEXTPAGINATOR_H