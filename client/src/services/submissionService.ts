import type { Room } from "@/types/room";

export interface CodeSubmissionRequest {
  playerId: string;
  roundNumber: number;
  topicId: string;
  previousSubmissionId: string;
  html: string;
  css: string;
  js: string;
}

export async function submitCode(code: string, request: CodeSubmissionRequest) {
  const response = await fetch(
    `${import.meta.env.VITE_API_URL}/api/rooms/${code}/submissions`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(request),
    },
  );
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error();
  }
  return payload as { room: Room };
}
