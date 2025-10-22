import QtQuick 2.15
import QtQuick.Controls 2.15

Page {
    id: readingPage
    
    // Signals for navigation
    signal openLibrary()
    signal showSage()
    
    background: Rectangle {
        color: "#ffffff"
    }
    
    // Main reading area
    Rectangle {
        id: readingArea
        anchors.fill: parent
        color: "transparent"
        
        // Book content display
        ScrollView {
            id: scrollView
            anchors {
                fill: parent
                margins: 20
            }
            
            // Hide scrollbars for clean e-ink look
            ScrollBar.horizontal.policy: ScrollBar.AlwaysOff
            ScrollBar.vertical.policy: ScrollBar.AlwaysOff
            
            // Main text display
            Text {
                id: bookText
                width: scrollView.width - 40 // Account for margins
                
                // Bind to BookEngine
                text: typeof bookEngine !== 'undefined' ? bookEngine.currentText : demoText
                
                // E-ink optimized typography
                font {
                    family: "Liberation Serif"
                    pixelSize: 16
                    weight: Font.Normal
                }
                
                color: "#1a1a1a"
                lineHeight: 1.4
                wrapMode: Text.Wrap
                textFormat: Text.PlainText
                
                // Selection and interaction (manual handling via MouseArea)
                // selectByMouse: true           // Not supported in Qt6
                // selectedTextColor: "#ffffff"  // Not supported in Qt6
                // selectionColor: "#4a90e2"     // Not supported in Qt6
            }
        }
        
        // Touch areas for page turning
        MouseArea {
            id: leftTouchArea
            anchors {
                left: parent.left
                top: parent.top
                bottom: parent.bottom
            }
            width: parent.width * 0.25
            
            onClicked: {
                console.log("Previous page")
                if (typeof bookEngine !== 'undefined') {
                    bookEngine.turnPage(-1)
                }
            }
            
            // Visual feedback (optional)
            Rectangle {
                anchors.fill: parent
                color: "#4a90e2"
                opacity: parent.pressed ? 0.1 : 0.0
                Behavior on opacity { NumberAnimation { duration: 100 } }
            }
        }
        
        MouseArea {
            id: rightTouchArea
            anchors {
                right: parent.right
                top: parent.top
                bottom: parent.bottom
            }
            width: parent.width * 0.25
            
            onClicked: {
                console.log("Next page")
                if (typeof bookEngine !== 'undefined') {
                    bookEngine.turnPage(1)
                }
            }
            
            // Visual feedback (optional)
            Rectangle {
                anchors.fill: parent
                color: "#4a90e2"
                opacity: parent.pressed ? 0.1 : 0.0
                Behavior on opacity { NumberAnimation { duration: 100 } }
            }
        }
        
        // Center tap area for UI toggle
        MouseArea {
            id: centerTouchArea
            anchors {
                left: leftTouchArea.right
                right: rightTouchArea.left
                top: parent.top
                bottom: parent.bottom
            }
            
            onClicked: {
                console.log("Toggle UI")
                topBar.visible = !topBar.visible
                bottomBar.visible = !bottomBar.visible
            }
        }
        
        // Selection handling for highlighting  
        MouseArea {
            anchors.fill: scrollView
            acceptedButtons: Qt.LeftButton
            
            property int startPosition: -1
            property int endPosition: -1
            
            onPressed: function(mouse) {
                // Start text selection
                startPosition = bookText.positionAt(mouse.x, mouse.y)
            }
            
            onReleased: function(mouse) {
                endPosition = bookText.positionAt(mouse.x, mouse.y)
                
                if (startPosition >= 0 && endPosition >= 0 && startPosition !== endPosition) {
                    var selectedText = bookText.text.substring(
                        Math.min(startPosition, endPosition),
                        Math.max(startPosition, endPosition)
                    )
                    
                    console.log("Selected text:", selectedText)
                    showAnnotationMenu(selectedText, Math.min(startPosition, endPosition), Math.max(startPosition, endPosition))
                }
            }
        }
    }
    
    // Top bar (hideable)
    Rectangle {
        id: topBar
        anchors {
            top: parent.top
            left: parent.left
            right: parent.right
        }
        height: 50
        color: "#f8f8f8"
        border.color: "#e0e0e0"
        border.width: 1
        visible: false
        
        Row {
            anchors {
                left: parent.left
                verticalCenter: parent.verticalCenter
                margins: 10
            }
            spacing: 10
            
            Button {
                text: "Library"
                flat: true
                onClicked: openLibrary()
            }
            
            Button {
                text: "Sage"
                flat: true  
                onClicked: showSage()
            }
        }
        
        // Book title (center)
        Text {
            anchors.centerIn: parent
            text: typeof bookEngine !== 'undefined' ? bookEngine.currentTitle : "The Republic"
            font.pixelSize: 14
            font.bold: true
            color: "#333"
            elide: Text.ElideRight
        }
        
        // Settings button
        Button {
            anchors {
                right: parent.right
                verticalCenter: parent.verticalCenter
                margins: 10
            }
            text: "⚙"
            flat: true
            font.pixelSize: 16
            
            onClicked: {
                // TODO: Show settings menu
                console.log("Settings clicked")
            }
        }
    }
    
    // Bottom bar (hideable)
    Rectangle {
        id: bottomBar
        anchors {
            bottom: parent.bottom
            left: parent.left
            right: parent.right
        }
        height: 50
        color: "#f8f8f8"
        border.color: "#e0e0e0"
        border.width: 1
        visible: false
        
        // Progress bar
        Rectangle {
            id: progressBackground
            anchors {
                left: parent.left
                right: pageInfo.left
                verticalCenter: parent.verticalCenter
                margins: 10
            }
            height: 4
            color: "#e0e0e0"
            radius: 2
            
            Rectangle {
                id: progressFill
                anchors {
                    left: parent.left
                    top: parent.top
                    bottom: parent.bottom
                }
                width: parent.width * (typeof bookEngine !== 'undefined' ? bookEngine.progress : 0.25)
                color: "#4a90e2"
                radius: 2
                
                Behavior on width {
                    NumberAnimation { duration: 200 }
                }
            }
        }
        
        // Page info
        Text {
            id: pageInfo
            anchors {
                right: parent.right
                verticalCenter: parent.verticalCenter
                margins: 10
            }
            text: typeof bookEngine !== 'undefined' ? 
                  bookEngine.currentPage + " / " + bookEngine.totalPages :
                  "1 / 42"
            font.pixelSize: 12
            color: "#666"
        }
    }
    
    // Annotation menu (shows on text selection)
    Popup {
        id: annotationMenu
        width: 250
        height: 180
        modal: false
        
        property string selectedText: ""
        property int startPosition: -1
        property int endPosition: -1
        property int pageNumber: 1
        
        background: Rectangle {
            color: "#ffffff"
            border.color: "#ccc"
            border.width: 1
            radius: 4
            
            // Shadow effect
            Rectangle {
                anchors.fill: parent
                anchors.margins: -2
                color: "#00000030"
                radius: 6
                z: -1
            }
        }
        
        Column {
            anchors {
                centerIn: parent
                margins: 10
            }
            spacing: 8
            width: parent.width - 20
            
            // Selection preview
            Text {
                width: parent.width
                text: '"' + (selectedText.length > 50 ? selectedText.substring(0, 50) + '...' : selectedText) + '"'
                font.pixelSize: 12
                color: "#666"
                wrapMode: Text.Wrap
                horizontalAlignment: Text.AlignHCenter
                font.italic: true
            }
            
            Rectangle {
                width: parent.width
                height: 1
                color: "#e0e0e0"
            }
            
            Row {
                anchors.horizontalCenter: parent.horizontalCenter
                spacing: 8
                
                Button {
                    text: "🖍️ Highlight"
                    font.pixelSize: 12
                    onClicked: {
                        if (typeof annotationManager !== 'undefined') {
                            var bookId = typeof bookEngine !== 'undefined' ? bookEngine.currentBookId : "demo_book"
                            var highlightId = annotationManager.createHighlight(
                                bookId, 
                                pageNumber,
                                startPosition,
                                endPosition,
                                selectedText,
                                "#ffff00"  // Yellow highlight
                            )
                            console.log("Created highlight:", highlightId)
                        }
                        annotationMenu.close()
                    }
                }
                
                Button {
                    text: "📝 Note"
                    font.pixelSize: 12
                    onClicked: {
                        noteDialog.selectedText = selectedText
                        noteDialog.startPosition = startPosition
                        noteDialog.endPosition = endPosition
                        noteDialog.pageNumber = pageNumber
                        annotationMenu.close()
                        noteDialog.open()
                    }
                }
            }
            
            Row {
                anchors.horizontalCenter: parent.horizontalCenter
                spacing: 8
                
                Button {
                    text: "🔖 Bookmark"
                    font.pixelSize: 12
                    onClicked: {
                        if (typeof annotationManager !== 'undefined') {
                            var bookId = typeof bookEngine !== 'undefined' ? bookEngine.currentBookId : "demo_book"
                            var bookmarkId = annotationManager.createBookmark(bookId, pageNumber)
                            console.log("Created bookmark:", bookmarkId)
                        }
                        annotationMenu.close()
                    }
                }
                
                Button {
                    text: "🧙 Ask Sage"
                    font.pixelSize: 12
                    onClicked: {
                        console.log("Asking Sage about:", selectedText)
                        showSage()
                        annotationMenu.close()
                    }
                }
            }
        }
    }
    
    // Note creation dialog
    Dialog {
        id: noteDialog
        anchors.centerIn: parent
        width: Math.min(400, parent.width * 0.9)
        height: Math.min(300, parent.height * 0.7)
        title: "Add Note"
        
        property string selectedText: ""
        property int startPosition: -1
        property int endPosition: -1
        property int pageNumber: 1
        
        Column {
            anchors.fill: parent
            spacing: 10
            
            // Selected text preview
            Text {
                width: parent.width
                text: "Selected text:"
                font.bold: true
                font.pixelSize: 12
            }
            
            Rectangle {
                width: parent.width
                height: 60
                color: "#f5f5f5"
                border.color: "#ddd"
                radius: 4
                
                ScrollView {
                    anchors.fill: parent
                    anchors.margins: 5
                    
                    Text {
                        text: '"' + selectedText + '"'
                        font.pixelSize: 11
                        font.italic: true
                        color: "#666"
                        wrapMode: Text.Wrap
                        width: parent.width
                    }
                }
            }
            
            // Note input
            Text {
                text: "Your note:"
                font.bold: true
                font.pixelSize: 12
            }
            
            ScrollView {
                width: parent.width
                height: 120
                
                TextArea {
                    id: noteTextArea
                    placeholderText: "Enter your thoughts about this passage..."
                    wrapMode: TextArea.Wrap
                    font.pixelSize: 12
                    selectByMouse: true
                }
            }
        }
        
        standardButtons: Dialog.Ok | Dialog.Cancel
        
        onAccepted: {
            if (noteTextArea.text.trim() !== "" && typeof annotationManager !== 'undefined') {
                var bookId = typeof bookEngine !== 'undefined' ? bookEngine.currentBookId : "demo_book"
                var noteId = annotationManager.createNote(
                    bookId,
                    pageNumber,
                    startPosition,
                    endPosition,
                    selectedText,
                    noteTextArea.text.trim()
                )
                console.log("Created note:", noteId)
            }
            noteTextArea.text = ""
        }
        
        onRejected: {
            noteTextArea.text = ""
        }
    }
    
    // Demo text (for when BookEngine is not available)
    property string demoText: `THE REPUBLIC
by Plato

BOOK I

I went down yesterday to the Piraeus with Glaucon the son of Ariston, that I might offer up my prayers to the goddess; and also because I wanted to see in what manner they would celebrate the festival, which was a new thing. I was delighted with the procession of the inhabitants; but that of the Thracians was equally, if not more, beautiful. 

When we had finished our prayers and viewed the spectacle, we turned in the direction of the city; and at that instant Polemarchus the son of Cephalus chanced to catch sight of us from a distance as we were starting on our way home, and told his servant to run and bid us wait for him.

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
    
    function showAnnotationMenu(selectedText, startPos, endPos) {
        annotationMenu.selectedText = selectedText
        annotationMenu.startPosition = startPos
        annotationMenu.endPosition = endPos
        annotationMenu.pageNumber = typeof bookEngine !== 'undefined' ? bookEngine.currentPage : 1
        annotationMenu.open()
    }
    
    // Auto-hide bars after a delay
    Timer {
        id: hideTimer
        interval: 3000
        running: topBar.visible || bottomBar.visible
        onTriggered: {
            topBar.visible = false
            bottomBar.visible = false
        }
    }
}