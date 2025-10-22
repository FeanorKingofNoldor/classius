import QtQuick 2.15
import QtQuick.Controls 2.15

Rectangle {
    id: sageInterface
    color: "#f8f8f8"
    
    signal close()
    
    Column {
        anchors {
            fill: parent
            margins: 20
        }
        spacing: 20
        
        // Header
        Row {
            width: parent.width
            
            Text {
                text: "The Sage"
                font.pixelSize: 18
                font.bold: true
                color: "#333"
                anchors.verticalCenter: parent.verticalCenter
            }
            
            Item { width: parent.width - 200 }
            
            Button {
                text: "×"
                flat: true
                font.pixelSize: 18
                onClicked: close()
            }
        }
        
        // Chat area
        ScrollView {
            width: parent.width
            height: parent.height - 120
            
            ListView {
                id: chatList
                model: ListModel {
                    ListElement {
                        type: "sage"
                        message: "Welcome! I'm here to help you understand classical texts. Ask me anything about what you're reading."
                    }
                }
                
                delegate: Rectangle {
                    width: chatList.width
                    height: messageText.height + 20
                    color: model.type === "user" ? "#e3f2fd" : "#f5f5f5"
                    radius: 8
                    
                    Text {
                        id: messageText
                        anchors {
                            left: parent.left
                            right: parent.right
                            top: parent.top
                            margins: 10
                        }
                        text: model.message
                        wrapMode: Text.Wrap
                        font.pixelSize: 14
                        color: "#333"
                    }
                }
            }
        }
        
        // Input area
        Row {
            width: parent.width
            spacing: 10
            
            TextField {
                id: questionField
                width: parent.width - sendButton.width - 10
                placeholderText: "Ask the Sage a question..."
                
                Keys.onReturnPressed: sendQuestion()
            }
            
            Button {
                id: sendButton
                text: "Ask"
                onClicked: sendQuestion()
            }
        }
    }
    
    function sendQuestion() {
        if (questionField.text.trim() === "") return
        
        // Add user question to chat
        chatList.model.append({
            type: "user",
            message: questionField.text
        })
        
        var question = questionField.text
        questionField.text = ""
        
        // Simulate AI response (placeholder)
        Timer {
            interval: 1000
            running: true
            repeat: false
            onTriggered: {
                var responses = [
                    "That's a fascinating question about classical philosophy. Let me explain...",
                    "This passage relates to the broader themes of justice and virtue in Plato's work.",
                    "The historical context here is important to understand...",
                    "This concept has influenced Western thought for over 2000 years."
                ]
                
                chatList.model.append({
                    type: "sage",
                    message: responses[Math.floor(Math.random() * responses.length)]
                })
            }
        }
    }
}