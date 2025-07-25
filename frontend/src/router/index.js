import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import TagListView from '../views/TagListView.vue'
import ScanView from '../views/ScanView.vue'
import LoginView from '../views/LoginView.vue'
import TagDetailView from '../views/TagDetailView.vue'
import { checkTokenExists } from '../utils/request'
import { ElMessage } from 'element-plus'

const routes = [
  { path: '/', name: 'home', component: HomeView },
  { path: '/tags/:moveUid', name: 'tagList', component: TagListView, props: true },
  { path: '/scan', name: 'scan', component: ScanView },
  { path: '/login', name: 'login', component: LoginView },
  { path: '/tag/:tagUid', name: 'tagDetail', component: TagDetailView, props: true }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 添加全局前置守卫
router.beforeEach((to, from, next) => {
  // 不需要登录的页面
  const publicPages = ['/login'];
  const authRequired = !publicPages.includes(to.path);
  
  if (authRequired && !checkTokenExists()) {
      // 未登录，显示提示并延迟2秒跳转
      ElMessage.warning({
        message: '未登录，请先登录',
        duration: 2000
      });
      setTimeout(() => {
          // 保存目标路径作为 redirect 参数
          next(`/login?redirect=${encodeURIComponent(to.path)}`);
        }, 2000);
      return;
    }
  
  next();
})

export default router