#include "TextPaginator.h"
#include <QDebug>
#include <QTimer>
#include <QTextLayout>
#include <QTextOption>
#include <QRegularExpression>

TextPaginator::TextPaginator(QObject *parent)
    : QObject(parent)
    , m_lineSpacing(1.2f)
    , m_fontMetrics(QFont())
    , m_lineHeight(0)
    , m_linesPerPage(0)
    , m_charactersPerLine(0)
    , m_isPaginating(false)
{
}

void TextPaginator::paginateText(const QString &text, 
                                const QFont &font,
                                const QSize &pageSize,
                                const QMargins &margins,
                                float lineSpacing)
{
    if (m_isPaginating) {
        qWarning() << "Pagination already in progress";
        return;
    }
    
    // Store parameters
    m_sourceText = text;
    m_font = font;
    m_pageSize = pageSize;
    m_margins = margins;
    m_lineSpacing = lineSpacing;
    
    // Clear previous results
    m_pages.clear();
    
    m_isPaginating = true;
    
    // Calculate text area (page size minus margins)
    m_textArea = QSize(
        pageSize.width() - margins.left() - margins.right(),
        pageSize.height() - margins.top() - margins.bottom()
    );
    
    // Update font metrics
    m_fontMetrics = QFontMetrics(font);
    
    qDebug() << "Starting pagination:"
             << "Page size:" << pageSize
             << "Text area:" << m_textArea
             << "Font:" << font.family() << font.pointSize() << "pt";
    
    // Start pagination in next event loop to avoid blocking UI
    QTimer::singleShot(0, this, &TextPaginator::performPagination);
}

void TextPaginator::performPagination()
{
    calculateTextMetrics();
    breakIntoPages();
    optimizeForEInk();
    
    m_isPaginating = false;
    
    qDebug() << "Pagination complete:" << m_pages.size() << "pages generated";
    emit paginationComplete();
}

void TextPaginator::calculateTextMetrics()
{
    m_lineHeight = static_cast<int>(m_fontMetrics.height() * m_lineSpacing);
    m_linesPerPage = m_textArea.height() / m_lineHeight;
    
    // Estimate characters per line (rough approximation)
    m_charactersPerLine = m_textArea.width() / m_fontMetrics.averageCharWidth();
    
    qDebug() << "Text metrics:"
             << "Line height:" << m_lineHeight
             << "Lines per page:" << m_linesPerPage
             << "Chars per line (est):" << m_charactersPerLine;
}

void TextPaginator::breakIntoPages()
{
    if (m_sourceText.isEmpty() || m_linesPerPage <= 0) {
        return;
    }
    
    // Split text into paragraphs
    QStringList paragraphs = splitIntoParagraphs(m_sourceText);
    
    QString currentPage;
    int currentPageLines = 0;
    int totalParagraphs = paragraphs.size();
    
    for (int i = 0; i < totalParagraphs; ++i) {
        const QString &paragraph = paragraphs[i];
        
        // Emit progress
        float progress = static_cast<float>(i) / totalParagraphs;
        emit paginationProgress(progress);
        
        // Wrap paragraph to fit line width
        QStringList wrappedLines = wrapParagraphToLines(paragraph);
        
        for (const QString &line : wrappedLines) {
            // Check if adding this line would exceed page capacity
            if (currentPageLines >= m_linesPerPage) {
                // Save current page and start new one
                if (!currentPage.trimmed().isEmpty()) {
                    m_pages.append(currentPage.trimmed());
                }
                currentPage.clear();
                currentPageLines = 0;
            }
            
            // Add line to current page
            if (!currentPage.isEmpty()) {
                currentPage += "\n";
            }
            currentPage += line;
            currentPageLines++;
        }
        
        // Add extra line between paragraphs (if not at start of page)
        if (i < totalParagraphs - 1 && currentPageLines < m_linesPerPage && !currentPage.isEmpty()) {
            currentPage += "\n";
            currentPageLines++;
        }
    }
    
    // Add final page if it has content
    if (!currentPage.trimmed().isEmpty()) {
        m_pages.append(currentPage.trimmed());
    }
}

QStringList TextPaginator::splitIntoParagraphs(const QString &text)
{
    // Split on double newlines (paragraph breaks)
    QStringList paragraphs = text.split(QRegularExpression("\n\\s*\n"), Qt::SkipEmptyParts);
    
    // Clean up each paragraph
    for (QString &paragraph : paragraphs) {
        paragraph = paragraph.trimmed();
        // Replace single newlines with spaces within paragraphs
        paragraph = paragraph.replace('\n', ' ');
        // Reduce multiple spaces to single space
        paragraph = paragraph.replace(QRegularExpression("\\s+"), " ");
    }
    
    return paragraphs;
}

QStringList TextPaginator::wrapParagraphToLines(const QString &paragraph)
{
    QStringList lines;
    
    if (paragraph.isEmpty()) {
        return lines;
    }
    
    QTextLayout textLayout(paragraph, m_font);
    textLayout.beginLayout();
    
    QTextOption option;
    option.setWrapMode(QTextOption::WrapAtWordBoundaryOrAnywhere);
    textLayout.setTextOption(option);
    
    int lineWidth = m_textArea.width();
    
    forever {
        QTextLine line = textLayout.createLine();
        if (!line.isValid()) {
            break;
        }
        
        line.setLineWidth(lineWidth);
        
        QString lineText = paragraph.mid(line.textStart(), line.textLength()).trimmed();
        if (!lineText.isEmpty()) {
            lines.append(lineText);
        }
    }
    
    textLayout.endLayout();
    
    return lines;
}

void TextPaginator::optimizeForEInk()
{
    // E-ink specific optimizations
    
    // 1. Avoid orphaned lines (single lines at top of page)
    for (int i = 1; i < m_pages.size(); ++i) {
        QString &currentPage = m_pages[i];
        QString &previousPage = m_pages[i - 1];
        
        QStringList currentLines = currentPage.split('\n');
        QStringList previousLines = previousPage.split('\n');
        
        // If current page starts with a single orphaned line, move it to previous page
        if (currentLines.size() > 1 && previousLines.size() < m_linesPerPage - 1) {
            QString firstLine = currentLines.takeFirst();
            previousPage += "\n" + firstLine;
            currentPage = currentLines.join('\n');
        }
    }
    
    // 2. Ensure minimum page content
    for (int i = 0; i < m_pages.size(); ++i) {
        QString &page = m_pages[i];
        QStringList lines = page.split('\n', Qt::SkipEmptyParts);
        
        // Remove pages with too little content
        if (lines.size() < 2 && i < m_pages.size() - 1) {
            // Merge with next page
            if (i + 1 < m_pages.size()) {
                m_pages[i + 1] = page + "\n\n" + m_pages[i + 1];
                m_pages.removeAt(i);
                --i; // Recheck this index
            }
        }
    }
}

void TextPaginator::adjustForReadability()
{
    // Additional readability optimizations can go here
    // For example:
    // - Widow/orphan control
    // - Hyphenation handling
    // - Chapter break detection
}

QString TextPaginator::getPage(int pageIndex) const
{
    if (pageIndex >= 0 && pageIndex < m_pages.size()) {
        return m_pages[pageIndex];
    }
    return QString();
}

int TextPaginator::getTextWidth(const QString &text) const
{
    return m_fontMetrics.horizontalAdvance(text);
}

int TextPaginator::getLineHeight() const
{
    return m_lineHeight;
}