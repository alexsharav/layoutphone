<template>
  <main class="results-page">
    <section class="results-shell">
      <header class="results-header">
        <p>LAYOUT PHONE</p>
        <h1>цепочка игры</h1>
      </header>

      <section v-for="result in results" :key="result.id" class="result-block">
        <div class="result-heading">
          <p>{{ result.author }}</p>
          <h2>{{ result.topic }}</h2>
        </div>

        <div class="chain-grid">
          <article
            v-for="submission in result.submissions"
            :key="submission.id"
            class="result-card"
          >
            <div class="result-title">
              <span>раунд {{ submission.round }}</span>
              <h3>{{ submission.player }}</h3>
            </div>

            <iframe
              class="result-preview"
              sandbox="allow-scripts"
              :srcdoc="submission.preview"
            ></iframe>
          </article>
        </div>
      </section>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { getResults } from "@/services/gameService";
import { useGameStore } from "@/state/gameState";
import { buildPreviewHtml } from "@/utils/buildPreviewHtml";
import type { TopicResult } from "@/types/game";

const props = defineProps<{ lobbyId: string }>();
const gameState = useGameStore();

interface ResultView {
  id: string;
  topic: string;
  author: string;
  submissions: Array<{
    id: string;
    round: number;
    player: string;
    preview: string;
  }>;
}

const results = ref<ResultView[]>([]);

onMounted(loadResults);

async function loadResults() {
  const code = props.lobbyId.toUpperCase();
  if (gameState.hasPlayerInRoom(code)) {
    gameState.connectRoomSocket(code);
  }

  const response = await getResults(code);
  gameState.results = response.results;
  results.value = response.results.map(toResultView);
}

function toResultView(result: TopicResult): ResultView {
  return {
    id: result.topic.id,
    topic: result.topic.text,
    author: `${result.author.nickname} придумал тему`,
    submissions: result.submissions.map((submission) => ({
      id: submission.id,
      round: submission.roundNumber,
      player: submission.player.nickname,
      preview: buildPreviewHtml(submission.html, submission.css, submission.js),
    })),
  };
}
</script>

<style scoped>
.results-page {
  min-height: 100dvh;
  display: grid;
  place-items: center;
  padding: clamp(3px, 3vw, 38px);
  background: linear-gradient(135deg, #c0ffaf, #f5ffbe);
}

.results-shell {
  width: min(1160px, 100%);
  min-height: min(690px, 97dvh);
  display: grid;
  grid-template-rows: auto 1fr;
  gap: clamp(20px, 4vw, 34px);
  padding: clamp(20px, 5vw, 58px);
  border: 7px solid #222222;
  border-radius: 40px;
  background:
    linear-gradient(rgba(255, 255, 255, 0.9), rgba(255, 255, 255, 0.96)),
    url("../assets/result.png");
  background-size: cover;
  background-position: center;
  box-shadow: 12px 15px 0 rgba(17, 17, 17, 0.13);
}

.results-header {
  display: grid;
  justify-items: center;
  text-align: center;
}

.results-header p {
  font-size: clamp(1.1rem, 3vw, 2.2rem);
  font-weight: 950;
  line-height: 1;
}

h1 {
  font-size: clamp(2.4rem, 7vw, 5rem);
  font-weight: 950;
  line-height: 0.9;
}

.result-block {
  display: grid;
  gap: 18px;
}

.result-heading {
  display: grid;
  gap: 6px;
}

.result-heading p {
  width: max-content;
  padding: 3px 14px;
  border: 5px solid #111111;
  border-radius: 999px;
  background: #adff8a;
  font-size: clamp(1rem, 2vw, 1.35rem);
  font-weight: 950;
  line-height: 1;
}

h2 {
  font-size: clamp(1.9rem, 5vw, 3.8rem);
  font-weight: 950;
  line-height: 0.95;
}

.chain-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: clamp(18px, 4vw, 34px);
  align-items: stretch;
}

.result-card {
  min-width: 0;
  display: grid;
  grid-template-rows: auto 1fr;
  gap: 16px;
  padding: clamp(18px, 3vw, 30px);
  border: 6px solid #111111;
  border-radius: 28px;
  background: #ffffff;
  box-shadow: 5px 6px 0 rgba(17, 17, 17, 0.1);
}

.result-title {
  display: grid;
  gap: 6px;
}

.result-title span {
  width: max-content;
  padding: 3px 14px;
  border: 5px solid #111111;
  border-radius: 999px;
  background: #fff08f;
  font-size: clamp(1rem, 2vw, 1.35rem);
  font-weight: 950;
  line-height: 1;
}

h3 {
  font-size: clamp(1.6rem, 4vw, 2.6rem);
  font-weight: 950;
  line-height: 0.95;
}

.result-preview {
  width: 100%;
  min-height: clamp(260px, 38dvh, 390px);
  border: 6px solid #111111;
  border-radius: 20px;
  background: #ffffff;
}

@media (max-width: 820px) {
  .results-page {
    overflow: auto;
  }

  .results-shell {
    min-height: auto;
  }

  .chain-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 560px) {
  .results-shell,
  .result-card,
  .result-heading p,
  .result-title span,
  .result-preview {
    border-width: 4px;
  }
}
</style>
