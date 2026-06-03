<template>
  <main class="home-page">
    <section class="host-container">
      <div class="phone-layout">
        <p>LAYOUT PHONE</p>
      </div>

      <h1>Ты зашел не туда</h1>
      <RouterLink to="/" class="home-link">Вернуться на главную</RouterLink>
    </section>
  </main>
</template>

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
  align-content: center;
  gap: clamp(20px, 4vw, 34px);
  padding: clamp(18px, 8vw, 130px);
  border: 7px solid #222222;
  border-radius: 9px;
  background: rgb(255, 255, 255);
  box-shadow: 12px 15px 0 rgba(17, 17, 17, 0.13);
}

.phone-layout {
  position: absolute;
  top: clamp(28px, 5vw, 64px);
  left: 50%;
  transform: translateX(-50%) rotate(1deg);
  font-size: clamp(1.25rem, 3vw, 2.2rem);
  font-weight: 900;
}

h1 {
  max-width: 880px;
  text-align: center;
  font-size: clamp(2.2rem, 6vw, 4.6rem);
  font-weight: 950;
}

.home-link {
  width: min(420px, 100%);
  min-height: clamp(58px, 9vw, 72px);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 12px clamp(16px, 5vw, 28px);
  border: 6px solid #141414;
  border-radius: 20px;
  background: #adff8a;
  color: #1e1e1e;
  text-align: center;
  font-size: clamp(1.05rem, 4vw, 1.8rem);
  font-weight: 950;
  line-height: 1.1;
  overflow-wrap: anywhere;
  box-shadow: 7px 8px 0 #111;
  transform: rotate(-1deg);
  transition: 150ms ease;
}

.home-link:hover,
.home-link:focus-visible {
  background: #fff08f;
}

.home-link:active {
  transform: rotate(1deg);
  box-shadow: 0 0 0 #111;
}

@media (max-width: 520px) {
  .phone-layout {
    top: 36px;
    font-size: 1.04rem;
  }

  h1 {
    max-width: 360px;
  }

  .home-link {
    min-height: 62px;
    border-width: 5px;
    border-radius: 16px;
    box-shadow: 5px 6px 0 #111;
  }
}
</style>
