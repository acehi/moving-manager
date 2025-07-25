import api from './api';

// 获取搬运列表
export const getMoveList = async (params) => {
  const response = await api.post('/move/list', params);
  return response.data;
};

// 创建搬运记录
export const createMove = async (data) => {
  const response = await api.post('/move/create', data);
  return response.data;
};

// 更新搬运记录
export const updateMove = async (data) => {
  const response = await api.post('/move/update', data);
  return response.data;
};

// 下载搬运记录PDF
export const downloadMovePdf = async (moveId) => {
  const response = await api.post('/tag/generate-pdf', {
    move_uid: moveId
  }, {
    responseType: 'blob'
  });
  return response.data;
};