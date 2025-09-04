package interfaces

// WebSocketBroadcaster WebSocket广播接口
type WebSocketBroadcaster interface {
	BroadcastFeeEvent(chain interface{}, feeData interface{})
	BroadcastTransactionStatusUpdate(chain interface{}, txData interface{})
}
