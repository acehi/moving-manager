// 添加模块加载日志

import axios from 'axios';
import router from '@/router';

// 创建axios实例
const request = axios.create({
  timeout: 5000
});

// 请求拦截器
request.interceptors.request.use(
  config => {
    // 只检查localStorage中的token
    if (checkTokenExists()) {
      config.headers.Authorization = `Bearer ${localStorage.getItem('token')}`;
    } else {
      // token不存在或无效，重定向到登录
      router.push('/login');
    }
    return config;
  },
  error => {
    return Promise.reject(error);
  }
);

// 响应拦截器
request.interceptors.response.use(
  response => {
    // 处理HTTP 200响应中的code值
    
    // 移除基于API错误码的登录检查，统一使用localStorage判断
    return response.data;
    return response.data;
  },
  error => {
    return Promise.reject(error);
  }
);

// 检查token是否存在的辅助函数
 export const checkTokenExists = () => {
  const token = localStorage.getItem('token');
  return !!token && token.trim() !== '' && token.trim() !== 'underfunded';
};

export default request;
export { request };