<template>
  <div class="login-container">
    <div class="login-card glass-effect">
      <h2 class="login-title">搬运管家</h2>
      <p class="login-desc">请输入您的手机号登录</p>

      <el-form ref="loginForm" :model="loginForm" :rules="rules" @submit.prevent="handleLogin" class="login-form">
        <el-form-item prop="mobile">
            <input
              v-model="loginForm.mobile"
              placeholder="请输入手机号"
              type="text"
              style="width: 100%; padding: 10px; border: 1px solid #ddd; border-radius: 4px;"
            >
          </el-form-item>
        <el-form-item>
          <el-button type="primary" native-type="submit" class="login-btn">登录 / 注册</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import api from '../services/api';
import { ElForm, ElFormItem, ElInput, ElButton, ElMessage } from 'element-plus';

const loginForm = ref({ mobile: '' });
const rules = ref({
  mobile: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的手机号', trigger: 'blur' }
  ]
});
const router = useRouter();
const route = useRoute();

const handleLogin = async () => {
  try {
      const response = await api.post('/user/auth', { mobile: loginForm.value.mobile });
      // 验证响应状态
      // 使用后端定义的成功状态码（common.CodeSuccess）
      if (response.data.code === 200) {
        // 正确保存token和userName
        localStorage.setItem('token', response.data.token);
        localStorage.setItem('userName', response.data.user.user_name);
        console.log('登录成功，已保存token:', response.data.token);
        console.log('用户信息:', response.user);
        // 登录成功后跳转到之前的页面
        router.push(route.query.redirect ? decodeURIComponent(route.query.redirect) : '/');
      } else {
        throw new Error(response.message || '登录失败: 服务器返回非成功状态');
      }
    } catch (error) {
      console.error('登录失败详情:', error);
      ElMessage.error(`登录失败: ${error.message || '未知错误'}`);
    }
};
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, rgba(123, 31, 162, 0.15), rgba(103, 58, 183, 0.15), rgba(63, 81, 181, 0.15));
  padding: 0 16px;
}

.login-card {
  width: 100%;
  max-width: 400px;
  padding: 40px 20px;
  border-radius: 20px;
  text-align: center;
}

.glass-effect {
  background: rgba(255, 255, 255, 0.75);
  backdrop-filter: blur(15px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 8px 32px rgba(103, 58, 183, 0.1);
}

.login-title {
  color: #673AB7;
  margin-bottom: 10px;
  font-size: 2em;
}

.login-desc {
  color: #888;
  margin-bottom: 30px;
}

.login-btn {
  width: 100%;
}

/* 修复输入框文字不显示和无法输入问题 */
:deep(.el-input__inner), :deep(.el-input__inner:focus), :deep(.el-input__inner.is-focus) {
    color: #333 !important;
    background-color: rgba(255, 255, 255, 0.9) !important;
    caret-color: #333 !important;
}
</style>