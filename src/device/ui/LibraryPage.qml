import QtQuick 2.15
import QtQuick.Controls 2.15

Page {
    id: libraryPage
    
    signal bookSelected(string bookId)
    signal back()
    
    background: Rectangle {
        color: "#ffffff"
    }
    
    header: Rectangle {
        height: 60
        color: "#f8f8f8"
        border.color: "#e0e0e0"
        
        Row {
            anchors {
                left: parent.left
                verticalCenter: parent.verticalCenter
                margins: 10
            }
            spacing: 10
            
            Button {
                text: "← Back"
                flat: true
                onClicked: back()
            }
            
            Text {
                text: "My Library"
                font.pixelSize: 18
                font.bold: true
                color: "#333"
                anchors.verticalCenter: parent.verticalCenter
            }
        }
    }
    
    // Book list
    ListView {
        id: booksList
        anchors.fill: parent
        anchors.margins: 20
        spacing: 10
        
        model: ListModel {
            ListElement {
                title: "The Republic"
                author: "Plato"
                bookId: "demo_republic"
                progress: 0.25
                format: "TXT"
            }
            ListElement {
                title: "Meditations"
                author: "Marcus Aurelius"
                bookId: "demo_meditations"
                progress: 0.0
                format: "TXT"
            }
            ListElement {
                title: "The Iliad"
                author: "Homer"
                bookId: "demo_iliad"
                progress: 0.67
                format: "TXT"
            }
        }
        
        delegate: Rectangle {
            width: booksList.width
            height: 80
            color: "#f9f9f9"
            border.color: "#e0e0e0"
            radius: 4
            
            MouseArea {
                anchors.fill: parent
                onClicked: bookSelected(model.bookId)
            }
            
            Row {
                anchors {
                    left: parent.left
                    verticalCenter: parent.verticalCenter
                    margins: 15
                }
                spacing: 15
                
                // Book cover placeholder
                Rectangle {
                    width: 50
                    height: 60
                    color: "#ddd"
                    radius: 2
                    
                    Text {
                        anchors.centerIn: parent
                        text: "📚"
                        font.pixelSize: 24
                    }
                }
                
                // Book info
                Column {
                    spacing: 5
                    anchors.verticalCenter: parent.verticalCenter
                    
                    Text {
                        text: model.title
                        font.pixelSize: 16
                        font.bold: true
                        color: "#333"
                    }
                    
                    Text {
                        text: "by " + model.author
                        font.pixelSize: 14
                        color: "#666"
                    }
                    
                    Row {
                        spacing: 10
                        
                        Text {
                            text: model.format
                            font.pixelSize: 12
                            color: "#888"
                        }
                        
                        Text {
                            text: Math.round(model.progress * 100) + "% read"
                            font.pixelSize: 12
                            color: "#888"
                            visible: model.progress > 0
                        }
                    }
                }
            }
            
            // Progress indicator
            Rectangle {
                anchors {
                    bottom: parent.bottom
                    left: parent.left
                }
                width: parent.width * model.progress
                height: 3
                color: "#4a90e2"
                visible: model.progress > 0
            }
        }
    }
    
    // Add book button
    RoundButton {
        anchors {
            bottom: parent.bottom
            right: parent.right
            margins: 20
        }
        text: "+"
        font.pixelSize: 24
        
        onClicked: {
            console.log("Add book clicked")
            // TODO: Show file picker
        }
    }
}