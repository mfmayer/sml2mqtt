module github.com/mfmayer/sml2mqtt

go 1.19

require github.com/mfmayer/gosml v0.0.1

require (
	github.com/eclipse/paho.mqtt.golang v1.4.2
	github.com/mfmayer/goham v0.0.1
)

require (
	github.com/gorilla/websocket v1.5.0 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
)

// replace github.com/mfmayer/goham => ../goham
