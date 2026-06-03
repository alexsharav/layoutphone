export interface SocketEvent {
  type: string;
  payload: unknown;
}

export function connectToRoom(code: string, playerId: string, onEvent: (event: SocketEvent) => void) {
  const socket = new WebSocket(`${import.meta.env.VITE_WS_URL}/ws/rooms/${code}?playerId=${encodeURIComponent(playerId)}`);

  socket.addEventListener("message", (message) => {
    onEvent(JSON.parse(message.data) as SocketEvent);
  });

  return socket;
}
