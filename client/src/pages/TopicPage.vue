<template>
  <main class="topic-page">
    <section class="topic-shell">
      <div class="brand-title">
        <p>LAYOUT PHONE</p>
      </div>

      <section class="topic-card">
        <p class="timer">{{ timerText }}</p>
        <h1>придумай крутой сайт</h1>

        <label class="topic-field" for="topic">
          <textarea
            id="topic"
            v-model="topic"
            maxlength="160"
            :readonly="isFinal"
            placeholder="стартап щобы как яндекс и денег много"
          ></textarea>
        </label>

        <div class="actions-row">
          <button
            class="final-button"
            type="button"
            :class="{ locked: isFinal }"
            :disabled="!topic.trim() || isFinal"
            @click="finalizeTopic"
          >
            {{ isFinal ? "тема записана" : "я всё!" }}
          </button>
          <p v-if="gameState.error" class="page-error">{{ gameState.error }}</p>
        </div>
      </section>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { getRoom } from "@/services/roomService";
import { submitTopic } from "@/services/topicService";
import { useGameStore } from "@/state/gameState";
import { formatTime } from "@/utils/formatTime";

const props = defineProps<{ lobbyId: string }>();
const router = useRouter();
const gameState = useGameStore();
const code = computed(() => props.lobbyId.toUpperCase());

const topic = ref("");
const isFinal = ref(false);
const timerText = computed(() =>
  formatTime(
    gameState.secondsLeft || gameState.room?.settings.topicWritingTime || 0,
  ),
);

onMounted(loadTopic);

watch(topic, (value) => {
  if (isFinal.value || !gameState.player.id) return;
  gameState.topicDrafts[draftKey()] = value;
});

async function loadTopic() {
  if (!gameState.hasPlayerInRoom(code.value)) {
    router.push(`/room/${code.value}`);
    return;
  }

  const response = await getRoom(code.value);
  gameState.setRoom(response.room);
  if (!roomHasCurrentPlayer()) {
    gameState.clearPlayer();
    router.push(`/room/${code.value}`);
    return;
  }
  syncTopicFromRoom();
  gameState.connectRoomSocket(code.value);

  if (response.room.status === "coding_round") {
    router.push(`/room/${code.value}/round`);
  }
  if (response.room.status === "results") {
    router.push(`/room/${code.value}/results`);
  }
}

function syncTopicFromRoom() {
  const existingTopic = gameState.room?.topics.find(
    (item) => item.authorId === gameState.player.id,
  );
  if (existingTopic) {
    topic.value = existingTopic.text;
    isFinal.value = true;
    return;
  }

  topic.value = gameState.topicDrafts[draftKey()] || "";
  isFinal.value = false;
}

async function finalizeTopic() {
  const finalTopic = topic.value.trim();
  if (!finalTopic) return;

  try {
    const response = await submitTopic(
      code.value,
      gameState.player.id,
      finalTopic,
    );
    topic.value = finalTopic;
    isFinal.value = true;
    delete gameState.topicDrafts[draftKey()];
    gameState.error = "";
    gameState.setRoom(response.room);
  } catch (error) {
    gameState.error =
      error instanceof Error && error.message.trim()
        ? error.message
        : "Тема не улетела. Попробуй еще раз.";
  }
}

function draftKey() {
  return `${code.value}:${gameState.player.id}`;
}

function roomHasCurrentPlayer() {
  return Boolean(
    gameState.room?.players.some((player) => player.id === gameState.player.id),
  );
}
</script>

<style scoped>
.topic-page {
  min-height: 100dvh;
  display: grid;
  place-items: center;
  overflow: hidden;
  padding: clamp(3px, 3vw, 38px);
  background: linear-gradient(135deg, #c0ffaf, #f5ffbe);
}

.topic-shell {
  position: relative;
  width: min(1160px, 100%);
  min-height: min(690px, 97dvh);
  display: grid;
  place-items: center;
  padding: clamp(28px, 8vw, 96px);
  border: 7px solid #222222;
  border-radius: 40px;
  background: #ffffff;
  box-shadow: 12px 15px 0 rgba(17, 17, 17, 0.13);
}

.brand-title {
  position: absolute;
  z-index: 2;
  top: clamp(22px, 5vw, 58px);
  left: 50%;
  text-align: center;
  transform: translateX(-50%) rotate(-2deg);
}

.brand-title p {
  font-size: clamp(1.4rem, 4vw, 3rem);
  font-weight: 950;
  line-height: 1;
}

.topic-card {
  width: min(820px, 100%);
  display: grid;
  gap: clamp(18px, 3vw, 28px);
  margin-top: clamp(58px, 8vw, 96px);
  padding: clamp(24px, 5vw, 54px);
  border: 6px solid #111111;
  border-radius: 34px;
  background: #ffffff;
  box-shadow: 7px 8px 0 #11111118;
}

.timer {
  justify-self: end;
  padding: 4px 18px;
  border: 5px solid #111111;
  border-radius: 999px;
  background: #adff8a;
  font-size: clamp(1.1rem, 3vw, 2rem);
  font-weight: 950;
  line-height: 1;
}

h1 {
  font-size: clamp(2.2rem, 6vw, 4.8rem);
  font-weight: 950;
  line-height: 0.95;
}

.topic-field {
  display: grid;
  gap: 8px;
  font-size: clamp(1.1rem, 3vw, 1.7rem);
  font-weight: 950;
}

.topic-field textarea {
  width: 100%;
  min-height: clamp(150px, 28dvh, 230px);
  padding: 18px 22px;
  resize: vertical;
  border: 6px solid #111111;
  border-radius: 18px;
  background: #fffef8;
  font-size: clamp(1.1rem, 2.4vw, 1.55rem);
  font-weight: 850;
  line-height: 1.25;
  box-shadow: 4px 5px 0 rgba(17, 17, 17, 0.1);
}

.topic-field textarea:focus {
  background: #fff08f;
  outline: none;
}

.topic-field textarea[readonly] {
  background: #e8ffd8;
  cursor: not-allowed;
}

.actions-row {
  display: grid;
  gap: 12px;
}

.final-button {
  min-height: 66px;
  display: grid;
  place-items: center;
  border: 6px solid #111111;
  border-radius: 999px;
  background: #adff8a;
  cursor: pointer;
  font-size: clamp(1.25rem, 3vw, 2.2rem);
  font-weight: 950;
  line-height: 1;
  text-align: center;
  box-shadow: 4px 5px 0 rgba(17, 17, 17, 0.12);
  transition: 150ms ease;
}

.final-button:hover {
  background: #fff08f;
}

.final-button:active {
  box-shadow: 0 0 0 #11111122;
}

.final-button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.final-button.locked {
  background: #ffffac;
  opacity: 0.5;
}

.page-error {
  color: #d32f2f;
  font-size: clamp(1rem, 2vw, 1.25rem);
  font-weight: 950;
}

@media (max-width: 640px) {
  .topic-page {
    overflow: auto;
  }

  .topic-shell {
    min-height: auto;
    align-content: start;
    gap: 18px;
    border-width: 5px;
    padding: 28px 14px;
  }

  .brand-title {
    position: static;
    width: 100%;
    max-width: 100%;
    justify-self: center;
    transform: rotate(-2deg);
  }

  .brand-title p {
    font-size: clamp(1.25rem, 8vw, 2rem);
    overflow-wrap: anywhere;
  }

  .topic-card {
    margin-top: 0;
    border-width: 5px;
  }
}
</style>
