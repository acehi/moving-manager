import axios from 'axios';
import router from '../router';
import { ElMessage } from 'element-plus';

// 创建axios实例
const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://192.168.2.22:8080/api/v1',
  timeout: 5000,
  headers: {
    'Content-Type': 'application/json'
  }
});

// 请求拦截器 - 添加认证token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    // 统一处理网络错误
    if (!error.response) {
      setTimeout(() => {
        ElMessage.error('网络错误，请检查您的网络连接');
      }, 0);
    }
    return Promise.reject(error);
  }
);

// 响应拦截器 - 统一错误处理
api.interceptors.response.use(
  (response) => {
      // 移除基于接口返回值的登录检查，统一使用localStorage判断
      return response;
      return response;
  },
  (error) => {
    return Promise.reject(error);
  }
);

export default api;