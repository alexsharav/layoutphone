import type { Assignment, TopicResult } from "@/types/game";
import type { Room } from "@/types/room";

export async function startGame(code: string, playerId: string) {
  const response = await fetch(
    `${import.meta.env.VITE_API_URL}/api/rooms/${code}/start`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ playerId }),
    },
  );
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error();
  }
  return payload as { room: Room };
}

export async function getAssignment(code: string, playerId: string) {
  const response = await fetch(
    `${import.meta.env.VITE_API_URL}/api/rooms/${code}/assignment?playerId=${encodeURIComponent(playerId)}`,
  );
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error();
  }
  return payload as { assignment: Assignment };
}

export async function getResults(code: string) {
  const response = await fetch(
    `${import.meta.env.VITE_API_URL}/api/rooms/${code}/results`,
  );
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error();
  }
  return payload as { results: TopicResult[] };
}
