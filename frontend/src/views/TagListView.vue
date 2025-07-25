<template>
  <div class="tag-list-container">
    <el-page-header @back="goBack" content="标签列表" class="page-header page-title" />

    <div  style="padding: 0 16px 16px;">
        <el-card class="move-details glass-effect">
        <el-descriptions :column="2" border >
            <el-descriptions-item label="出发地">{{ moveDetails.start_location }}</el-descriptions-item>
            <el-descriptions-item label="目的地">{{ moveDetails.end_location }}</el-descriptions-item>
            <el-descriptions-item label="搬运时间">{{ formatDate(moveDetails.move_at) }}</el-descriptions-item>
            <el-descriptions-item label="状态">
            <el-tag :type="moveDetails.is_completed ? 'success' : 'info'">{{ moveDetails.is_completed ? '已完成' : '未完成' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="未核销">
            <el-tag type="warning">{{ moveDetails.unverified_tag_count }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="已核销">
            <el-tag type="success">{{ moveDetails.verified_tag_count }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="标签数量">{{ moveDetails.tag_count }}</el-descriptions-item>
            <el-descriptions-item label="备注">{{ moveDetails.remark || '无' }}</el-descriptions-item>
        </el-descriptions>
        <el-button @click="downloadPdf" type="primary" class="pdf-btn" style="width: 100%; margin-top: 15px;">下载标签PDF</el-button>
        </el-card>
    </div>

    <main class="tag-cards-container">
      <div v-if="tagList.length === 0" class="empty-state">
        <p>暂无标签记录</p>
        <el-button @click="openCreateTagModal" type="primary" class="create-first-tag">创建第一个标签</el-button>
      </div>

      <div v-else class="tag-cards">
        <el-card v-for="tag in tagList" :key="tag.tag_uid" :class="{ 'verified-card': tag.is_verified }" class="tag-card" @click="toggleTagMenu(tag.tag_uid)">
          <div class="tag-info">
            <div class="tag-header">
              <div class="tag-name">标签名：{{ tag.tag_name }}</div>
              <div class="tag-verification">
                核销状态：<el-tag :type="tag.is_verified ? 'success' : 'warning'"> {{ tag.is_verified ? '已核销' : '未核销' }}</el-tag>
              </div>
            </div>
            <div class="tag-remark">备注：{{ tag.remark || '无备注' }}</div>
          </div>
          <div class="tag-menu" v-if="activeTagUid === tag.tag_uid">
            <el-button @click="editTag(tag)"  type="primary" size="small">编辑</el-button>
            <el-button @click="toggleVerifyTag(tag.tag_uid)" :type="tag.is_verified ? 'warning' : 'success'" size="small">{{ tag.is_verified ? '取消核销' : '核销' }}</el-button>
            <el-button @click="deleteTag(tag.tag_uid)" type="danger" size="small">删除</el-button>
          </div>
        </el-card>
      </div>
    </main>

    <div class="action-buttons">
      <el-button class="scan-btn" @click="goToScan" type="primary" icon="Scan">扫描</el-button>
      <el-button class="create-btn" @click="openCreateTagModal" type="success" icon="Plus">创建标签</el-button>
    </div>

    <!-- 创建/编辑标签弹窗 -->
    <el-dialog v-model="showTagModal" :title="isEditing ? '编辑标签' : '创建标签'" :center="true" width="70%">
      <el-form :model="currentTag" ref="tagForm" :rules="tagRules" label-width="80px" label-position="left">
        <el-form-item label="标签名称" prop="tag_name">
          <el-input v-model="currentTag.tag_name" maxlength="50"></el-input>
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="currentTag.remark" type="textarea" maxlength="500"></el-input>
        </el-form-item>
        <el-form-item label="核销状态" prop="is_verified">
          <el-select v-model="currentTag.is_verified" class="verification-select">
            <el-option label="未核销" :value="0"></el-option>
            <el-option label="已核销" :value="1"></el-option>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showTagModal = false">取消</el-button>
        <el-button @click="saveTag" type="primary">保存</el-button>
      </template>
    </el-dialog>

    <!-- 删除确认弹窗 -->
    <el-dialog v-model="showDeleteModal" title="确认删除" :center="true" width="40%">
      <p>确定要删除这个标签吗？</p>
      <template #footer>
        <el-button @click="showDeleteModal = false">取消</el-button>
        <el-button @click="confirmDelete" type="danger">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, reactive } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import api from '../services/api';
import { ElPageHeader, ElCard, ElDescriptions, ElDescriptionsItem, ElTag, ElButton, ElDialog, ElForm, ElFormItem, ElInput, ElSelect, ElOption, ElMessageBox } from 'element-plus';

const route = useRoute();
const router = useRouter();
const moveUid = route.params.moveUid;

const moveDetails = ref({});
const tagList = ref([]);
const activeTagUid = ref('');
const showTagModal = ref(false);
const showDeleteModal = ref(false);
const isEditing = ref(false);
const currentTag = ref({});
const currentDeleteUid = ref('');
const showVerifyBtn = ref({});
const touchStartX = ref(0);
const touchMoveX = ref(0);
const swipeThreshold = 30; // 降低滑动阈值提高灵敏度
const tagForm = ref(null);

// 表单验证规则
const tagRules = reactive({
  tag_name: [
    { required: true, message: '请输入标签名称', trigger: 'blur' },
    { max: 50, message: '标签名称不能超过50个字符', trigger: 'blur' }
  ],
  is_verified: [
    { required: true, message: '请选择核销状态', trigger: 'change' }
  ]
});

// 获取搬运详情
const fetchMoveDetails = async () => {
  try {
    const response = await api.post('/move/detail', { move_uid: moveUid });
    moveDetails.value = response.data.move;
  } catch (error) {
    console.error('获取搬运详情失败:', error);
  }
};

// 获取标签列表
const fetchTagList = async () => {
  try {
    const response = await api.post('/tag/list', {
      move_uid: moveUid,
      page: 1,
      page_size: 8
    });
    tagList.value = response.data.tags || [];
  } catch (error) {
    console.error('获取标签列表失败:', error);
  }
};

// 创建/编辑标签
const saveTag = async () => {
  try {
    if (isEditing.value) {
      await api.post('/tag/update', {
        tag_uid: currentTag.value.tag_uid,
        tag_name: currentTag.value.tag_name,
        remark: currentTag.value.remark,
        is_verified: currentTag.value.is_verified
      });
    } else {
      await api.post('/tag/create', {
          move_uid: moveUid,
          tag_name: currentTag.value.tag_name,
          remark: currentTag.value.remark, // 修复备注参数
          is_verified: currentTag.value.is_verified
      });
    }
    showTagModal.value = false;
    activeTagUid.value = '';
    fetchTagList();
    resetCurrentTag();
  } catch (error) {
    console.error(isEditing.value ? '更新标签失败:' : '创建标签失败:', error);
  }
};

// 获取状态文本
const getStatusText = (status) => {
  switch(status) {
    case 0: return '正常';
    case 1: return '锁定';
    case 2: return '已完成';
    default: return '未知';
  }
};

// 获取状态样式
const getStatusClass = (status) => {
  switch(status) {
    case 0: return 'status-normal';
    case 1: return 'status-locked';
    case 2: return 'status-completed';
    default: return '';
  }
};

// 删除标签
const deleteTag = (tagUid) => {
  currentDeleteUid.value = tagUid;
  ElMessageBox.confirm(
    '确定要删除这个标签吗？',
    '确认删除',
    {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'warning',
    }
  )
  .then(() => {
    confirmDelete();
  })
  .catch(() => {
    showDeleteModal.value = false;
  });
};

// 确认删除
const confirmDelete = async () => {
  try {
    await api.post('/tag/delete', {
      tag_uid: currentDeleteUid.value,
      is_deleted: 1  // 添加is_deleted字段
    });
    showDeleteModal.value = false;
    fetchTagList();
  } catch (error) {
    console.error('删除标签失败:', error);
  }
};

// 验证/取消验证标签
const toggleVerifyTag = async (tagUid) => {
  try {
    const tag = tagList.value.find(t => t.tag_uid === tagUid);
    await api.post('/tag/verify', { tag_uid: tagUid, is_verified: tag.is_verified ? 0 : 1 });
    fetchTagList();
    // 隐藏核销按钮
    showVerifyBtn.value[tagUid] = false;
  } catch (error) {
    console.error('更新标签验证状态失败:', error);
  }
};

// 点击其他地方隐藏核销按钮
const hideVerifyButtons = (e) => {
  if (!e.target.closest('.verify-btn')) {
    showVerifyBtn.value = {};
  }
};

// 下载PDF
const downloadPdf = async () => {
  try {
    const response = await api.post('/tag/generate-pdf', { move_uid: moveUid }, {
      responseType: 'blob'
    });
    const url = window.URL.createObjectURL(new Blob([response.data]));
    const a = document.createElement('a');
    a.href = url;
    a.download = `${moveDetails.value.start_location}-${moveDetails.value.end_location}-${formatDate(moveDetails.value.move_at)}.pdf`;
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
  } catch (error) {
    console.error('下载PDF失败:', error);
  }
};

// 格式化日期
const formatDate = (dateString) => {
  const date = new Date(dateString);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  });
};

// 打开创建标签弹窗
const openCreateTagModal = () => {
  isEditing.value = false;
  resetCurrentTag();
  showTagModal.value = true;
};

// 打开编辑标签弹窗
const editTag = (tag) => {
  isEditing.value = true;
  currentTag.value = { ...tag };
  showTagModal.value = true;
  activeTagUid.value = '';
};

// 重置当前标签对象
const resetCurrentTag = () => {
  currentTag.value = {
    tag_name: '',
    remark: '',
    is_verified: 0 // 默认未核销
  };
};

// 处理触摸开始
const handleTouchStart = (e, tagUid) => {
  touchStartX.value = e.touches[0].clientX;
  showVerifyBtn.value[tagUid] = false; // 重置显示状态
};

// 处理触摸移动
const handleTouchMove = (e) => {
  touchMoveX.value = e.touches[0].clientX;
};

// 处理触摸结束
const handleTouchEnd = (e, tagUid) => {
  const swipeDistance = touchStartX.value - touchMoveX.value;
  // 向右滑动超过阈值显示核销按钮
  if (swipeDistance > swipeThreshold) {
    showVerifyBtn.value[tagUid] = true;
  } else if (swipeDistance < -swipeThreshold) {
    showVerifyBtn.value[tagUid] = true;
  } else {
    showVerifyBtn.value[tagUid] = false;
  }
};

// 切换标签菜单
const toggleTagMenu = (tagUid) => {
  showVerifyBtn.value = {}; // 关闭所有核销按钮
  if (activeTagUid.value === tagUid) {
    activeTagUid.value = '';
  } else {
    activeTagUid.value = tagUid;
  }
};

// 返回上一页
const goBack = () => {
  router.go(-1);
};

// 前往扫描页面
const goToScan = () => {
  if (!checkTokenExists()) {
    if (window.confirm('用户未登录或未注册，请登录后再试。')) {
      window.location.href = '/login';
      return;
    }
  }
  router.push('/scan');
};

// 初始化
onMounted(() => {
  fetchMoveDetails();
  fetchTagList();
  resetCurrentTag();
  document.addEventListener('click', hideVerifyButtons);
});

onUnmounted(() => {
  document.removeEventListener('click', hideVerifyButtons);
});
</script>

<style scoped>

.page-title {
  font-size: 26px !important;
  padding: 20px 16px;
}

/* 保留必要样式，移除Element Plus已处理的自定义样式 */
.tag-list-container {
  min-height: 100vh;
  padding-bottom: 80px;
}

.tag-cards-container {
  padding: 0 16px;
}

.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: #666;
}

.tag-cards {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.tag-card {
  transition: all 0.3s ease;
  position: relative;
  background-color: rgba(255, 230, 200, 0.5); /* 未核销：淡桔红色背景 */
}

.tag-card.verified-card {
  opacity: 0.7;
  background-color: rgba(200, 255, 200, 0.5); /* 已核销：淡绿色背景 */
}

.tag-card.verified-card::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(200, 200, 200, 0.3); /* 浅灰色蒙版 */
  pointer-events: none;
}

.tag-name {
  color: #673AB7;
  font-weight: 500;
}

.tag-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 5px;
}

.tag-verification {
  margin-left: 10px;
}

.tag-remark {
  color: #666;
  font-size: 0.9em;
  margin-bottom: 5px;
}

.tag-menu {
  display: flex;
  justify-content: space-around;
  width: 100%;
  margin-top: 10px;
}

.tag-menu .el-button {
  flex: 1;
  margin: 0 4px;
}

.action-buttons {
  display: flex;
  gap: 15px;
  padding: 16px;
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(10px);
  z-index: 10;
}

.scan-btn, .create-btn {
  flex: 1;
}
</style>