<template>
  <div class="tag-detail-container">
    <el-card v-loading="loading" :bordered="false">
      <template #header>
        <h2>标签详情</h2>
      </template>
      
      <div v-if="tagDetail">
        <div class="two-column-layout">
            <!-- 标签信息块 -->
            <div class="info-section">
                <h3 class="block-title grid-span-2">标签信息</h3>
                <div class="key-value-pair">
                    <span class="key">标签名：</span>
                    <span class="value">{{ tagDetail.tag_name || '无' }}</span>
                </div>
                <div class="key-value-pair">
                    <span class="key">状态：</span>
                    <span class="value">
                    <el-tag :type="tagDetail.is_verified ? 'success' : 'warning'">{{ tagDetail.is_verified ? '已核销' : '未核销' }}</el-tag>
                    </span>
                </div>
                <div class="key-value-pair full-width">
                    <span class="key">备注：</span>
                    <span class="value">{{ tagDetail.remark || '无' }}</span>
                </div>
            </div>

            <!-- 搬运信息块 -->
            <div class="info-section">
                <h3 class="block-title grid-span-2">搬运信息</h3>
                <div class="key-value-pair full-width">
                    <span class="key">搬运时间：</span>
                    <span class="field-value">{{ formatDate(tagDetail.move_at) }}</span>
                </div>
                <div class="key-value-pair">
                    <span class="location-item">{{ tagDetail.start_location }}</span>
                    <img class="arrow" src="@/assets/arrow.svg" alt="箭头图标" style="width: 40px; height: 32px;" />
                    <span class="location-item">{{ tagDetail.end_location }}</span>
                </div>
                <div class="key-value-pair">
                    <span class="key">标签数：</span>
                    <el-tag>{{ tagDetail.tag_count || 0 }}</el-tag>
                </div>
                <div class="key-value-pair">
                    <span class="key">未核销：</span>
                    <el-tag type="warning">{{ tagDetail.unverified_tag_count || 0 }}</el-tag>
                </div>
                <div class="key-value-pair">
                    <span class="key">状态：</span>
                    <span class="value">
                    <el-tag :type="tagDetail.is_completed ? 'success' : 'warning'">{{ tagDetail.is_completed ? '已完成' : '未完成' }}</el-tag>
                    </span>
                </div>
                <div class="key-value-pair full-width">
                    <span class="key">备注：</span>
                    <span class="value">{{ tagDetail.remark || '无' }}</span>
                </div>
            </div>
          </div>

        <!-- 操作按钮 -->
        <div class="action-buttons">
          <el-button type="primary" size="large" style="width: 100%;font-size: 16px;" @click="handleVerification" :loading="buttonLoading">{{ tagDetail.is_verified ? '取消核销' : '核销' }}</el-button>
        </div>
      </div>

      <div v-else-if="!loading" class="empty-state">
        <el-empty description="未找到标签信息"></el-empty>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { getTagDetail, verifyTag, unverifyTag } from '@/services/tagService';
import { formatDate } from '@/utils/date';
import { ElMessage } from 'element-plus';

const route = useRoute();
const tagDetail = ref(null);
const loading = ref(true);
const error = ref(null);
const buttonLoading = ref(false);

const handleVerification = async () => {
  if (!tagDetail.value) return;
  
  buttonLoading.value = true;
  try {
      const newStatus = !tagDetail.value.is_verified;
      await verifyTag(tagDetail.value.tag_uid, newStatus);
      // 重新获取标签详情以刷新所有数据
      const updatedResponse = await getTagDetail(tagDetail.value.tag_uid);
      if (updatedResponse && updatedResponse.data && updatedResponse.data.tag) {
        tagDetail.value = updatedResponse.data.tag;
      } else {
        console.error('Failed to refresh tag detail after verification');
      }
      ElMessage.success(newStatus ? '核销成功' : '取消核销成功');
    } catch (err) {
      console.error('Verification update failed:', err);
      ElMessage.error('操作失败，请重试');
    } finally {
      buttonLoading.value = false;
    }
};

onMounted(async () => {
  try {
    const tagUid = route.params.tagUid;
    if (!tagUid) {
      console.error('Tag UID is missing from route parameters');
      ElMessage.error('标签ID不存在');
      return;
    }
    
    const response = await getTagDetail(tagUid);
      if (!response) {
        console.error('getTagDetail returned undefined');
        ElMessage.error('获取标签详情失败: 无效的响应');
        return;
      }
      if (!response.data) {
        console.error('Response data is undefined');
        ElMessage.error('获取标签详情失败: 响应数据为空');
        return;
      }
      console.log('Tag detail response:', response);
      // 直接访问返回数据的属性，无需通过data嵌套
      if (response.data?.code === 200) {
        tagDetail.value = response.data.tag;
      } else {
        console.error('API error:', response.data?.message || '未知错误');
        ElMessage.error(response.data?.message || '获取标签详情失败');
      }
      loading.value = false;
  } catch (error) {
        console.error('Error fetching tag detail:', error);
        console.error('Error details:', { message: error.message, stack: error.stack, response: error.response });
        ElMessage.error(`网络错误：${error.message || '无法获取标签详情'}`);
    } finally {
        loading.value = false;
    }
});
</script>

<style scoped>
.tag-detail-container {
  padding: 20px;
}

.two-column-layout {
  display: flex;
  flex-direction: column;
  gap: 20px;
  margin-bottom: 20px;
  max-width: 1000px;
  margin-left: auto;
  margin-right: auto;
}

.info-section {
  padding: 15px;
  border: 1px solid #e8e8e8;
  border-radius: 4px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
  margin-bottom: 10px;
  width: 100%;
}

.block-title {
  margin-bottom: 6px;
  color: #1890ff;
  font-size: 16px;
  font-weight: bold;
}

.grid-span-2 {
  grid-column: span 2;
}

.key-value-pair {
  display: flex;
  margin-bottom: 12px;
  line-height: 24px;
  align-items: center;
}

.key {  flex: 0 0 80px;  font-size: 16px;  color: #666;}

.full-width {
  grid-column: span 2;
}

.value {
  flex: 1;
  color: #303133;
  word-break: break-word;
}
.location-item {
  background-color: #ad6598;
  opacity: 0.8;
  color: aliceblue;
  border-radius: 5px;
  padding: 3px 9px;
}

.action-buttons {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.empty-state {
  margin: 40px 0;
  text-align: center;
}

.empty-move-info {
  text-align: center;
  padding: 20px;
  color: #909399;
}
.origin-destination {
  margin: 10px 0;
  padding: 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.origin-destination .el-icon-arrow-right {
  color: #1890ff;
  margin: 0 10px;
  font-size: 16px;
}

</style>