import { createRouter, createWebHistory } from 'vue-router'
import { useInitStore } from '@/stores/init'
import { useUserStore } from '@/stores/user'

// 所有路由组件使用懒加载，优化首屏加载性能
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('../views/home/HomeView.vue'),
      meta: {
        title: 'Home',
        description: 'Ech0 home timeline for publishing and browsing thoughts, notes, and links.',
        optionalAuth: true,
      },
    },
    {
      path: '/publish',
      name: 'publish',
      redirect: { name: 'home', query: { tab: 'publish' } },
    },
    {
      path: '/panel',
      name: 'panel',
      component: () => import('../views/panel/PanelView.vue'),
      redirect: '/panel/dashboard',
      meta: {
        title: 'Panel',
        description: 'Ech0 management panel.',
        requiresAuth: true,
        noindex: true,
      },
      children: [
        {
          path: 'dashboard',
          name: 'panel-dashboard',
          component: () => import('../views/panel/modules/TheDashboard.vue'),
        },
        {
          path: 'setting',
          name: 'panel-setting',
          component: () => import('../views/panel/modules/TheSetting.vue'),
        },
        {
          path: 'user',
          name: 'panel-user',
          component: () => import('../views/panel/modules/TheUser.vue'),
        },
        {
          path: 'storage',
          name: 'panel-storage',
          component: () => import('../views/panel/modules/TheStorage.vue'),
        },
        {
          path: 'data-management',
          name: 'panel-data-management',
          component: () => import('../views/panel/modules/TheDataManagement.vue'),
        },
        {
          path: 'sso',
          name: 'panel-sso',
          component: () => import('../views/panel/modules/TheSSO.vue'),
        },
        {
          path: 'extension',
          name: 'panel-extension',
          component: () => import('../views/panel/modules/TheExtension.vue'),
        },
        {
          path: 'comment',
          name: 'panel-comment',
          component: () => import('../views/panel/modules/TheCommentManager.vue'),
        },
        {
          path: 'advance',
          name: 'panel-advance',
          component: () => import('../views/panel/modules/TheAdvance.vue'),
        },
        {
          path: 'system-log',
          name: 'panel-system-log',
          component: () => import('../views/panel/modules/TheSystemLog.vue'),
        },
      ],
      // beforeEnter: (to, from, next) => {
      //   const userStore = useUserStore()
      //   if (userStore.isLogin) {
      //     next()
      //   } else {
      //     next({ name: 'auth' })
      //   }
      // },
    },
    {
      path: '/auth',
      name: 'auth',
      component: () => import('../views/auth/AuthView.vue'),
      meta: {
        title: 'Sign In',
        description: 'Sign in to your Ech0 workspace.',
        noindex: true,
      },
    },
    {
      path: '/widget',
      name: 'widget',
      component: () => import('../views/widget/WidgetView.vue'),
      meta: {
        title: 'Widget',
        description: 'Ech0 embeddable widget.',
        noindex: true,
      },
    },
    {
      path: '/init',
      name: 'init',
      component: () => import('../views/init/InitView.vue'),
      meta: {
        title: 'Initialize',
        description: 'Initialize your Ech0 instance.',
        noindex: true,
      },
    },
    {
      path: '/hub',
      name: 'hub',
      component: () => import('../views/hub/HubView.vue'),
      meta: {
        title: 'Hub',
        description: 'Discover and explore curated content in Ech0 hub.',
      },
    },
    {
      path: '/echo/:echoId',
      name: 'echo',
      component: () => import('../views/echo/EchoView.vue'),
      meta: {
        title: 'Echo',
        description: 'Read a shared Ech0 post.',
      },
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('../views/404/NotFoundView.vue'),
      meta: {
        title: '404',
        description: 'Requested page was not found.',
        noindex: true,
      },
    },
  ],
})

// 全局路由守卫
router.beforeEach(async (to) => {
  const initStore = useInitStore()
  const userStore = useUserStore()

  if (!initStore.ready) {
    await initStore.init()
  }

  const isInitReady = initStore.initialized || initStore.ownerExists

  if (!isInitReady && to.name !== 'init') {
    return { name: 'init' }
  }

  if (isInitReady && to.name === 'init') {
    return { name: 'auth' }
  }

  // 等待用户信息初始化完成
  if (!userStore.initialized) {
    await userStore.init()
  }

  const needRedirect = localStorage.getItem('needLoginRedirect')

  if (
    (to.meta.requiresAuth && !userStore.isLogin) ||
    (to.meta.optionalAuth && !userStore.isLogin && needRedirect === 'true')
  ) {
    localStorage.removeItem('needLoginRedirect')
    return { name: 'auth' }
  }

  return true
})

export default router
