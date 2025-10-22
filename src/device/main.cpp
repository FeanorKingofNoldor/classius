#include <QApplication>
#include <QQmlApplicationEngine>
#include <QQmlContext>
#include <QQuickStyle>
#include <QDir>
#include <QStandardPaths>
#include <QLoggingCategory>

#include "ui/ReadingView.h"
#include "ui/LibraryView.h"
#include "ui/SageInterface.h"
#include "core/BookEngine.h"
#include "core/AnnotationManager.h"

Q_LOGGING_CATEGORY(classicusMain, "classius.main")

void setupApplicationSettings()
{
    QCoreApplication::setApplicationName("Classius");
    QCoreApplication::setApplicationVersion("0.1.0");
    QCoreApplication::setOrganizationName("Classius");
    QCoreApplication::setOrganizationDomain("classius.com");
}

void registerQmlTypes()
{
    // Register C++ types for QML
    qmlRegisterType<ReadingView>("Classius", 1, 0, "ReadingView");
    qmlRegisterType<LibraryView>("Classius", 1, 0, "LibraryView");
    qmlRegisterType<SageInterface>("Classius", 1, 0, "SageInterface");
}

void setupDataDirectories()
{
    // Create application data directories
    QString dataPath = QStandardPaths::writableLocation(QStandardPaths::AppDataLocation);
    QDir().mkpath(dataPath + "/books");
    QDir().mkpath(dataPath + "/notes");
    QDir().mkpath(dataPath + "/cache");
    
    qCDebug(classicusMain) << "Data directory:" << dataPath;
}

int main(int argc, char *argv[])
{
    QApplication app(argc, argv);
    
    // Setup application
    setupApplicationSettings();
    setupDataDirectories();
    
    // Set quick style for e-ink optimization
    QQuickStyle::setStyle("Material");
    
    // Register QML types
    registerQmlTypes();
    
    // Create core components
    BookEngine bookEngine;
    AnnotationManager annotationManager;
    
    // Create QML engine
    QQmlApplicationEngine engine;
    
    // Expose C++ objects to QML
    engine.rootContext()->setContextProperty("bookEngine", &bookEngine);
    engine.rootContext()->setContextProperty("annotationManager", &annotationManager);
    
    // Load main QML file
    const QUrl url("qrc:/ui/main.qml");
    QObject::connect(&engine, &QQmlApplicationEngine::objectCreated,
                     &app, [url](QObject *obj, const QUrl &objUrl) {
        if (!obj && url == objUrl)
            QCoreApplication::exit(-1);
    }, Qt::QueuedConnection);
    
    engine.load(url);
    
    // Check if QML loaded successfully
    if (engine.rootObjects().isEmpty()) {
        qCCritical(classicusMain) << "Failed to load QML interface";
        return -1;
    }
    
    qCInfo(classicusMain) << "Classius device application started successfully";
    
    return app.exec();
}