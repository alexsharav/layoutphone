<template>
  <main class="round-page">
    <section class="round-shell">
      <header class="round-header">
        <h1>{{ roundTopic }}</h1>
        <p>{{ timerText }}</p>
      </header>

      <section class="workspace">
        <div class="editors-panel">
          <label>
            <span>HTML</span>
            <textarea
              v-model="html"
              :readonly="isFinal"
              spellcheck="false"
            ></textarea>
          </label>

          <label>
            <span>CSS</span>
            <textarea
              v-model="css"
              :readonly="isFinal"
              spellcheck="false"
            ></textarea>
          </label>

          <label>
            <span>JS</span>
            <textarea
              v-model="js"
              :readonly="isFinal"
              spellcheck="false"
            ></textarea>
          </label>
        </div>

        <div class="preview-panel">
          <p>как вышло</p>
          <iframe
            class="preview-frame"
            sandbox="allow-scripts"
            :srcdoc="previewHtml"
          ></iframe>
          <div class="round-actions">
            <button
              type="button"
              :class="{ locked: isFinal }"
              :disabled="isFinal"
              @click="finalizeResult"
            >
              {{ isFinal ? "раунд записан" : "сдать раунд" }}
            </button>
            <p v-if="gameState.error" class="page-error">
              {{ gameState.error }}
            </p>
          </div>
        </div>
      </section>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { buildPreviewHtml } from "@/utils/buildPreviewHtml";
import { formatTime } from "@/utils/formatTime";
import { getAssignment } from "@/services/gameService";
import { submitCode } from "@/services/submissionService";
import { getRoom } from "@/services/roomService";
import { useGameStore } from "@/state/gameState";

const props = defineProps<{ lobbyId: string }>();
const router = useRouter();
const gameState = useGameStore();
const code = computed(() => props.lobbyId.toUpperCase());

const html = ref(starterHTML());
const css = ref(starterCSS());
const js = ref("");
const isFinal = ref(false);
const timerText = computed(() =>
  formatTime(
    gameState.secondsLeft || gameState.room?.settings.codingRoundTime || 0,
  ),
);
const roundTopic = computed(() => {
  const assignment = gameState.assignment;
  if (!assignment) return "тема раунда";

  const topic = gameState.room?.topics.find(
    (item) => item.id === assignment.topicId,
  );
  return assignment.topicText || topic?.text || "тема раунда";
});

onMounted(loadRound);

watch(
  () => gameState.assignment,
  (assignment) => {
    if (!assignment) return;
    applyAssignment();
  },
);

watch(
  () => gameState.room?.currentRound,
  (round, previousRound) => {
    if (!round || round === previousRound) return;
    loadRound();
  },
);

watch([html, css, js], saveDraft);

async function loadRound() {
  if (!gameState.hasPlayerInRoom(code.value)) {
    router.push(`/room/${code.value}`);
    return;
  }

  const roomResponse = await getRoom(code.value);
  gameState.setRoom(roomResponse.room);
  if (!roomHasCurrentPlayer()) {
    gameState.clearPlayer();
    router.push(`/room/${code.value}`);
    return;
  }
  gameState.connectRoomSocket(code.value);

  if (roomResponse.room.status === "results") {
    router.push(`/room/${code.value}/results`);
    return;
  }

  const assignmentResponse = await getAssignment(
    code.value,
    gameState.player.id,
  );
  gameState.assignment = assignmentResponse.assignment;
  applyAssignment();
}

function applyAssignment() {
  const assignment = gameState.assignment;
  if (!assignment) return;

  const submission = gameState.room?.submissions.find((item) => {
    return (
      item.playerId === gameState.player.id &&
      item.roundNumber === assignment.roundNumber
    );
  });
  if (submission) {
    isFinal.value = true;
    html.value = submission.html;
    css.value = submission.css;
    js.value = submission.js;
    return;
  }

  const draft = gameState.codeDrafts[draftKey()];
  isFinal.value = false;
  html.value =
    draft?.html ?? assignment.html ?? starterHTML(assignment.topicText);
  css.value = draft?.css ?? assignment.css ?? starterCSS();
  js.value = draft?.js ?? assignment.js ?? "";
}

async function finalizeResult() {
  if (!gameState.assignment || isFinal.value) return;

  try {
    const response = await submitCode(code.value, {
      playerId: gameState.player.id,
      roundNumber: gameState.assignment.roundNumber,
      topicId: gameState.assignment.topicId,
      previousSubmissionId: gameState.assignment.previousSubmissionId || "",
      html: html.value,
      css: css.value,
      js: js.value,
    });
    delete gameState.codeDrafts[draftKey()];
    isFinal.value = true;
    gameState.error = "";
    gameState.setRoom(response.room);
  } catch (error) {
    gameState.error =
      error instanceof Error && error.message.trim()
        ? error.message
        : "Код не отправился. Попробуй еще раз.";
  }
}

function saveDraft() {
  if (isFinal.value || !gameState.assignment) return;

  gameState.codeDrafts[draftKey()] = {
    html: html.value,
    css: css.value,
    js: js.value,
  };
}

function draftKey() {
  const assignment = gameState.assignment;
  if (!assignment) return `${code.value}:${gameState.player.id}:0`;
  return `${code.value}:${gameState.player.id}:${assignment.roundNumber}:${assignment.topicId}`;
}

function roomHasCurrentPlayer() {
  return Boolean(
    gameState.room?.players.some((player) => player.id === gameState.player.id),
  );
}

function starterHTML(topic = "яндекс новый)") {
  return `<main class="page">
  <p class="eyebrow">яндекс</p>
  <h1>${topic}</h1>
  <p>мы яндекс компания круто</button>
</main>`;
}

function starterCSS() {
  return `body {
  background: #f7ffd7;
  color: #111;
}
.page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  align-content: center;
  gap: 18px;
  padding: 42px;
  text-align: center;
}
.eyebrow {
  font-weight: 900;
  text-transform: uppercase;
}
h1 {
  max-width: 720px;
  font-size: clamp(36px, 7vw, 72px);
  line-height: .95;
}
button {
  border: 5px solid #111;
  border-radius: 999px;
  background: #adff8a;
  padding: 14px 28px;
  font-weight: 900;
}`;
}

const previewHtml = computed(() =>
  buildPreviewHtml(html.value, css.value, js.value),
);
</script>

<style scoped>
.round-page {
  min-height: 100dvh;
  display: grid;
  place-items: center;
  padding: clamp(3px, 3vw, 38px);
  background: linear-gradient(135deg, #c0ffaf, #f5ffbe);
}

.round-shell {
  position: relative;
  width: min(1240px, 100%);
  min-height: min(740px, 97dvh);
  display: grid;
  grid-template-rows: auto 1fr;
  gap: clamp(14px, 2.2vw, 24px);
  padding: clamp(18px, 4vw, 44px);
  border: 7px solid #222222;
  border-radius: 40px;
  background: #ffffff;
  box-shadow: 12px 15px 0 rgba(17, 17, 17, 0.13);
}

.round-header {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 16px;
  align-items: center;
}

.round-header p,
.editors-panel,
.preview-panel {
  border: 6px solid #111111;
  background: #ffffff;
  box-shadow: 5px 6px 0 rgba(17, 17, 17, 0.1);
}

.round-header p {
  min-height: 58px;
  display: grid;
  place-items: center;
  border-radius: 999px;
  font-size: clamp(1rem, 2.4vw, 1.6rem);
  font-weight: 950;
}

.round-header p {
  background: #adff8a;
}

h1 {
  text-align: left;
  font-size: clamp(2rem, 5vw, 4rem);
  font-weight: 950;
  line-height: 0.9;
}

.preview-panel p,
.editors-panel span {
  font-size: clamp(1rem, 2vw, 1.35rem);
  font-weight: 950;
  text-transform: uppercase;
}

.workspace {
  min-height: 0;
  display: grid;
  grid-template-columns: minmax(280px, 0.42fr) minmax(380px, 0.58fr);
  gap: clamp(16px, 3vw, 28px);
}

.editors-panel {
  min-height: 0;
  display: grid;
  gap: 14px;
  padding: clamp(16px, 3vw, 26px);
  border-radius: 28px;
}

.editors-panel label {
  min-height: 0;
  display: grid;
  grid-template-rows: auto 1fr;
  gap: 8px;
}

.editors-panel textarea {
  width: 100%;
  min-height: 190px;
  resize: vertical;
  padding: 12px 14px;
  border: 5px solid #111111;
  border-radius: 18px;
  background: #fffef8;
  font-family: "SFMono-Regular", Consolas, monospace;
  font-size: 0.9rem;
  font-weight: 800;
  line-height: 1.45;
}

.editors-panel textarea:focus {
  background: #fff08f;
  outline: none;
}

.editors-panel textarea[readonly] {
  background: #e8ffd8;
  cursor: not-allowed;
}

.preview-panel {
  min-width: 0;
  display: grid;
  grid-template-rows: auto minmax(280px, 1fr) auto;
  gap: 14px;
  padding: clamp(16px, 3vw, 26px);
  border-radius: 28px;
}

.preview-frame {
  width: 100%;
  height: 100%;
  min-height: 340px;
  border: 6px solid #111111;
  border-radius: 20px;
  background: #ffffff;
}

.round-actions {
  display: grid;
  gap: 12px;
}

button {
  min-height: 62px;
  display: grid;
  place-items: center;
  border: 6px solid #111111;
  border-radius: 999px;
  background: #adff8a;
  cursor: pointer;
  font-size: clamp(1.1rem, 2.7vw, 1.9rem);
  font-weight: 950;
  box-shadow: 4px 5px 0 rgba(17, 17, 17, 0.12);
  transition: 150ms ease;
}

button:hover {
  background: #fff08f;
}

button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

button.locked {
  background: #ffffac;
}

.page-error {
  color: #d32f2f;
  font-size: clamp(1rem, 2vw, 1.25rem);
  font-weight: 950;
}

@media (max-width: 900px) {
  .round-page {
    overflow: auto;
  }

  .round-shell {
    min-height: auto;
  }

  .round-header,
  .workspace {
    grid-template-columns: 1fr;
  }

  .preview-panel {
    grid-template-rows: auto auto auto;
  }
}

@media (max-width: 560px) {
  .round-shell,
  .round-header p,
  .editors-panel,
  .preview-panel,
  .preview-frame,
  button {
    border-width: 4px;
  }
}
</style>
