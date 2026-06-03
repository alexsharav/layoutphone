<template>
  <main class="lobby-page">
    <section class="lobby-shell">
      <aside class="players-panel">
        <p class="players-count">
          {{ players.length }}/{{ settingsForm.maxPlayers }}
        </p>

        <ul class="players-list">
          <li
            v-for="player in players"
            :key="player.id"
            class="player-chip"
            :class="{ host: player.isHost }"
          >
            <span>{{ player.nickname }}</span>
          </li>
        </ul>
      </aside>

      <div class="lobby-content">
        <h1 id="lobby-title">LAYOUT PHONE</h1>

        <div class="invite-row">
          <div class="invite-link">{{ lobbyLink }}</div>

          <button
            class="copy-button"
            type="button"
            :class="{ copied: isLinkCopied }"
            @click="copyLobbyLink"
          ></button>
        </div>

        <section class="settings-panel">
          <form
            v-if="!hasCurrentPlayer"
            class="join-form"
            @submit.prevent="joinByLink"
          >
            <label class="setting-field">
              <span>твое имя</span>

              <input
                v-model.trim="joinNickname"
                type="text"
                maxlength="20"
                required
                placeholder="олег"
              />
            </label>

            <button
              class="start-button"
              type="submit"
              :disabled="joinNickname.length < 3"
            >
              войти
            </button>

            <p v-if="lobbyError" class="lobby-error">{{ lobbyError }}</p>
          </form>

          <template v-else>
            <div class="settings-grid">
              <label class="setting-field">
                <span>игроков</span>

                <input
                  v-model.number="settingsForm.maxPlayers"
                  type="number"
                  min="2"
                  max="12"
                  :disabled="!isHost"
                />
              </label>

              <label class="setting-field">
                <span>на темы, с</span>

                <input
                  v-model.number="settingsForm.topicWritingTime"
                  type="number"
                  min="30"
                  max="180"
                  step="1"
                  :disabled="!isHost"
                />
              </label>

              <label class="setting-field">
                <span>на верстку, с</span>

                <input
                  v-model.number="settingsForm.codingRoundTime"
                  type="number"
                  min="30"
                  step="10"
                  :disabled="!isHost"
                />
              </label>
            </div>

            <button
              class="rules-button"
              type="button"
              :disabled="!isHost"
              @click="saveSettings"
            >
              сохранить
            </button>
            <button
              class="start-button"
              :class="{ started: isGameStarted }"
              type="button"
              :disabled="!isHost || players.length < 2"
              @click="startGame"
            >
              стартовать
            </button>
            <p v-if="lobbyError" class="lobby-error">{{ lobbyError }}</p>
          </template>
        </section>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import router from "@/router";

import type { GameSettings } from "@/types/gameSettings";
import { useGameStore } from "@/state/gameState";
import { copyToClipboard } from "@/utils/copyToClipboard";
import { generateRoomLink } from "@/utils/generateRoomLink";
import { getRoom, joinRoom, updateSettings } from "@/services/roomService";
import { startGame as requestStartGame } from "@/services/gameService";

const props = defineProps<{ lobbyId: string }>();
const gameState = useGameStore();
const lobbyCode = computed(() => props.lobbyId.toUpperCase());
const lobbyLink = computed(() => generateRoomLink(lobbyCode.value));
const isLinkCopied = ref(false);
const isGameStarted = ref(false);
const joinNickname = ref("");
const lobbyError = ref("");

const players = computed(() => gameState.room?.players ?? []);
const hasCurrentPlayer = computed(() =>
  gameState.hasPlayerInRoom(lobbyCode.value),
);

const isHost = computed(
  () =>
    hasCurrentPlayer.value && gameState.room?.hostId === gameState.player.id,
);

const settingsForm = reactive<GameSettings>({
  maxPlayers: 6,
  topicWritingTime: 60,
  codingRoundTime: 300,
});

onMounted(loadLobby);

watch(
  () => gameState.room?.settings,
  (settings) => {
    if (!settings) return;
    settingsForm.maxPlayers = settings.maxPlayers;
    settingsForm.topicWritingTime = settings.topicWritingTime;
    settingsForm.codingRoundTime = settings.codingRoundTime;
  },
  { immediate: true, deep: true },
);

async function loadLobby() {
  lobbyError.value = "";
  try {
    const response = await getRoom(lobbyCode.value);
    gameState.setRoom(response.room);
    const rememberedPlayer = gameState.hasPlayerInRoom(lobbyCode.value);
    if (rememberedPlayer && !roomHasCurrentPlayer()) {
      gameState.clearPlayer();
      return;
    }
    if (rememberedPlayer) {
      gameState.connectRoomSocket(lobbyCode.value);
      routeCurrentRoom();
    }
  } catch (error) {
    lobbyError.value = "Ошибка загрузки";
  }
}

async function joinByLink() {
  lobbyError.value = "";

  try {
    const response = await joinRoom(lobbyCode.value, joinNickname.value);
    gameState.setRoom(response.room);
    gameState.setPlayer(response.player, response.room.code);
    gameState.connectRoomSocket(response.room.code);
    routeCurrentRoom();
  } catch (error) {
    lobbyError.value = "Ошибка присоединения";
  }
}

async function copyLobbyLink() {
  try {
    await copyToClipboard(lobbyLink.value);
    isLinkCopied.value = true;

    window.setTimeout(() => {
      isLinkCopied.value = false;
    }, 1200);
  } catch {
    isLinkCopied.value = false;
  }
}

async function saveSettings() {
  if (!gameState.player.id) return;
  try {
    const response = await updateSettings(
      lobbyCode.value,
      gameState.player.id,
      {
        maxPlayers: Number(settingsForm.maxPlayers),
        topicWritingTime: Number(settingsForm.topicWritingTime),
        codingRoundTime: Number(settingsForm.codingRoundTime),
      },
    );
    gameState.setRoom(response.room);
  } catch (error) {
    lobbyError.value = "Настройки не сохранились";
  }
}

async function startGame() {
  if (!gameState.player.id) return;
  try {
    const response = await requestStartGame(
      lobbyCode.value,
      gameState.player.id,
    );
    gameState.setRoom(response.room);
    isGameStarted.value = true;
    router.push(`/room/${response.room.code}/topic`);
  } catch (error) {
    lobbyError.value = "Игра не началась";
  }
}

function routeCurrentRoom() {
  const room = gameState.room;
  if (!room) return;

  if (room.status === "writing_topics") {
    router.push(`/room/${room.code}/topic`);
  }
  if (room.status === "coding_round") {
    router.push(`/room/${room.code}/round`);
  }
  if (room.status === "results") {
    router.push(`/room/${room.code}/results`);
  }
}

function roomHasCurrentPlayer() {
  return Boolean(
    gameState.room?.players.some((player) => player.id === gameState.player.id),
  );
}
</script>

<style scoped>
.lobby-page {
  min-height: 100dvh;
  display: grid;
  place-items: center;
  overflow: auto;
  padding: clamp(3px, 3vw, 38px);
  background: linear-gradient(135deg, #c0ffaf, #f5ffbe);
}

.lobby-shell {
  position: relative;
  width: min(1160px, 100%);
  min-height: min(690px, 97dvh);
  display: grid;
  grid-template-columns: minmax(240px, 34%) 1fr;
  gap: clamp(20px, 4vw, 54px);
  align-items: end;
  padding: clamp(58px, 8vw, 96px) clamp(20px, 6vw, 72px) clamp(32px, 5vw, 58px);
  border: 7px solid #222222;
  border-radius: 52px;
  background: rgb(255, 255, 255);
  box-shadow: 12px 15px 0 rgba(17, 17, 17, 0.13);
}

.players-panel,
.settings-panel,
.invite-row {
  border: 6px solid #111111;
  background: #ffffff;
  box-shadow: 5px 6px 0 rgba(17, 17, 17, 0.1);
}

.players-panel {
  position: relative;
  min-height: clamp(390px, 55dvh, 530px);
  display: grid;
  align-content: start;
  justify-items: center;
  padding: clamp(28px, 5vw, 46px) clamp(18px, 4vw, 32px);
  border-radius: 28px;
}

.players-count {
  margin-bottom: clamp(22px, 4vw, 38px);
  font-size: clamp(1.6rem, 3.6vw, 2.8rem);
  font-weight: 950;
  line-height: 1;
}

.players-list {
  width: 100%;
  display: grid;
  gap: clamp(18px, 3vw, 30px);
}

.player-chip {
  width: min(250px, 100%);
  min-height: 58px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  justify-self: center;
  border: 6px solid #111111;
  border-radius: 999px;
  background: #fffef8;
  color: #111111;
  font-size: clamp(1.05rem, 2.2vw, 1.65rem);
  font-weight: 900;
  line-height: 1;
  box-shadow: 4px 5px 0 rgba(17, 17, 17, 0.12);
  padding: 0 16px;
}

.player-chip.host {
  background: #adff8a;
}

.player-chip.host::after {
  content: "*";
  font-size: 1.25em;
  line-height: 1;
}

.lobby-content {
  min-width: 0;
  display: grid;
  align-content: end;
  padding-top: clamp(90px, 12vw, 142px);
}

h1 {
  position: absolute;
  top: clamp(26px, 5vw, 58px);
  left: 50%;
  z-index: 2;
  width: max-content;
  max-width: calc(100% - 48px);
  margin: 0;
  font-size: clamp(1.8rem, 4.7vw, 3.8rem);
  font-weight: 950;
  line-height: 0.9;
  letter-spacing: 0;
  text-align: center;
  transform: translateX(-50%) rotate(-2deg);
}

.invite-row {
  min-height: clamp(94px, 14vw, 132px);
  display: grid;
  grid-template-columns: 1fr minmax(86px, 130px);
  gap: clamp(12px, 2vw, 22px);
  align-items: center;
  padding: clamp(12px, 2.4vw, 22px);
  border-radius: 24px;
}

.invite-link {
  min-height: 66px;
  min-width: 0;
  display: flex;
  align-items: center;
  padding: 0 clamp(18px, 3vw, 30px);
  overflow-x: auto;
  overflow-y: hidden;
  border: 6px solid #111111;
  border-radius: 24px;
  background: #fffef8;
  font-size: clamp(0.9rem, 1.6vw, 1.25rem);
  font-weight: 950;
  line-height: 1;
  white-space: nowrap;
  scrollbar-width: thin;
  scrollbar-color: #111111 #fffef8;
}

.invite-link::-webkit-scrollbar {
  height: 8px;
}

.invite-link::-webkit-scrollbar-track {
  background: #fffef8;
}

.invite-link::-webkit-scrollbar-thumb {
  border: 2px solid #fffef8;
  border-radius: 999px;
  background: #111111;
}

.copy-button {
  position: relative;
  overflow: hidden;
  min-height: 78px;
  border: 6px solid #111111;
  border-radius: 18px;
  background: #ffffff;
  color: #111111;
  cursor: pointer;
  font-size: clamp(2.1rem, 5vw, 4rem);
  font-weight: 950;
  line-height: 1;
  text-transform: uppercase;
  box-shadow: 4px 5px 0 rgba(17, 17, 17, 0.1);
  transition: 150ms ease;
}

.copy-button::before,
.copy-button::after {
  content: "";
  position: absolute;
  width: clamp(24px, 34%, 42px);
  height: clamp(32px, 48%, 52px);
  border: 5px solid #111111;
  border-radius: 7px;
  background: #ffffff;
}

.copy-button::before {
  top: 44%;
  left: 44%;
  transform: translate(-50%, -50%);
}

.copy-button::after {
  top: 56%;
  left: 56%;
  transform: translate(-50%, -50%);
}

.copy-button:hover {
  background: #fff08f;
}

.copy-button:active {
  box-shadow: 0 0 0 #111111;
}

.copy-button.copied {
  background: #adff8a;
}

.settings-panel {
  min-width: 0;
  margin-top: clamp(16px, 3vw, 28px);
  padding: clamp(20px, 4vw, 38px);
  border-radius: 40px;
}

.join-form {
  display: grid;
  gap: clamp(14px, 3vw, 24px);
}

.settings-grid {
  display: grid;
  grid-template-columns: 1fr 1fr 1.12fr;
  gap: clamp(12px, 2.4vw, 28px);
  align-items: end;
}

.setting-field {
  min-width: 0;
  display: grid;
  gap: 6px;
  font-size: clamp(0.95rem, 1.8vw, 1.45rem);
  font-weight: 950;
  line-height: 1;
}

.setting-field span {
  min-height: 2.1em;
  display: flex;
  align-items: end;
}

.setting-field input {
  min-width: 0;
  width: 100%;
  min-height: 58px;
  padding: 0 clamp(14px, 2vw, 22px);
  border: 6px solid #111111;
  border-radius: 12px;
  background: #fffef8;
  color: #111111;
  cursor: text;
  font-size: clamp(1.15rem, 2.6vw, 2.1rem);
  font-weight: 950;
  line-height: 1;
  box-shadow: 4px 5px 0 rgba(17, 17, 17, 0.1);
}

.setting-field input:focus {
  background: #fff7a8;
}

.setting-field input:disabled {
  cursor: not-allowed;
  opacity: 0.62;
}

.rules-button,
.start-button {
  width: 100%;
  min-height: 66px;
  display: grid;
  place-items: center;
  border: 6px solid #111111;
  border-radius: 999px;
  background: #ffffff;
  color: #111111;
  cursor: pointer;
  font-size: clamp(1.35rem, 3vw, 2.65rem);
  font-weight: 950;
  line-height: 1;
  text-align: center;
  box-shadow: 4px 5px 0 rgba(17, 17, 17, 0.1);
  transition: 150ms ease;
}

.rules-button {
  margin-top: clamp(22px, 4vw, 34px);
}

.start-button {
  margin-top: clamp(12px, 2vw, 18px);
  background: #adff8a;
}

.rules-button:hover,
.start-button:hover {
  background: #fff08f;
}

.rules-button:disabled,
.start-button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.rules-button:active,
.start-button:active {
  box-shadow: 0 0 0 #111111;
}

.start-button.started {
  background: #c0ffaf;
}

.lobby-error {
  margin-top: 14px;
  color: #d32f2f;
  font-size: clamp(1rem, 2vw, 1.25rem);
  font-weight: 950;
}

@media (max-width: 900px) {
  .lobby-shell {
    grid-template-columns: 1fr;
    align-items: stretch;
    gap: 22px;
    padding: 28px clamp(14px, 4vw, 28px);
  }

  .players-panel {
    min-height: 260px;
  }

  .players-list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 14px;
  }

  .player-chip {
    min-height: 62px;
  }

  h1 {
    position: static;
    width: 100%;
    max-width: none;
    margin-bottom: 22px;
    transform: rotate(-2deg);
  }
}

@media (max-width: 640px) {
  .lobby-page {
    overflow: auto;
  }

  .lobby-shell {
    min-height: auto;
    border-width: 5px;
  }

  .players-panel,
  .settings-panel,
  .invite-row,
  .player-chip,
  .invite-link,
  .copy-button,
  .setting-field input,
  .rules-button,
  .start-button {
    border-width: 4px;
  }

  .invite-row {
    grid-template-columns: 1fr 66px;
  }

  .invite-link {
    min-height: 54px;
  }

  .copy-button {
    min-height: 60px;
  }

  .settings-grid {
    grid-template-columns: 1fr;
    gap: 14px;
  }

  .setting-field span {
    min-height: auto;
  }
}
</style>
