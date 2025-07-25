import request from '@/utils/request';
import api from './api';
import router from '../router';

/**
 * 获取标签详情
 * @param {string} tagId - 标签ID
 * @returns {Promise<Object>} 标签详情对象
 */
export const getTagDetail = async (tagUid) => {
  try {
    const response = await api({
    url: '/tag/detail',
    method: 'post',
    data: { tag_uid: tagUid }
  });
    return response;
  } catch (error) {
    console.error('获取标签详情API调用失败:', error);
    throw error; // 继续抛出错误，让调用方处理
  }
};

/**
 * 核销标签
 * @param {string} tagId - 标签ID
 * @returns {Promise<Object>} 核销结果
 */
export const verifyTag = async (tagUid, isVerified) => {
  try {
      const response = await api.post('/tag/verify', {
        tag_uid: tagUid,
        is_verified: isVerified ? 1 : 0
    });
    return response.data;
  } catch (error) {
    console.error('核销标签失败:', error);
    throw error;
  }
};

/**
 * 取消核销标签
 * @param {string} tagId - 标签ID
 * @returns {Promise<Object>} 取消核销结果
 */
export const unverifyTag = async (tagUid) => {
  try {
    const response = await api.post('/v1/tag/update', {
      tag_uid: tagUid,
      is_verified: 0
    });
    return response.data;
  } catch (error) {
    console.error('取消核销标签失败:', error);
    throw error;
  }
};

/**
 * 获取标签列表（按搬运单）
 * @param {string} moveUid - 搬运单UID
 * @returns {Promise<Array>} 标签列表
 */
export const getTagsByMove = async (moveUid) => {
  try {
    const response = await api.get(`/moves/${moveUid}/tags`);
    return response.data;
  } catch (error) {
    console.error('获取标签列表失败:', error);
    throw error;
  }
};

// 导出默认实例供其他可能的使用
export default api;