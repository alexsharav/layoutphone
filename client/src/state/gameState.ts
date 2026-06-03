import { defineStore } from "pinia";
import router from "@/router";
import { connectToRoom, type SocketEvent } from "@/services/websocketService";
import type { Assignment, TopicResult } from "@/types/game";
import type { Player } from "@/types/player";
import type { Room } from "@/types/room";

const PLAYER_KEY = "layout-phone-player";

interface StoredPlayer {
  id: string;
  nickname: string;
  isHost: boolean;
  roomCode: string;
}

interface CodeDraft {
  html: string;
  css: string;
  js: string;
}

interface GameState {
  room: Room | null;
  player: StoredPlayer;
  assignment: Assignment | null;
  results: TopicResult[];
  topicDrafts: Record<string, string>;
  codeDrafts: Record<string, CodeDraft>;
  secondsLeft: number;
  socket: WebSocket | null;
  socketRoomCode: string;
  connected: boolean;
  error: string;
}

export const useGameStore = defineStore("game", {
  state: (): GameState => ({
    room: null,
    player: {
      id: "",
      nickname: "",
      isHost: false,
      roomCode: "",
    },
    assignment: null,
    results: [],
    topicDrafts: {},
    codeDrafts: {},
    secondsLeft: 0,
    socket: null,
    socketRoomCode: "",
    connected: false,
    error: "",
  }),
  actions: {
    loadPlayer() {
      const raw = localStorage.getItem(PLAYER_KEY);
      if (!raw) return;
      try {
        Object.assign(this.player, JSON.parse(raw) as StoredPlayer);
      } catch {
        localStorage.removeItem(PLAYER_KEY);
      }
    },
    hasPlayerInRoom(code: string) {
      this.loadPlayer();
      return Boolean(this.player.id && this.player.roomCode === code);
    },
    setPlayer(player: Player, roomCode: string) {
      this.player.id = player.id;
      this.player.nickname = player.nickname;
      this.player.isHost = player.isHost;
      this.player.roomCode = roomCode;
      localStorage.setItem(PLAYER_KEY, JSON.stringify(this.player));
    },
    clearPlayer() {
      this.socket?.close();
      this.player.id = "";
      this.player.nickname = "";
      this.player.isHost = false;
      this.player.roomCode = "";
      this.socket = null;
      this.socketRoomCode = "";
      this.connected = false;
      localStorage.removeItem(PLAYER_KEY);
    },
    setRoom(room: Room) {
      this.room = room;
    },
    connectRoomSocket(code: string) {
      if (!this.player.id || this.player.roomCode !== code) return;
      if (this.socketRoomCode === code && (this.socket?.readyState === WebSocket.OPEN || this.socket?.readyState === WebSocket.CONNECTING)) {
        return;
      }

      this.socket?.close();
      this.socketRoomCode = code;
      const socket = connectToRoom(code, this.player.id, (event) => {
        this.handleSocketEvent(event);
      });
      this.socket = socket;
      socket.addEventListener("open", () => {
        this.connected = true;
      });
      socket.addEventListener("close", () => {
        if (this.socket !== socket) return;
        this.connected = false;
        this.socket = null;
        this.socketRoomCode = "";
      });
    },
    handleSocketEvent(event: SocketEvent) {
      if (event.type === "room_updated") {
        const room = event.payload as Room;
        this.setRoom(room);
        this.routeByStatus(room);
        return;
      }

      if (event.type === "timer_tick") {
        const payload = event.payload as { secondsLeft: number };
        this.secondsLeft = payload.secondsLeft;
        return;
      }

      if (event.type === "round_started") {
        const payload = event.payload as { assignment: Assignment };
        this.assignment = payload.assignment;
        if (this.room?.code) {
          router.push(`/room/${this.room.code}/round`);
        }
        return;
      }

      if (event.type === "game_finished") {
        const code = this.room?.code || this.player.roomCode;
        if (code) {
          router.push(`/room/${code}/results`);
        }
      }
    },
    routeByStatus(room: Room) {
      const currentPath = router.currentRoute.value.path;
      if (room.status === "writing_topics" && currentPath !== `/room/${room.code}/topic`) {
        router.push(`/room/${room.code}/topic`);
      }
      if (room.status === "coding_round" && currentPath !== `/room/${room.code}/round`) {
        router.push(`/room/${room.code}/round`);
      }
      if (room.status === "results" && currentPath !== `/room/${room.code}/results`) {
        router.push(`/room/${room.code}/results`);
      }
    },
  },
});
