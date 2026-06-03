import { createRouter, createWebHistory } from "vue-router";

import HomePage from "@/pages/HomePage.vue";
import LobbyPage from "@/pages/LobbyPage.vue";
import TopicPage from "@/pages/TopicPage.vue";
import RoundPage from "@/pages/RoundPage.vue";
import ResultsPage from "@/pages/ResultsPage.vue";
import NotFoundPage from "@/pages/NotFoundPage.vue";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "Home",
      component: HomePage,
    },
    {
      path: "/room/:lobbyId",
      name: "Lobby",
      meta: {
        title: "Игровая комната",
      },
      component: LobbyPage,
      props: true,
    },
    {
      path: "/room/:lobbyId/topic",
      name: "Topic",
      meta: {
        title: "Тема раунда",
      },
      component: TopicPage,
      props: true,
    },
    {
      path: "/room/:lobbyId/round",
      name: "Round",
      meta: {
        title: "Раунд верстки",
      },
      component: RoundPage,
      props: true,
    },
    {
      path: "/room/:lobbyId/results",
      name: "Results",
      meta: {
        title: "Результаты",
      },
      component: ResultsPage,
      props: true,
    },
    {
      path: "/:pathMatch(.*)*",
      name: "NotFound",
      meta: {
        title: "Страница не найдена",
      },
      component: NotFoundPage,
    },
  ],
});

router.afterEach((to) => {
  if (to.meta.title) {
    document.title = to.meta.title as string;
  }
});

export default router;
