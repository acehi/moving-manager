<template>
  <div class="home-container">
    <!-- 移除Back按钮 -->
    <div class="page-title">搬运列表</div>

    <!-- 搜索和创建按钮区域 -->
    <main class="tag-cards-container">
      <div v-if="moveList.length === 0" class="empty-state">
        <p>暂无搬运记录</p>
        <el-button @click="openCreateMoveModal" type="primary" class="create-first-move">创建第一个搬运记录</el-button>
      </div>

      <div v-else class="tag-cards">
        <el-card v-for="move in moveList" :key="move.move_uid" :class="{ 'verified-card': move.is_completed }" class="tag-card" @click="toggleMoveMenu(move.move_uid)">
          <div class="tag-info">
            <!-- 第一行：搬运时间 -->
            <div class="info-row single-column time-row">
              <span class="field-label">搬运时间：</span>
              <span class="field-value">{{ formatDate(move.move_at) }}</span>
              <el-button @click.stop="downloadPdf(move.move_uid)" type="primary" class="pdf-btn">
                <el-icon><Download /></el-icon> 下载PDF
              </el-button>
            </div>

            <!-- 第二行：出发地指向目的地 -->
            <div class="info-row single-column center">
              <span class="location-item">{{ move.start_location }}</span>
              <img class="arrow" src="@/assets/arrow.svg" alt="箭头图标" style="width: 40px; height: 32px; vertical-align: middle;" />
              <span class="location-item">{{ move.end_location }}</span>
            </div>

            <!-- 第三行：未核销标签和已核销标签（两列） -->
            <div class="info-row two-columns">
              <div class="column">
                <span class="field-label">未核销标签：</span>
                <el-tag type="warning">{{ move.unverified_tag_count || 0 }}</el-tag>
              </div>
              <div class="column">
                <span class="field-label">已核销标签：</span>
                <el-tag type="success">{{ move.verified_tag_count || 0 }}</el-tag>
              </div>
            </div>

            <!-- 第四行：标签数和状态（两列） -->
            <div class="info-row two-columns">
              <div class="column">
                <span class="field-label">标签数：</span>
                <el-tag>{{ move.tag_count || 0 }}</el-tag>
              </div>
              <div class="column">
                <span class="field-label">状态：</span>
                <el-tag :type="move.is_completed ? 'success' : 'info'">{{ move.is_completed ? '已完成' : '未完成' }}</el-tag>
              </div>
            </div>

            <!-- 第五行：备注 -->
            <div class="info-row single-column">
              <span class="field-label">备注：</span>
              <span class="field-value">{{ move.remark || '-' }}</span>
            </div>
          </div>

          <div class="tag-menu" v-if="activeMoveUid === move.move_uid">
            <el-button @click="goToTagList(move.move_uid)" type="primary" size="small">查看标签列表</el-button>
            <el-button @click="editMove(move)" type="info" size="small">编辑</el-button>
            <el-button @click="toggleCompleteMove(move.move_uid)" :type="move.is_completed ? 'warning' : 'success'" size="small">{{ move.is_completed ? '标记未完成' : '标记完成' }}</el-button>
            <el-button @click="handleDeleteMove(move.move_uid)" type="danger" size="small">删除</el-button>
          </div>
        </el-card>
      </div>
    </main>

    <div class="action-buttons">
      <el-button class="scan-btn" @click="goToScan" type="primary" icon="Scan">扫描</el-button>
      <el-button class="create-btn" @click="openCreateMoveModal" type="success" icon="Plus">创建搬运</el-button>
    </div>

    <!-- 创建/编辑搬运记录弹窗 -->
    <el-dialog v-model="showMoveModal" :title="isEditing ? '编辑搬运' : '创建搬运'" :center="true" width="70%">
      <el-form :model="currentMove" ref="moveForm" :rules="moveRules" label-width="80px" label-position="left">
        <el-form-item label="搬运时间" prop="move_at">
          <el-date-picker v-model="currentMove.move_at" type="datetime" format="YYYY-MM-DD HH:mm" placeholder="选择搬运时间"></el-date-picker>
        </el-form-item>
        <el-form-item label="出发地" prop="start_location">
          <el-input v-model="currentMove.start_location" maxlength="100"></el-input>
        </el-form-item>
        <el-form-item label="目的地" prop="end_location">
          <el-input v-model="currentMove.end_location" maxlength="100"></el-input>
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="currentMove.remark" type="textarea" maxlength="500"></el-input>
        </el-form-item>
        <el-form-item label="状态" prop="is_completed">
          <el-select v-model="currentMove.is_completed" placeholder="选择状态">
            <el-option label="未完成" :value="0"></el-option>
            <el-option label="已完成" :value="1"></el-option>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showMoveModal = false">取消</el-button>
        <el-button @click="saveMove" type="primary">保存</el-button>
      </template>
    </el-dialog>



    <!-- 删除确认弹窗 -->
    <el-dialog v-model="showDeleteModal" title="确认删除" :center="true" width="60%">
      <p>确定要删除这个搬运记录吗？</p>
      <template #footer>
        <el-button @click="showDeleteModal = false">取消</el-button>
        <el-button @click="confirmDelete" type="danger">确认</el-button>
      </template>
    </el-dialog>

    <!-- 登录弹窗 -->
    <el-dialog v-model="showLoginModal" title="登录提示" :center="true" width="60%">
      <p>将在2秒后自动跳转至登录页面</p>
      <template #footer>
        <el-button @click="showLoginModal = false">取消</el-button>
      </template>
    </el-dialog>

    <!-- 新用户引导 -->
    <div v-if="showGuide" class="guide-overlay">
      <div class="guide-card">
        <div class="guide-step-1">
          <img src="/guide-step-2.svg" alt="创建搬运引导" class="guide-img">
          <p>点击右下角按钮创建您的第一个搬运记录</p>
          <el-button @click="closeGuide" type="primary">我知道了</el-button>
        </div>
      </div>
    </div>

  </div>
</template>

<style scoped>
/* 完全匹配标签列表的样式结构 */
.home-container {
  min-height: 100vh;
  padding-bottom: 80px;
}
.page-title {
  font-size: 26px;
  padding: 20px 16px;
}

/* 容器层级样式 */
.tag-cards-container {
  padding: 0 16px; /* padding在tag-cards-container层 */
}

.tag-cards {
  display: flex;
  flex-direction: column;
  gap: 15px; /* 每个记录上下间隔在tag-cards层 */
}

/* 卡片样式 */
.tag-card {
  transition: all 0.3s ease;
  position: relative;
  background-color: rgba(255, 165, 0, 0.1); /* 未完成：淡桔红色背景 */
}

/* 已完成状态样式 */
:deep(.verified-card) {
  background-color: rgba(144, 238, 144, 0.1); /* 已完成：淡绿色背景 */
}

/* 已完成状态蒙版 */
:deep(.verified-card)::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(200, 200, 200, 0.3); /* 浅灰色蒙版 */
  pointer-events: none; /* 不阻止交互 */
}

/* 每条记录距离边框的padding在el-card__body层 */
:deep(.el-card__body) {
  margin: 0;
}

/* 内容区域样式 */
.tag-info {
  width: 100%;
}

/* 信息行布局 */
.info-row {
  margin-bottom: 10px;
  width: 100%;
}

.single-column {
  display: flex;
  align-items: center;
}

.two-columns {
  display: flex;
  justify-content: space-between;
}

.column {
  flex: 1;
  display: flex;
  align-items: center;
}

/* 字段样式 */
.field-label {
  color: #666;
  margin-right: 8px;
  white-space: nowrap;
}

.field-value {
  flex: 1;
}

/* 出发地目的地样式 */
.location-row {
  justify-content: center;
  margin: 12px 0;
}


/* 标签样式覆盖 */
:deep(.el-tag) {
  margin-left: 5px;
}

/* 底部按钮样式 */
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
  text-align: center;
  padding: 12px 0;
}

/* 页面标题样式 */
.page-header {
  padding: 20px 16px;
  color: #673AB7;
  border-bottom: 1px solid rgba(103, 58, 183, 0.1);
}

/* 空状态样式 */
.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: #666;
}

/* 已完成状态样式 */
.verified-card {
  opacity: 0.5;
  position: relative;
}
.verified-card::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.1);
  pointer-events: none;
}

/* 引导层样式 */
.guide-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.guide-card {
  background: white;
  padding: 20px;
  border-radius: 8px;
  max-width: 300px;
}

.guide-img {
  width: 100%;
  margin-bottom: 15px;
}

.location-item {
  background-color: #ad6598;
  opacity: 0.8;
  color: aliceblue;
  border-radius: 5px;
  padding: 3px 9px;
  margin: 0 2px;
}

.arrow {
  margin: 0 2px;
}
.time-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.pdf-btn {
  margin-left: auto;
}
.tag-menu {
  display: flex;
  justify-content: space-around;
  padding: 10px 0;
  gap: 10px;
}
.tag-menu .el-button {
  flex: 1;
}
</style>


<script setup>
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue';

// 登录倒计时相关变量
const showLoginModal = ref(false);
const intervalId = ref(null);
import { useRouter } from 'vue-router';
import api from '../services/api';
import { ElButton, ElCard, ElPageHeader, ElIcon, ElDialog, ElForm, ElFormItem, ElInput, ElSelect, ElOption, ElDatePicker, ElTag, ElMessage, ElMessageBox } from 'element-plus';
import { Download } from '@element-plus/icons-vue';
import { getMoveList, createMove, updateMove, downloadMovePdf } from '@/services/moveService';
import { formatDate } from '@/utils/dateUtils';
import { checkTokenExists } from '@/utils/request';

const router = useRouter();
const moveList = ref([]);
const activeMoveUid = ref('');
const showMoveModal = ref(false);
const showDeleteModal = ref(false);
const showGuide = ref(false);
const currentDeleteUid = ref('');
const isEditing = ref(false);
const currentMove = ref({
  start_location: '',
  end_location: '',
  move_at: '',
  remark: ''
});
const moveRules = ref({
  start_location: [{ required: true, message: '请输入出发地', trigger: 'blur' }],
  end_location: [{ required: true, message: '请输入目的地', trigger: 'blur' }],
  move_at: [{ required: true, message: '请选择搬运时间', trigger: 'blur' }]
});

// 返回上一页
const goBack = () => {
  router.go(-1);
};

// 获取搬运列表
const fetchMoveList = async () => {
  try {
    const response = await api.post('/move/list', { page: 1, page_size: 8 });
    moveList.value = response.data.moveList || [];
  } catch (error) {
    
  }
};

// 切换菜单显示
const toggleMoveMenu = (moveUid) => {
  activeMoveUid.value = activeMoveUid.value === moveUid ? '' : moveUid;
};

// 打开创建/编辑弹窗
const openCreateMoveModal = () => {
  if (!checkTokenExists()) {
    showLoginModal.value = true;
    startCountdown();
    return;
  }
  showMoveModal.value = true;
  currentMove.value = { start_location: '', end_location: '', move_at: new Date().toISOString().slice(0, 16), remark: '', is_completed: 0 };
  isEditing.value = false;
};

const editMove = (move) => {
  isEditing.value = true;
  currentMove.value = { ...move, move_at: new Date(move.move_at).toISOString().slice(0, 16), is_completed: Number(move.is_completed) };
  showMoveModal.value = true;
};

// 保存搬运记录
const saveMove = async () => {
  try {
    if (isEditing.value) {
      // 将日期字符串转换为时间戳（秒级）
      const moveAtTimestamp = Math.floor(new Date(currentMove.value.move_at).getTime() / 1000);
      await api.post('/move/update', {
        move_uid: currentMove.value.move_uid,
        move_at: moveAtTimestamp,
        start_location: currentMove.value.start_location,
        end_location: currentMove.value.end_location,
        remark: currentMove.value.remark,
        is_completed: parseInt(currentMove.value.is_completed)
      });
    } else {
      // 将日期字符串转换为时间戳（秒级）
      const moveAtTimestamp = Math.floor(new Date(currentMove.value.move_at).getTime() / 1000);
      await api.post('/move/create', {
        ...currentMove.value,
        move_at: moveAtTimestamp
      });
    }
    showMoveModal.value = false;
    fetchMoveList();
  } catch (error) {
    console.error('保存搬运记录失败:', error);
  }
};

// 删除搬运记录
const handleDeleteMove = (moveUid) => {
  currentDeleteUid.value = moveUid;
  showDeleteModal.value = true;
};

// 确认删除
const confirmDelete = async () => {
  try {
    await api.post('/move/delete', { move_uid: currentDeleteUid.value, is_deleted: 1 });
    showDeleteModal.value = false;
    fetchMoveList();
  } catch (error) {
    console.error('删除搬运记录失败:', error);
  }
};

// 切换完成状态
const toggleCompleteMove = async (moveUid) => {
  try {
    const move = moveList.value.find(m => m.move_uid === moveUid);
    // 将日期字符串转换为时间戳（秒级）
    const moveAtTimestamp = Math.floor(new Date(move.move_at).getTime() / 1000);
    await api.post('/move/update', {
      move_uid: moveUid,
      is_completed: move.is_completed ? 0 : 1,
      move_at: moveAtTimestamp,
      start_location: move.start_location,
      end_location: move.end_location,
      remark: move.remark
    });
    fetchMoveList();
  } catch (error) {
    console.error('更新状态失败:', error);
  }
};

// 前往扫描页面
const goToScan = () => {
  if (!checkTokenExists()) {
    showLoginModal.value = true;
    startCountdown();
    return;
  }
  router.push('/scan');
};

// 关闭引导
const closeGuide = () => {
  showGuide.value = false;
};

onMounted(() => {
  fetchMoveList();
  // 检查是否是新用户，这里可以添加新用户判断逻辑
  // showGuide.value = true;
});

// 添加PDF下载方法
const downloadPdf = async (moveId) => {
  try {
    const response = await downloadMovePdf(moveId);
    const url = window.URL.createObjectURL(new Blob([response]));
    const a = document.createElement('a');
    a.href = url;
    // 获取当前搬运记录以生成文件名
    const move = moveList.value.find(m => m.move_uid === moveId);
    a.download = `${move.start_location}-${move.end_location}-${formatDate(move.move_at)}.pdf`;
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
    ElMessage.success('PDF下载成功');
  } catch (error) {
    ElMessage.error('PDF下载失败');
    console.error('下载PDF失败:', error);
  }
};
const goToTagList = (moveUid) => {
  router.push(`/tags/${moveUid}`);
};

// 添加对登录模态框显示状态的监视
watch(showLoginModal, (newVal) => {
    if (newVal) {
      countdown.value = 2;
      startCountdown();
    } else {
    // 当弹窗关闭时清除定时器
    if (intervalId.value) {
      clearTimeout(intervalId.value);
        intervalId.value = null;
    }
  }
});

const startCountdown = () => {
    // 2秒后自动跳转，不显示倒计时
    setTimeout(() => {
      window.location.href = '/login';
      showLoginModal.value = false;
    }, 2000);
  };

</script>

<style scoped>
/* Updated deprecated ::v-deep to :deep() */
:deep(.el-dialog__footer) {
  margin-top: 15px;
}
</style>