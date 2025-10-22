import QtQuick 2.15
import QtQuick.Controls 2.15
import QtQuick.Window 2.15
import Classius 1.0

ApplicationWindow {
    id: window
    width: 600
    height: 800
    visible: true
    title: "Classius - Classical Education Reader"
    
    // E-ink optimized color scheme
    color: "#ffffff"
    
    property bool showUI: true
    property bool isDemoMode: true
    
    // Main content area
    StackView {
        id: stackView
        anchors.fill: parent
        
        // Start with reading page
        initialItem: readingPage
    }
    
    // Reading page component
    Component {
        id: readingPage
        
        ReadingPage {
            id: reading
            
            onOpenLibrary: {
                stackView.push(libraryPage)
            }
            
            onShowSage: {
                sageDrawer.open()
            }
        }
    }
    
    // Library page component  
    Component {
        id: libraryPage
        
        LibraryPage {
            onBookSelected: function(bookId) {
                console.log("Loading book:", bookId)
                bookEngine.loadBook(bookId)
                stackView.pop()
            }
            
            onBack: {
                stackView.pop()
            }
        }
    }
    
    // Sage AI drawer (slides in from right)
    Drawer {
        id: sageDrawer
        width: Math.min(400, window.width * 0.6)
        height: window.height
        edge: Qt.RightEdge
        
        background: Rectangle {
            color: "#f8f8f8"
            border.color: "#e0e0e0"
        }
        
        SageInterface {
            anchors.fill: parent
            onClose: sageDrawer.close()
        }
    }
    
    // Status bar (optional, hidden by default)
    Rectangle {
        id: statusBar
        anchors.bottom: parent.bottom
        width: parent.width
        height: 40
        color: "#f0f0f0"
        visible: false // Hidden for distraction-free reading
        
        Row {
            anchors.centerIn: parent
            spacing: 20
            
            Text {
                text: bookEngine.currentTitle
                font.pixelSize: 14
                color: "#333"
            }
            
            Text {
                text: bookEngine.currentPage + " / " + bookEngine.totalPages
                font.pixelSize: 14
                color: "#666"
            }
            
            Text {
                text: Math.round(bookEngine.progress * 100) + "%"
                font.pixelSize: 14  
                color: "#666"
            }
        }
    }
    
    // Demo mode overlay
    Rectangle {
        anchors.fill: parent
        color: "transparent"
        visible: isDemoMode
        
        Rectangle {
            anchors.top: parent.top
            width: parent.width
            height: 30
            color: "#ff6b35"
            
            Text {
                anchors.centerIn: parent
                text: "CLASSIUS PROTOTYPE - DEMO MODE"
                color: "white"
                font.bold: true
                font.pixelSize: 12
            }
        }
    }
    
    // Global key handlers
    Keys.onPressed: function(event) {
        switch(event.key) {
            case Qt.Key_Left:
            case Qt.Key_PageUp:
                bookEngine.turnPage(-1)
                event.accepted = true
                break
                
            case Qt.Key_Right:
            case Qt.Key_PageDown:
            case Qt.Key_Space:
                bookEngine.turnPage(1)
                event.accepted = true
                break
                
            case Qt.Key_Home:
                bookEngine.goToPage(1)
                event.accepted = true
                break
                
            case Qt.Key_End:
                bookEngine.goToPage(bookEngine.totalPages)
                event.accepted = true
                break
                
            case Qt.Key_F1:
                showUI = !showUI
                event.accepted = true
                break
                
            case Qt.Key_F2:
                sageDrawer.open()
                event.accepted = true
                break
                
            case Qt.Key_Escape:
                if (sageDrawer.opened) {
                    sageDrawer.close()
                } else if (stackView.depth > 1) {
                    stackView.pop()
                }
                event.accepted = true
                break
        }
    }
    
    // Connect to BookEngine signals
    Connections {
        target: bookEngine
        
        function onBookLoaded(bookId) {
            console.log("Book loaded successfully:", bookId)
        }
        
        function onError(message) {
            console.error("BookEngine error:", message)
            errorDialog.text = message
            errorDialog.open()
        }
        
        function onLoadingChanged(loading) {
            if (loading) {
                loadingOverlay.visible = true
            } else {
                loadingOverlay.visible = false
            }
        }
    }
    
    // Loading overlay
    Rectangle {
        id: loadingOverlay
        anchors.fill: parent
        color: "#80ffffff"
        visible: false
        
        BusyIndicator {
            anchors.centerIn: parent
            running: parent.visible
        }
        
        Text {
            anchors.centerIn: parent
            anchors.verticalCenterOffset: 60
            text: "Loading book..."
            font.pixelSize: 16
            color: "#333"
        }
    }
    
    // Error dialog
    Dialog {
        id: errorDialog
        anchors.centerIn: parent
        width: Math.min(400, window.width * 0.8)
        title: "Error"
        
        property alias text: errorText.text
        
        contentItem: Text {
            id: errorText
            wrapMode: Text.Wrap
            color: "#333"
        }
        
        standardButtons: Dialog.Ok
    }
    
    // Initialize demo content on startup
    Component.onCompleted: {
        console.log("Classius started - Loading demo content")
        loadDemoBook()
    }
    
    function loadDemoBook() {
        // Create a simple demo book for testing
        var demoText = `THE REPUBLIC
by Plato

BOOK I

I went down yesterday to the Piraeus with Glaucon the son of Ariston, that I might offer up my prayers to the goddess; and also because I wanted to see in what manner they would celebrate the festival, which was a new thing. I was delighted with the procession of the inhabitants; but that of the Thracians was equally, if not more, beautiful. When we had finished our prayers and viewed the spectacle, we turned in the direction of the city; and at that instant Polemarchus the son of Cephalus chanced to catch sight of us from a distance as we were starting on our way home, and told his servant to run and bid us wait for him.

The servant took hold of me by the cloak behind, and said: Polemarchus desires you to wait.

I turned round, and asked him where his master was.

There he is, said the youth, coming after you, if you will only wait.

Certainly we will, said Glaucon; and in a few minutes Polemarchus appeared, and with him Adeimantus, Glaucon's brother, and Niceratus the son of Nicias, and several others who had been at the procession.

Polemarchus said to me: I perceive, Socrates, that you and our companion are already on your way to the city.

You are not far wrong, I said.

But do you see, he rejoined, how many we are?

Of course.

And are you stronger than all these? for if not, you will have to remain where you are.

May there not be the alternative, I said, that we may persuade you to let us go?

But can you persuade us, if we refuse to listen to you? he said.

Certainly not, replied Glaucon.

Then we are not going to listen; of that you may be assured.`
        
        // For demo, simulate adding and loading a book
        if (typeof bookEngine !== 'undefined') {
            // In real implementation, this would load from a file
            console.log("Demo book ready")
        }
    }
}