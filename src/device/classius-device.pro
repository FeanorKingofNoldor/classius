QT += core gui widgets qml quick quickcontrols2 network sql

CONFIG += c++17

TARGET = classius
TEMPLATE = app

# Source files
SOURCES += \
    main.cpp \
    core/BookEngine.cpp \
    core/TextPaginator.cpp \
    core/AnnotationManager.cpp \
    ui/ReadingView.cpp \
    ui/LibraryView.cpp \
    ui/SageInterface.cpp

# Header files
HEADERS += \
    core/BookEngine.h \
    core/TextPaginator.h \
    core/AnnotationManager.h \
    ui/ReadingView.h \
    ui/LibraryView.h \
    ui/SageInterface.h

# QML files
RESOURCES += qml.qrc

# Include paths
INCLUDEPATH += . core audio ui

# Libraries (audio libraries commented out for basic build)
# LIBS += -lportaudio -lfftw3f -lm

# Cross-compilation for ARM
arm-cross {
    QMAKE_CXX = arm-linux-gnueabihf-g++
    QMAKE_CC = arm-linux-gnueabihf-gcc
    QMAKE_LINK = arm-linux-gnueabihf-g++
    QMAKE_AR = arm-linux-gnueabihf-ar
    
    # ARM-specific libraries
    LIBS += -L/usr/arm-linux-gnueabihf/lib
}

# Debug configuration
CONFIG(debug, debug|release) {
    DEFINES += DEBUG_MODE
    QMAKE_CXXFLAGS += -g -O0
}

# Release configuration
CONFIG(release, debug|release) {
    DEFINES += RELEASE_MODE
    QMAKE_CXXFLAGS += -O3 -DNDEBUG
}

# Install target
target.path = /usr/local/bin
INSTALLS += target