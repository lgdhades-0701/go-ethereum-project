import QtQuick 2.0
import QtWebKit 3.0
import QtWebKit.experimental 1.0
import QtQuick.Controls 1.0;
import QtQuick.Layouts 1.0;
import QtQuick.Window 2.1;
import Ethereum 1.0

ApplicationWindow {
        id: window
        title: "Ethereum"
        width: 900
        height: 600
        minimumHeight: 300

        property alias url: webview.url
        property alias webView: webview

        Item {
                objectName: "root"
                id: root
                anchors.fill: parent
                state: "inspectorShown"

                WebView {
                        objectName: "webView"
                        id: webview
                        anchors {
                                left: parent.left
                                right: parent.right
                                bottom: sizeGrip.top
                                top: parent.top
                        }

                        onTitleChanged: { window.title = title }
                        experimental.preferences.javascriptEnabled: true
                        experimental.preferences.navigatorQtObjectEnabled: true
                        experimental.preferences.developerExtrasEnabled: true
                        experimental.userScripts: [ui.assetPath("ethereum.js")]
                        experimental.onMessageReceived: {
                                console.log("[onMessageReceived]: ", message.data)
                                var data = JSON.parse(message.data)

				switch(data.call) {
				case "getBlockByNumber":
					var block = eth.getBlock("b9b56cf6f907fbee21db0cd7cbc0e6fea2fe29503a3943e275c5e467d649cb06")	
					postData(data._seed, block)
					break
				case "getBlockByHash":
					var block = eth.getBlock("b9b56cf6f907fbee21db0cd7cbc0e6fea2fe29503a3943e275c5e467d649cb06")	
					postData(data._seed, block)
					break
				case "createTx":
					if(data.args.length < 5) {
						postData(data._seed, null)
					} else {
						var tx = eth.createTx(data.args[0], data.args[1],data.args[2],data.args[3],data.args[4])
						postData(data._seed, tx)
					}
				}
                        }
			function postData(seed, data) {
				webview.experimental.postMessage(JSON.stringify({data: data, _seed: seed}))
			}
                }

                Rectangle {
                        id: sizeGrip
                        color: "gray"
                        visible: true
                        height: 10
                        anchors {
                                left: root.left
                                right: root.right
                        }
                        y: Math.round(root.height * 2 / 3)

                        MouseArea {
                                anchors.fill: parent
                                drag.target: sizeGrip
                                drag.minimumY: 0
                                drag.maximumY: root.height
                                drag.axis: Drag.YAxis
                        }
                }

                WebView {
                        id: inspector
                        visible: true
                        url: webview.experimental.remoteInspectorUrl
                        anchors {
                                left: root.left
                                right: root.right
                                top: sizeGrip.bottom
                                bottom: root.bottom
                        }
                }

                states: [
                        State {
                                name: "inspectorShown"
                                PropertyChanges {
                                        target: inspector
                                        url: webview.experimental.remoteInspectorUrl
                                }
                        }
                ]
        }
}
