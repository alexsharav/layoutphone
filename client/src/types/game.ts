import type { Player } from "./player";
import type { Topic } from "./room";

export interface Assignment {
  playerId: string;
  roundNumber: number;
  topicId: string;
  topicText?: string;
  previousSubmissionId?: string;
  html?: string;
  css?: string;
  js?: string;
}

export interface SubmissionResult {
  id: string;
  roundNumber: number;
  player: Player;
  previousSubmissionId: string;
  html: string;
  css: string;
  js: string;
}

export interface TopicResult {
  topic: Topic;
  author: Player;
  submissions: SubmissionResult[];
}
