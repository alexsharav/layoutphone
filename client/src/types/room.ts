import type { GameSettings } from "./gameSettings";
import type { Player } from "./player";

export type RoomStatus = "lobby" | "writing_topics" | "coding_round" | "results" | "finished";

export interface Topic {
  id: string;
  roomId: string;
  authorId: string;
  text: string;
}

export interface Submission {
  id: string;
  roomId: string;
  topicId: string;
  roundNumber: number;
  playerId: string;
  previousSubmissionId: string;
  html: string;
  css: string;
  js: string;
}

export interface Round {
  roomId: string;
  number: number;
  startedAt: number;
  endsAt: number;
}

export interface Room {
  id: string;
  code: string;
  hostId: string;
  status: RoomStatus;
  settings: GameSettings;
  players: Player[];
  topics: Topic[];
  submissions: Submission[];
  currentRound: number;
  activeRound?: Round;
}
