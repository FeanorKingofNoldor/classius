/****************************************************************************
** Meta object code from reading C++ file 'BookEngine.h'
**
** Created by: The Qt Meta Object Compiler version 68 (Qt 6.8.2)
**
** WARNING! All changes made in this file will be lost!
*****************************************************************************/

#include "core/BookEngine.h"
#include <QtNetwork/QSslError>
#include <QtCore/qmetatype.h>

#include <QtCore/qtmochelpers.h>

#include <memory>


#include <QtCore/qxptype_traits.h>
#if !defined(Q_MOC_OUTPUT_REVISION)
#error "The header file 'BookEngine.h' doesn't include <QObject>."
#elif Q_MOC_OUTPUT_REVISION != 68
#error "This file was generated using the moc from 6.8.2. It"
#error "cannot be used with the include files from this version of Qt."
#error "(The moc has changed too much.)"
#endif

#ifndef Q_CONSTINIT
#define Q_CONSTINIT
#endif

QT_WARNING_PUSH
QT_WARNING_DISABLE_DEPRECATED
QT_WARNING_DISABLE_GCC("-Wuseless-cast")
namespace {
struct qt_meta_tag_ZN10BookEngineE_t {};
} // unnamed namespace


#ifdef QT_MOC_HAS_STRINGDATA
static constexpr auto qt_meta_stringdata_ZN10BookEngineE = QtMocHelpers::stringData(
    "BookEngine",
    "currentBookChanged",
    "",
    "bookLoaded",
    "bookId",
    "pageChanged",
    "page",
    "progressChanged",
    "progress",
    "textChanged",
    "loadingChanged",
    "loading",
    "error",
    "message",
    "bookAdded",
    "onPaginationComplete",
    "onSyncComplete",
    "onError",
    "loadBook",
    "loadBookFromFile",
    "filePath",
    "addBookToLibrary",
    "getLibraryBooks",
    "getBookMetadata",
    "BookMetadata",
    "turnPage",
    "direction",
    "goToPage",
    "pageNumber",
    "goToPosition",
    "position",
    "goToChapter",
    "chapterIndex",
    "searchInBook",
    "query",
    "findTextPosition",
    "text",
    "saveProgress",
    "getProgress",
    "BookProgress",
    "addBookmark",
    "name",
    "getBookmarks",
    "isFormatSupported",
    "format",
    "convertFormat",
    "inputPath",
    "outputFormat",
    "setFont",
    "font",
    "setPageSize",
    "size",
    "setMargins",
    "left",
    "top",
    "right",
    "bottom",
    "setLineSpacing",
    "spacing",
    "currentBookId",
    "currentTitle",
    "currentPage",
    "totalPages",
    "currentText",
    "isLoading"
);
#else  // !QT_MOC_HAS_STRINGDATA
#error "qtmochelpers.h not found or too old."
#endif // !QT_MOC_HAS_STRINGDATA

Q_CONSTINIT static const uint qt_meta_data_ZN10BookEngineE[] = {

 // content:
      12,       // revision
       0,       // classname
       0,    0, // classinfo
      34,   14, // methods
       7,  310, // properties
       0,    0, // enums/sets
       0,    0, // constructors
       0,       // flags
       8,       // signalCount

 // signals: name, argc, parameters, tag, flags, initial metatype offsets
       1,    0,  218,    2, 0x06,    8 /* Public */,
       3,    1,  219,    2, 0x06,    9 /* Public */,
       5,    1,  222,    2, 0x06,   11 /* Public */,
       7,    1,  225,    2, 0x06,   13 /* Public */,
       9,    0,  228,    2, 0x06,   15 /* Public */,
      10,    1,  229,    2, 0x06,   16 /* Public */,
      12,    1,  232,    2, 0x06,   18 /* Public */,
      14,    1,  235,    2, 0x06,   20 /* Public */,

 // slots: name, argc, parameters, tag, flags, initial metatype offsets
      15,    0,  238,    2, 0x08,   22 /* Private */,
      16,    0,  239,    2, 0x08,   23 /* Private */,
      17,    1,  240,    2, 0x08,   24 /* Private */,

 // methods: name, argc, parameters, tag, flags, initial metatype offsets
      18,    1,  243,    2, 0x02,   26 /* Public */,
      19,    1,  246,    2, 0x02,   28 /* Public */,
      21,    1,  249,    2, 0x02,   30 /* Public */,
      22,    0,  252,    2, 0x102,   32 /* Public | MethodIsConst  */,
      23,    1,  253,    2, 0x102,   33 /* Public | MethodIsConst  */,
      25,    1,  256,    2, 0x02,   35 /* Public */,
      25,    0,  259,    2, 0x22,   37 /* Public | MethodCloned */,
      27,    1,  260,    2, 0x02,   38 /* Public */,
      29,    1,  263,    2, 0x02,   40 /* Public */,
      31,    1,  266,    2, 0x02,   42 /* Public */,
      33,    1,  269,    2, 0x102,   44 /* Public | MethodIsConst  */,
      35,    1,  272,    2, 0x102,   46 /* Public | MethodIsConst  */,
      37,    0,  275,    2, 0x02,   48 /* Public */,
      38,    1,  276,    2, 0x102,   49 /* Public | MethodIsConst  */,
      40,    1,  279,    2, 0x02,   51 /* Public */,
      40,    0,  282,    2, 0x22,   53 /* Public | MethodCloned */,
      42,    0,  283,    2, 0x102,   54 /* Public | MethodIsConst  */,
      43,    1,  284,    2, 0x102,   55 /* Public | MethodIsConst  */,
      45,    2,  287,    2, 0x02,   57 /* Public */,
      48,    1,  292,    2, 0x02,   60 /* Public */,
      50,    1,  295,    2, 0x02,   62 /* Public */,
      52,    4,  298,    2, 0x02,   64 /* Public */,
      57,    1,  307,    2, 0x02,   69 /* Public */,

 // signals: parameters
    QMetaType::Void,
    QMetaType::Void, QMetaType::QString,    4,
    QMetaType::Void, QMetaType::Int,    6,
    QMetaType::Void, QMetaType::Float,    8,
    QMetaType::Void,
    QMetaType::Void, QMetaType::Bool,   11,
    QMetaType::Void, QMetaType::QString,   13,
    QMetaType::Void, QMetaType::QString,    4,

 // slots: parameters
    QMetaType::Void,
    QMetaType::Void,
    QMetaType::Void, QMetaType::QString,   12,

 // methods: parameters
    QMetaType::Bool, QMetaType::QString,    4,
    QMetaType::Bool, QMetaType::QString,   20,
    QMetaType::QString, QMetaType::QString,   20,
    QMetaType::QStringList,
    0x80000000 | 24, QMetaType::QString,    4,
    QMetaType::Void, QMetaType::Int,   26,
    QMetaType::Void,
    QMetaType::Void, QMetaType::Int,   28,
    QMetaType::Void, QMetaType::Int,   30,
    QMetaType::Void, QMetaType::Int,   32,
    QMetaType::QStringList, QMetaType::QString,   34,
    QMetaType::Int, QMetaType::QString,   36,
    QMetaType::Void,
    0x80000000 | 39, QMetaType::QString,    4,
    QMetaType::Void, QMetaType::QString,   41,
    QMetaType::Void,
    QMetaType::QStringList,
    QMetaType::Bool, QMetaType::QString,   44,
    QMetaType::QString, QMetaType::QString, QMetaType::QString,   46,   47,
    QMetaType::Void, QMetaType::QFont,   49,
    QMetaType::Void, QMetaType::QSize,   51,
    QMetaType::Void, QMetaType::Int, QMetaType::Int, QMetaType::Int, QMetaType::Int,   53,   54,   55,   56,
    QMetaType::Void, QMetaType::Float,   58,

 // properties: name, type, flags, notifyId, revision
      59, QMetaType::QString, 0x00015001, uint(0), 0,
      60, QMetaType::QString, 0x00015001, uint(0), 0,
      61, QMetaType::Int, 0x00015103, uint(2), 0,
      62, QMetaType::Int, 0x00015001, uint(1), 0,
       8, QMetaType::Float, 0x00015001, uint(3), 0,
      63, QMetaType::QString, 0x00015001, uint(4), 0,
      64, QMetaType::Bool, 0x00015001, uint(5), 0,

       0        // eod
};

Q_CONSTINIT const QMetaObject BookEngine::staticMetaObject = { {
    QMetaObject::SuperData::link<QObject::staticMetaObject>(),
    qt_meta_stringdata_ZN10BookEngineE.offsetsAndSizes,
    qt_meta_data_ZN10BookEngineE,
    qt_static_metacall,
    nullptr,
    qt_incomplete_metaTypeArray<qt_meta_tag_ZN10BookEngineE_t,
        // property 'currentBookId'
        QtPrivate::TypeAndForceComplete<QString, std::true_type>,
        // property 'currentTitle'
        QtPrivate::TypeAndForceComplete<QString, std::true_type>,
        // property 'currentPage'
        QtPrivate::TypeAndForceComplete<int, std::true_type>,
        // property 'totalPages'
        QtPrivate::TypeAndForceComplete<int, std::true_type>,
        // property 'progress'
        QtPrivate::TypeAndForceComplete<float, std::true_type>,
        // property 'currentText'
        QtPrivate::TypeAndForceComplete<QString, std::true_type>,
        // property 'isLoading'
        QtPrivate::TypeAndForceComplete<bool, std::true_type>,
        // Q_OBJECT / Q_GADGET
        QtPrivate::TypeAndForceComplete<BookEngine, std::true_type>,
        // method 'currentBookChanged'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        // method 'bookLoaded'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        QtPrivate::TypeAndForceComplete<const QString &, std::false_type>,
        // method 'pageChanged'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        QtPrivate::TypeAndForceComplete<int, std::false_type>,
        // method 'progressChanged'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        QtPrivate::TypeAndForceComplete<float, std::false_type>,
        // method 'textChanged'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        // method 'loadingChanged'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        QtPrivate::TypeAndForceComplete<bool, std::false_type>,
        // method 'error'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        QtPrivate::TypeAndForceComplete<const QString &, std::false_type>,
        // method 'bookAdded'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        QtPrivate::TypeAndForceComplete<const QString &, std::false_type>,
        // method 'onPaginationComplete'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        // method 'onSyncComplete'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        // method 'onError'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        QtPrivate::TypeAndForceComplete<const QString &, std::false_type>,
        // method 'loadBook'
        QtPrivate::TypeAndForceComplete<bool, std::false_type>,
        QtPrivate::TypeAndForceComplete<const QString &, std::false_type>,
        // method 'loadBookFromFile'
        QtPrivate::TypeAndForceComplete<bool, std::false_type>,
        QtPrivate::TypeAndForceComplete<const QString &, std::false_type>,
        // method 'addBookToLibrary'
        QtPrivate::TypeAndForceComplete<QString, std::false_type>,
        QtPrivate::TypeAndForceComplete<const QString &, std::false_type>,
        // method 'getLibraryBooks'
        QtPrivate::TypeAndForceComplete<QStringList, std::false_type>,
        // method 'getBookMetadata'
        QtPrivate::TypeAndForceComplete<BookMetadata, std::false_type>,
        QtPrivate::TypeAndForceComplete<const QString &, std::false_type>,
        // method 'turnPage'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        QtPrivate::TypeAndForceComplete<int, std::false_type>,
        // method 'turnPage'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        // method 'goToPage'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        QtPrivate::TypeAndForceComplete<int, std::false_type>,
        // method 'goToPosition'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        QtPrivate::TypeAndForceComplete<int, std::false_type>,
        // method 'goToChapter'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        QtPrivate::TypeAndForceComplete<int, std::false_type>,
        // method 'searchInBook'
        QtPrivate::TypeAndForceComplete<QStringList, std::false_type>,
        QtPrivate::TypeAndForceComplete<const QString &, std::false_type>,
        // method 'findTextPosition'
        QtPrivate::TypeAndForceComplete<int, std::false_type>,
        QtPrivate::TypeAndForceComplete<const QString &, std::false_type>,
        // method 'saveProgress'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        // method 'getProgress'
        QtPrivate::TypeAndForceComplete<BookProgress, std::false_type>,
        QtPrivate::TypeAndForceComplete<const QString &, std::false_type>,
        // method 'addBookmark'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        QtPrivate::TypeAndForceComplete<const QString &, std::false_type>,
        // method 'addBookmark'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        // method 'getBookmarks'
        QtPrivate::TypeAndForceComplete<QStringList, std::false_type>,
        // method 'isFormatSupported'
        QtPrivate::TypeAndForceComplete<bool, std::false_type>,
        QtPrivate::TypeAndForceComplete<const QString &, std::false_type>,
        // method 'convertFormat'
        QtPrivate::TypeAndForceComplete<QString, std::false_type>,
        QtPrivate::TypeAndForceComplete<const QString &, std::false_type>,
        QtPrivate::TypeAndForceComplete<const QString &, std::false_type>,
        // method 'setFont'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        QtPrivate::TypeAndForceComplete<const QFont &, std::false_type>,
        // method 'setPageSize'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        QtPrivate::TypeAndForceComplete<const QSize &, std::false_type>,
        // method 'setMargins'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        QtPrivate::TypeAndForceComplete<int, std::false_type>,
        QtPrivate::TypeAndForceComplete<int, std::false_type>,
        QtPrivate::TypeAndForceComplete<int, std::false_type>,
        QtPrivate::TypeAndForceComplete<int, std::false_type>,
        // method 'setLineSpacing'
        QtPrivate::TypeAndForceComplete<void, std::false_type>,
        QtPrivate::TypeAndForceComplete<float, std::false_type>
    >,
    nullptr
} };

void BookEngine::qt_static_metacall(QObject *_o, QMetaObject::Call _c, int _id, void **_a)
{
    auto *_t = static_cast<BookEngine *>(_o);
    if (_c == QMetaObject::InvokeMetaMethod) {
        switch (_id) {
        case 0: _t->currentBookChanged(); break;
        case 1: _t->bookLoaded((*reinterpret_cast< std::add_pointer_t<QString>>(_a[1]))); break;
        case 2: _t->pageChanged((*reinterpret_cast< std::add_pointer_t<int>>(_a[1]))); break;
        case 3: _t->progressChanged((*reinterpret_cast< std::add_pointer_t<float>>(_a[1]))); break;
        case 4: _t->textChanged(); break;
        case 5: _t->loadingChanged((*reinterpret_cast< std::add_pointer_t<bool>>(_a[1]))); break;
        case 6: _t->error((*reinterpret_cast< std::add_pointer_t<QString>>(_a[1]))); break;
        case 7: _t->bookAdded((*reinterpret_cast< std::add_pointer_t<QString>>(_a[1]))); break;
        case 8: _t->onPaginationComplete(); break;
        case 9: _t->onSyncComplete(); break;
        case 10: _t->onError((*reinterpret_cast< std::add_pointer_t<QString>>(_a[1]))); break;
        case 11: { bool _r = _t->loadBook((*reinterpret_cast< std::add_pointer_t<QString>>(_a[1])));
            if (_a[0]) *reinterpret_cast< bool*>(_a[0]) = std::move(_r); }  break;
        case 12: { bool _r = _t->loadBookFromFile((*reinterpret_cast< std::add_pointer_t<QString>>(_a[1])));
            if (_a[0]) *reinterpret_cast< bool*>(_a[0]) = std::move(_r); }  break;
        case 13: { QString _r = _t->addBookToLibrary((*reinterpret_cast< std::add_pointer_t<QString>>(_a[1])));
            if (_a[0]) *reinterpret_cast< QString*>(_a[0]) = std::move(_r); }  break;
        case 14: { QStringList _r = _t->getLibraryBooks();
            if (_a[0]) *reinterpret_cast< QStringList*>(_a[0]) = std::move(_r); }  break;
        case 15: { BookMetadata _r = _t->getBookMetadata((*reinterpret_cast< std::add_pointer_t<QString>>(_a[1])));
            if (_a[0]) *reinterpret_cast< BookMetadata*>(_a[0]) = std::move(_r); }  break;
        case 16: _t->turnPage((*reinterpret_cast< std::add_pointer_t<int>>(_a[1]))); break;
        case 17: _t->turnPage(); break;
        case 18: _t->goToPage((*reinterpret_cast< std::add_pointer_t<int>>(_a[1]))); break;
        case 19: _t->goToPosition((*reinterpret_cast< std::add_pointer_t<int>>(_a[1]))); break;
        case 20: _t->goToChapter((*reinterpret_cast< std::add_pointer_t<int>>(_a[1]))); break;
        case 21: { QStringList _r = _t->searchInBook((*reinterpret_cast< std::add_pointer_t<QString>>(_a[1])));
            if (_a[0]) *reinterpret_cast< QStringList*>(_a[0]) = std::move(_r); }  break;
        case 22: { int _r = _t->findTextPosition((*reinterpret_cast< std::add_pointer_t<QString>>(_a[1])));
            if (_a[0]) *reinterpret_cast< int*>(_a[0]) = std::move(_r); }  break;
        case 23: _t->saveProgress(); break;
        case 24: { BookProgress _r = _t->getProgress((*reinterpret_cast< std::add_pointer_t<QString>>(_a[1])));
            if (_a[0]) *reinterpret_cast< BookProgress*>(_a[0]) = std::move(_r); }  break;
        case 25: _t->addBookmark((*reinterpret_cast< std::add_pointer_t<QString>>(_a[1]))); break;
        case 26: _t->addBookmark(); break;
        case 27: { QStringList _r = _t->getBookmarks();
            if (_a[0]) *reinterpret_cast< QStringList*>(_a[0]) = std::move(_r); }  break;
        case 28: { bool _r = _t->isFormatSupported((*reinterpret_cast< std::add_pointer_t<QString>>(_a[1])));
            if (_a[0]) *reinterpret_cast< bool*>(_a[0]) = std::move(_r); }  break;
        case 29: { QString _r = _t->convertFormat((*reinterpret_cast< std::add_pointer_t<QString>>(_a[1])),(*reinterpret_cast< std::add_pointer_t<QString>>(_a[2])));
            if (_a[0]) *reinterpret_cast< QString*>(_a[0]) = std::move(_r); }  break;
        case 30: _t->setFont((*reinterpret_cast< std::add_pointer_t<QFont>>(_a[1]))); break;
        case 31: _t->setPageSize((*reinterpret_cast< std::add_pointer_t<QSize>>(_a[1]))); break;
        case 32: _t->setMargins((*reinterpret_cast< std::add_pointer_t<int>>(_a[1])),(*reinterpret_cast< std::add_pointer_t<int>>(_a[2])),(*reinterpret_cast< std::add_pointer_t<int>>(_a[3])),(*reinterpret_cast< std::add_pointer_t<int>>(_a[4]))); break;
        case 33: _t->setLineSpacing((*reinterpret_cast< std::add_pointer_t<float>>(_a[1]))); break;
        default: ;
        }
    }
    if (_c == QMetaObject::IndexOfMethod) {
        int *result = reinterpret_cast<int *>(_a[0]);
        {
            using _q_method_type = void (BookEngine::*)();
            if (_q_method_type _q_method = &BookEngine::currentBookChanged; *reinterpret_cast<_q_method_type *>(_a[1]) == _q_method) {
                *result = 0;
                return;
            }
        }
        {
            using _q_method_type = void (BookEngine::*)(const QString & );
            if (_q_method_type _q_method = &BookEngine::bookLoaded; *reinterpret_cast<_q_method_type *>(_a[1]) == _q_method) {
                *result = 1;
                return;
            }
        }
        {
            using _q_method_type = void (BookEngine::*)(int );
            if (_q_method_type _q_method = &BookEngine::pageChanged; *reinterpret_cast<_q_method_type *>(_a[1]) == _q_method) {
                *result = 2;
                return;
            }
        }
        {
            using _q_method_type = void (BookEngine::*)(float );
            if (_q_method_type _q_method = &BookEngine::progressChanged; *reinterpret_cast<_q_method_type *>(_a[1]) == _q_method) {
                *result = 3;
                return;
            }
        }
        {
            using _q_method_type = void (BookEngine::*)();
            if (_q_method_type _q_method = &BookEngine::textChanged; *reinterpret_cast<_q_method_type *>(_a[1]) == _q_method) {
                *result = 4;
                return;
            }
        }
        {
            using _q_method_type = void (BookEngine::*)(bool );
            if (_q_method_type _q_method = &BookEngine::loadingChanged; *reinterpret_cast<_q_method_type *>(_a[1]) == _q_method) {
                *result = 5;
                return;
            }
        }
        {
            using _q_method_type = void (BookEngine::*)(const QString & );
            if (_q_method_type _q_method = &BookEngine::error; *reinterpret_cast<_q_method_type *>(_a[1]) == _q_method) {
                *result = 6;
                return;
            }
        }
        {
            using _q_method_type = void (BookEngine::*)(const QString & );
            if (_q_method_type _q_method = &BookEngine::bookAdded; *reinterpret_cast<_q_method_type *>(_a[1]) == _q_method) {
                *result = 7;
                return;
            }
        }
    }
    if (_c == QMetaObject::ReadProperty) {
        void *_v = _a[0];
        switch (_id) {
        case 0: *reinterpret_cast< QString*>(_v) = _t->currentBookId(); break;
        case 1: *reinterpret_cast< QString*>(_v) = _t->currentTitle(); break;
        case 2: *reinterpret_cast< int*>(_v) = _t->currentPage(); break;
        case 3: *reinterpret_cast< int*>(_v) = _t->totalPages(); break;
        case 4: *reinterpret_cast< float*>(_v) = _t->progress(); break;
        case 5: *reinterpret_cast< QString*>(_v) = _t->currentText(); break;
        case 6: *reinterpret_cast< bool*>(_v) = _t->isLoading(); break;
        default: break;
        }
    }
    if (_c == QMetaObject::WriteProperty) {
        void *_v = _a[0];
        switch (_id) {
        case 2: _t->setCurrentPage(*reinterpret_cast< int*>(_v)); break;
        default: break;
        }
    }
}

const QMetaObject *BookEngine::metaObject() const
{
    return QObject::d_ptr->metaObject ? QObject::d_ptr->dynamicMetaObject() : &staticMetaObject;
}

void *BookEngine::qt_metacast(const char *_clname)
{
    if (!_clname) return nullptr;
    if (!strcmp(_clname, qt_meta_stringdata_ZN10BookEngineE.stringdata0))
        return static_cast<void*>(this);
    return QObject::qt_metacast(_clname);
}

int BookEngine::qt_metacall(QMetaObject::Call _c, int _id, void **_a)
{
    _id = QObject::qt_metacall(_c, _id, _a);
    if (_id < 0)
        return _id;
    if (_c == QMetaObject::InvokeMetaMethod) {
        if (_id < 34)
            qt_static_metacall(this, _c, _id, _a);
        _id -= 34;
    }
    if (_c == QMetaObject::RegisterMethodArgumentMetaType) {
        if (_id < 34)
            *reinterpret_cast<QMetaType *>(_a[0]) = QMetaType();
        _id -= 34;
    }
    if (_c == QMetaObject::ReadProperty || _c == QMetaObject::WriteProperty
            || _c == QMetaObject::ResetProperty || _c == QMetaObject::BindableProperty
            || _c == QMetaObject::RegisterPropertyMetaType) {
        qt_static_metacall(this, _c, _id, _a);
        _id -= 7;
    }
    return _id;
}

// SIGNAL 0
void BookEngine::currentBookChanged()
{
    QMetaObject::activate(this, &staticMetaObject, 0, nullptr);
}

// SIGNAL 1
void BookEngine::bookLoaded(const QString & _t1)
{
    void *_a[] = { nullptr, const_cast<void*>(reinterpret_cast<const void*>(std::addressof(_t1))) };
    QMetaObject::activate(this, &staticMetaObject, 1, _a);
}

// SIGNAL 2
void BookEngine::pageChanged(int _t1)
{
    void *_a[] = { nullptr, const_cast<void*>(reinterpret_cast<const void*>(std::addressof(_t1))) };
    QMetaObject::activate(this, &staticMetaObject, 2, _a);
}

// SIGNAL 3
void BookEngine::progressChanged(float _t1)
{
    void *_a[] = { nullptr, const_cast<void*>(reinterpret_cast<const void*>(std::addressof(_t1))) };
    QMetaObject::activate(this, &staticMetaObject, 3, _a);
}

// SIGNAL 4
void BookEngine::textChanged()
{
    QMetaObject::activate(this, &staticMetaObject, 4, nullptr);
}

// SIGNAL 5
void BookEngine::loadingChanged(bool _t1)
{
    void *_a[] = { nullptr, const_cast<void*>(reinterpret_cast<const void*>(std::addressof(_t1))) };
    QMetaObject::activate(this, &staticMetaObject, 5, _a);
}

// SIGNAL 6
void BookEngine::error(const QString & _t1)
{
    void *_a[] = { nullptr, const_cast<void*>(reinterpret_cast<const void*>(std::addressof(_t1))) };
    QMetaObject::activate(this, &staticMetaObject, 6, _a);
}

// SIGNAL 7
void BookEngine::bookAdded(const QString & _t1)
{
    void *_a[] = { nullptr, const_cast<void*>(reinterpret_cast<const void*>(std::addressof(_t1))) };
    QMetaObject::activate(this, &staticMetaObject, 7, _a);
}
QT_WARNING_POP
