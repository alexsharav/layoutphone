<template>
  <main class="home-page">
    <section class="host-container">
      <a class="icon-link github-link" href="https://github.com/alexsharav">
        <svg viewBox="0 0 24 24">
          <path
            d="M12 2.2a10 10 0 0 0-3.16 19.48c.5.09.68-.22.68-.48v-1.78c-2.76.6-3.34-1.18-3.34-1.18a2.63 2.63 0 0 0-1.1-1.45c-.9-.62.07-.61.07-.61a2.08 2.08 0 0 1 1.52 1.02 2.11 2.11 0 0 0 2.88.82 2.12 2.12 0 0 1 .63-1.33c-2.2-.25-4.52-1.1-4.52-4.9a3.84 3.84 0 0 1 1.02-2.66 3.57 3.57 0 0 1 .1-2.62s.84-.27 2.75 1.02a9.45 9.45 0 0 1 5 0c1.9-1.29 2.74-1.02 2.74-1.02.54 1.35.2 2.35.1 2.62a3.82 3.82 0 0 1 1.02 2.67c0 3.8-2.32 4.64-4.53 4.89a2.38 2.38 0 0 1 .67 1.85v2.74c0 .26.18.58.69.48A10 10 0 0 0 12 2.2Z"
          />
        </svg>
      </a>

      <a class="icon-link contact-link" href="https://t.me/aleksandr_93839">
        <svg viewBox="0 0 24 24">
          <path
            d="M21.66 3.77 18.4 20.01c-.24 1.13-.9 1.4-1.82.87l-5.03-3.8-2.43 2.4c-.27.28-.5.5-1.02.5l.37-5.22 9.43-8.73c.41-.37-.09-.58-.64-.21L5.6 13.36.57 11.76c-1.1-.36-1.12-1.12.23-1.66L20.42 2.35c.9-.35 1.7.22 1.24 1.42Z"
          />
        </svg>
      </a>

      <div class="phone-layout">
        <p>LAYOUT PHONE</p>
      </div>

      <section class="host-card">
        <form class="host-form" @submit.prevent="createLobby">
          <h1 id="host-title">СЫГРАЕМ?</h1>

          <label for="nickname">твое имя</label>
          <input
            id="nickname"
            v-model.trim="nickname"
            type="text"
            name="nickname"
            maxlength="20"
            required
            placeholder="олег"
          />

          <button type="submit" :disabled="nickname.length < 3">
            <span>общий сбор</span>
          </button>

          <p v-if="createLobbyError" class="status error">
            {{ createLobbyError }}
          </p>
        </form>
      </section>
    </section>
  </main>
</template>

<script setup lang="ts">
import { ref } from "vue";
import router from "@/router";
import { createRoom } from "@/services/roomService";
import { useGameStore } from "@/state/gameState";

const gameState = useGameStore();
const nickname = ref("");
const createLobbyError = ref("");

async function createLobby() {
  try {
    const response = await createRoom(nickname.value);
    gameState.setPlayer(response.player, response.room.code);
    router.push(`/room/${response.room.code}`);
  } catch (error) {
    createLobbyError.value = "Не удалось создать комнату";
  }
}
</script>

<style scoped>
.home-page {
  min-height: 100dvh;
  display: grid;
  place-items: center;
  overflow: hidden;
  padding: clamp(3px, 3vw, 38px);
  background: linear-gradient(135deg, #c0ffaf, #f5ffbe);
}

.host-container {
  position: relative;
  width: min(1160px, 100%);
  min-height: min(690px, 97dvh);
  display: grid;
  place-items: center;
  padding: clamp(18px, 8vw, 130px);
  border: 7px solid #222222;
  border-radius: 9px;
  background: rgb(255, 255, 255);
  box-shadow: 12px 15px 0 rgba(17, 17, 17, 0.13);
}

.icon-link {
  position: absolute;
  z-index: 999;
  width: clamp(54px, 7vw, 86px);
  aspect-ratio: 1;
  display: grid;
  place-items: center;
  border: 5px solid #111;
  border-radius: 7px;
  background: #fff;
  color: #111;
  transform: rotate(-5deg);
  transition: 150ms ease;
  box-shadow: 4px 5px 0 rgba(17, 17, 17, 0.1);
}

.icon-link:hover {
  transform: rotate(3deg) scale(1.05);
  background: #ffec75;
}

.icon-link svg {
  width: 64%;
}

.github-link {
  top: clamp(18px, 4vw, 58px);
  left: clamp(18px, 5vw, 72px);
}

.contact-link {
  top: clamp(18px, 4vw, 58px);
  right: clamp(18px, 5vw, 72px);
}

.phone-layout {
  position: absolute;
  top: clamp(28px, 5vw, 64px);
  left: 50%;
  transform: translateX(-50%) rotate(1deg);
  font-size: clamp(1.25rem, 3vw, 2.2rem);
  font-weight: 900;
}

.host-card {
  text-align: center;
  position: relative;
  z-index: 94;
  width: min(700px, 100%);
  display: grid;
  align-items: center;
  padding: clamp(28px, 5vw, 58px);
  border: 6px solid rgb(17, 17, 17);
  border-radius: 34px;
  background: white;
  box-shadow: 4px 5px 0 rgba(17, 17, 17, 0.1);
}

.host-form {
  min-width: 0;
}

h1 {
  font-size: clamp(2.2rem, 6vw, 4.6rem);
  font-weight: 950;
}

label {
  margin-bottom: 6px;
  font-size: clamp(1.35rem, 3vw, 2.2rem);
  font-weight: 900;
}

input {
  width: 100%;
  height: 64px;
  padding: 0 22px;
  border: 6px solid #111;
  background: #fffef8;
  font-size: clamp(1rem, 3vw, 1.45rem);
  font-weight: 800;
  box-shadow: 4px 5px 0 rgba(17, 17, 17, 0.1);
}

input:focus {
  background: #fff7a8;
}

button {
  width: min(310px, 100%);
  min-height: 70px;
  display: inline-grid;
  place-items: center;
  margin-top: 34px;
  border: 6px solid #141414;
  border-radius: 20px;
  background: #adff8a;
  color: #1e1e1e;
  cursor: pointer;
  font-size: clamp(1.2rem, 3vw, 1.8rem);
  font-weight: 950;
  box-shadow: 7px 8px 0 #111;
  transform: rotate(-1deg);
  transition: 150ms ease;
}

button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

button:hover {
  background: #fff08f;
}

button:active {
  transform: rotate(1deg);
  box-shadow: 0 0 0 #111;
}

.status {
  margin-top: 18px;
  font-size: clamp(1rem, 2.5vw, 1.35rem);
  font-weight: 900;
}

.error {
  color: #d32f2f;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 820px) {
  .host-card {
    justify-items: center;
    min-height: 0;
    padding: 28px 22px 32px;
    text-align: center;
  }

  .host-form {
    width: 100%;
  }

  input {
    height: 58px;
    padding-inline: 16px;
  }
}

@media (max-width: 520px) {
  .icon-link {
    width: 52px;
    border-width: 4px;
  }

  .github-link {
    left: 18px;
  }

  .contact-link {
    right: 18px;
  }

  .phone-layout {
    top: 36px;
    font-size: 1.04rem;
  }

  .host-card {
    border-width: 5px;
  }

  h1 {
    margin-bottom: 20px;
  }

  button {
    min-height: 62px;
    box-shadow: 5px 6px 0 #111;
  }
}
</style>
