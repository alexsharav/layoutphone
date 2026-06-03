import type { GameSettings } from "@/types/gameSettings";
import type { Player } from "@/types/player";
import type { Room } from "@/types/room";

export interface RoomResponse {
  room: Room;
  player: Player;
}

export async function createRoom(nickname: string) {
  const response = await fetch(`${import.meta.env.VITE_API_URL}/api/rooms`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ nickname }),
  });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error();
  }
  return payload as RoomResponse;
}

export async function joinRoom(code: string, nickname: string) {
  const response = await fetch(
    `${import.meta.env.VITE_API_URL}/api/rooms/${code}/join`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ nickname }),
    },
  );
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error();
  }
  return payload as RoomResponse;
}

export async function getRoom(code: string) {
  const response = await fetch(
    `${import.meta.env.VITE_API_URL}/api/rooms/${code}`,
  );
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error();
  }
  return payload as { room: Room };
}

export async function updateSettings(
  code: string,
  playerId: string,
  settings: GameSettings,
) {
  const response = await fetch(
    `${import.meta.env.VITE_API_URL}/api/rooms/${code}/settings`,
    {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ playerId, ...settings }),
    },
  );
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error();
  }
  return payload as { room: Room };
}
