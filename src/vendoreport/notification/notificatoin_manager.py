from fastapi import FastAPI, WebSocket, WebSocketDisconnect


class ConnectionManager:
    def __init__(self):
        self.active_connections: list[WebSocket] = []

    async def connect(self, websocket: WebSocket):
        await websocket.accept()
        self.active_connections.append(websocket)

    def disconnect(self, websocket: WebSocket):
        self.active_connections.remove(websocket)

    async def broadcast(self, message: str):
        for connection in self.active_connections:
            await connection.send_text(message)

# manager = ConnectionManager()
#
# @app.websocket("/ws")
# async def websocket_endpoint(websocket: WebSocket):
#     await manager.connect(websocket)
#     try:
#         while True:
#             # You can handle incoming messages from clients here if needed
#             data = await websocket.receive_text()
#             # Example: Echoing back received message
#             await manager.broadcast(f"Client said: {data}")
#     except WebSocketDisconnect:
#         manager.disconnect(websocket)
#         await manager.broadcast(f"Client disconnected")
#
# @app.post("/notify")
# async def send_notification(message: str):
#     await manager.broadcast(message)
#     return {"message": "Notifications sent"}
