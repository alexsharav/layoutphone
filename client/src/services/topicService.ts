import type { Room } from "@/types/room";

export async function submitTopic(
  code: string,
  playerId: string,
  text: string,
) {
  const response = await fetch(
    `${import.meta.env.VITE_API_URL}/api/rooms/${code}/topics`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ playerId, text }),
    },
  );
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error();
  }
  return payload as { room: Room };
}
