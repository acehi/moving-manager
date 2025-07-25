<template>
  <div class="scan-container">
    <header class="scan-header">
      <el-button @click="goBack" icon="ArrowLeft" class="back-btn">返回</el-button>
      <h1>扫描标签</h1>
    </header>

    <el-card v-if="!scannedTag && !isScanning" class="scanner-initial glass-effect">
      <div class="scanner-icon">
        <i class="qrcode-icon"></i>
      </div>
      <p class="scan-tip">准备扫描标签二维码</p>
      <el-button @click="startScan" type="primary" class="start-scan-btn">开始扫描</el-button>
    </el-card>

    <div v-else-if="isScanning && !scannedTag" class="scanner-view">
      <div class="scanner-stream-container">
        <QrcodeStream
          v-if="isScanning"
          :key="scanKey"
          @detected="onDetected"
          @error="handleScannerError"
          @init="onScannerInit"
          class="scanner-stream"
          :constraints="{ video: { facingMode: { ideal: 'environment' } }}"
          style="width: 100%; height: 100%; object-fit: cover;"
        />
      </div>
      <el-button @click="stopScan" type="default" class="stop-scan-btn">取消扫描</el-button>
    </div>

    <el-card v-else class="tag-result glass-effect">
      <el-tabs v-model="activeTab" type="border-card">
        <el-tab-pane label="搬运信息">
          <el-descriptions :column="1" border>
            <el-descriptions-item label="出发地">{{ scannedTag.move.start_location }}</el-descriptions-item>
            <el-descriptions-item label="目的地">{{ scannedTag.move.end_location }}</el-descriptions-item>
            <el-descriptions-item label="搬运时间">{{ formatDate(scannedTag.move.move_at) }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="scannedTag.move.is_completed ? 'success' : 'info'">{{ scannedTag.move.is_completed ? '已完成' : '未完成' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="标签总数">{{ scannedTag.move.tag_count }}</el-descriptions-item>
            <el-descriptions-item label="已核销">
              <el-tag type="success">{{ scannedTag.move.verified_tag_count }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="未核销">
              <el-tag type="warning">{{ scannedTag.move.unverified_tag_count }}</el-tag>
            </el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>

        <el-tab-pane label="标签信息">
          <el-descriptions :column="1" border>
            <el-descriptions-item label="标签名称">{{ scannedTag.tag.name }}</el-descriptions-item>
            <el-descriptions-item label="物品尺寸">{{ scannedTag.tag.size || '未填写' }}</el-descriptions-item>
            <el-descriptions-item label="物品数量">{{ scannedTag.tag.quantity || '1' }}</el-descriptions-item>
            <el-descriptions-item label="物品位置">{{ scannedTag.tag.location || '未填写' }}</el-descriptions-item>
            <el-descriptions-item label="备注">{{ scannedTag.tag.remark || '无' }}</el-descriptions-item>
            <el-descriptions-item label="核销状态">
              <el-tag :type="scannedTag.tag.status ? 'success' : 'warning'">{{ scannedTag.tag.status ? '已核销' : '未核销' }}</el-tag>
            </el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>
      </el-tabs>

      <div class="action-buttons">
        <el-button @click="resetScan" type="default" class="reset-btn">重新扫描</el-button>
        <el-button @click="toggleVerifyStatus" :type="scannedTag.tag.status ? 'warning' : 'primary'" class="verify-btn">{{ scannedTag.tag.status ? '取消核销' : '确认核销' }}</el-button>
      </div>
    </el-card>

    <!-- 错误提示 -->
    <el-dialog v-model="errorDialogVisible" title="错误提示" :center="true" width="30%">
      <p>{{ errorMessage }}</p>
      <template #footer>
        <el-button @click="clearError">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onUnmounted, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import { QrcodeStream } from 'vue3-qrcode-reader';
import api from '../services/api';
import { ElButton, ElCard, ElDescriptions, ElDescriptionsItem, ElTag, ElTabs, ElTabPane, ElDialog } from 'element-plus';

const router = useRouter();
const isScanning = ref(false);
const scannedTag = ref(null);
const errorMessage = ref('');
const scanKey = ref(0);
const activeTab = ref('0');
const errorDialogVisible = ref(false);

// 开始扫描
  const startScan = async () => {
      try {
          // 分步检查媒体设备API并记录日志
          console.log('[相机兼容性] 检查navigator.mediaDevices:', navigator.mediaDevices);
          console.log('[相机兼容性] 检查window.mediaDevices:', window.mediaDevices);
          console.log('[相机兼容性] 检查navigator.webkitMediaDevices:', navigator.webkitMediaDevices);
          console.log('[相机兼容性] 检查window.webkitMediaDevices:', window.webkitMediaDevices);
          console.log('[相机兼容性] 检查navigator.mozMediaDevices:', navigator.mozMediaDevices);
          console.log('[相机兼容性] 检查window.mozMediaDevices:', window.mozMediaDevices);
          console.log('[相机兼容性] 检查navigator.msMediaDevices:', navigator.msMediaDevices);
          console.log('[相机兼容性] 检查window.msMediaDevices:', window.msMediaDevices);
          
          const mediaDevices = navigator.mediaDevices || 
                               window.mediaDevices || 
                               navigator.webkitMediaDevices || 
                               window.webkitMediaDevices || 
                               navigator.mozMediaDevices || 
                               window.mozMediaDevices || 
                               navigator.msMediaDevices || 
                               window.msMediaDevices;
          
          // 获取getUserMedia函数及其上下文（支持所有可能的前缀和标准版本）
          let getUserMedia = null;
          let getUserMediaContext = null;
          
          // 使用数组迭代方式重构兼容性检查，提高可维护性
          const getUserMediaSources = [
            { obj: mediaDevices, prop: 'getUserMedia', context: mediaDevices },
            { obj: navigator, prop: 'getUserMedia', context: navigator },
            { obj: window, prop: 'getUserMedia', context: window },
            { obj: navigator, prop: 'webkitGetUserMedia', context: navigator },
            { obj: window, prop: 'webkitGetUserMedia', context: window },
            { obj: navigator, prop: 'mozGetUserMedia', context: navigator },
            { obj: window, prop: 'mozGetUserMedia', context: window },
            { obj: navigator, prop: 'msGetUserMedia', context: navigator },
            { obj: window, prop: 'msGetUserMedia', context: window },
            { obj: navigator, prop: 'oGetUserMedia', context: navigator },
            { obj: window, prop: 'oGetUserMedia', context: window }
          ];
          
          // 迭代检测所有可能的getUserMedia来源
          for (const source of getUserMediaSources) {
            console.log(`[相机兼容性] 检查${source.obj === window ? 'window' : 'navigator'}.${source.prop}:`, source.obj && source.obj[source.prop]);
            if (source.obj && typeof source.obj[source.prop] === 'function') {
              getUserMedia = source.obj[source.prop];
              getUserMediaContext = source.context;
              break; // 找到第一个可用实现后停止检测
            }
          }
          
          // 添加兼容性诊断日志
          console.log('[相机兼容性] 检测到的媒体设备对象:', mediaDevices);
          console.log('[相机兼容性] 检测到的getUserMedia函数:', getUserMedia);
          console.log('[相机兼容性] 浏览器用户代理:', navigator.userAgent);
          
          if (!getUserMedia) {
            // 在错误消息中包含浏览器标识信息，便于诊断
            const browserInfo = navigator.userAgent.substring(0, 100); // 限制长度
            showError(`您的浏览器不支持相机访问功能，请使用最新版浏览器。浏览器信息: ${browserInfo}`);
            // 记录详细错误信息到控制台
            console.error('[相机兼容性错误] 未找到支持的getUserMedia实现', {
              mediaDevices: !!mediaDevices,
              navigatorMediaDevices: !!navigator.mediaDevices,
              userAgent: navigator.userAgent
            });
            return;
          }
          
          // 预检查并请求相机权限（使用显式上下文确保兼容性）
          await new Promise((resolve, reject) => {
            if (!getUserMedia || !getUserMediaContext) {
              reject(new Error('未找到有效的getUserMedia实现'));
              return;
            }
            
            console.log('[相机兼容性] 使用上下文调用getUserMedia:', getUserMediaContext);
            
            // 现代浏览器：Promise-based API
            if (getUserMedia.length <= 1) {
              getUserMedia.call(getUserMediaContext, { video: true })
                .then(resolve)
                .catch(reject);
            } else {
              // 旧版浏览器：Callback-based API
              getUserMedia.call(getUserMediaContext, { video: true }, resolve, reject);
            }
          });
        
        scanKey.value++;
        isScanning.value = true;
        await nextTick();
        
        // 安卓设备添加延迟确保组件渲染完成
        if (/Android/i.test(navigator.userAgent)) {
          setTimeout(() => {
            console.log('[安卓优化] 扫描初始化延迟完成，等待相机响应');
          }, 2000);
        }
      } catch (err) {
        if (err.name === 'NotAllowedError') {
          showError('需要相机权限才能扫描，请在浏览器设置中启用相机权限');
        } else if (err.name === 'NotFoundError') {
          showError('未检测到可用相机设备');
        } else {
          showError('扫描组件初始化失败: 未获取到相机实例');
        }
        return;
      }
    };

// 扫描仪初始化处理
const onScannerInit = (payload) => {
  console.log('[安卓设备] 扫描器初始化 payload:', payload);
  console.log('[安卓设备] 浏览器信息:', navigator.userAgent);
  console.log('[安卓设备] 相机约束:', { video: { facingMode: { ideal: 'environment' } } });
  
  // 检查payload是否包含必要方法
  if (!payload) {
    showError('扫描组件初始化失败: 未获取到相机实例');
    stopScan();
    return;
  }
  
  // 移除显式start()调用，由组件内部处理相机启动
  console.log('[安卓设备] 扫描组件已初始化，等待内部相机启动');
};

// 停止扫描
const stopScan = () => {
  isScanning.value = false;
};

// 处理扫描结果
const onDetected = async (result) => {
  if (result && result.content) {
    try {
      stopScan();
      const tagUid = result.content;
      // 获取标签详情
      const response = await api.post('/tags/detail', { tag_uid: tagUid });
      const tagDetail = response.data.tag;

      // 获取搬运详情
      const moveResponse = await api.post('/moves/detail', { move_uid: tagDetail.move_uid });
      const moveDetail = moveResponse.data.move;

      scannedTag.value = {
        tag: tagDetail,
        move: moveDetail
      };
    } catch (error) {
      console.error('扫描失败:', error);
      showError('扫描失败，无法识别该标签');
    }
  }
};

// 处理扫描错误
const handleScannerError = (error) => {
  console.error('扫描器错误:', { name: error.name, message: error.message, stack: error.stack });
  let errorMsg = `扫描出错: ${error.name || '未知错误'}`;
    console.error('[扫描错误] 完整错误信息:', error);
  
  // 更详细的错误类型判断
  if (error.name === 'NotAllowedError') {
    errorMsg = '相机权限被拒绝，请在浏览器设置中启用相机权限';
  } else if (error.name === 'NotFoundError') {
    errorMsg = '未检测到可用相机，请确保您的设备有摄像头并尝试重新加载';
  } else if (error.name === 'NotSupportedError') {
    errorMsg = '当前环境不支持相机访问，请使用HTTPS或localhost';
  } else if (error.name === 'OverconstrainedError') {
    errorMsg = '无法满足相机要求，请尝试更换设备';
  } else if (error.name === 'NotReadableError') {
    errorMsg = '相机被占用或无法访问';
  } else if (error.name === 'SecurityError') {
    if (window.location.protocol !== 'https:') {
      errorMsg = '请在HTTPS环境下使用相机功能';
    } else {
      errorMsg = '安全错误：无法访问相机，请检查您的浏览器设置';
    }
  }
  
  // 针对安卓设备的特定提示
  if (/Android/i.test(navigator.userAgent)) {
    errorMsg += '\n\n安卓用户提示：请确保应用拥有相机权限并尝试使用Chrome浏览器';
  }
  
  displayError(errorMsg);
  isScanning.value = false;
};

// 显示错误信息
const displayError = (message) => {
  errorMessage.value = message;
  errorDialogVisible.value = true;
};
const clearError = () => {
  errorDialogVisible.value = false;
  errorMessage.value = '';
};

// 切换核销状态
const toggleVerifyStatus = async () => {
  try {
    await api.post('/tags/verify', { tag_uid: scannedTag.value.tag.tag_uid });
    scannedTag.value.tag.status = scannedTag.value.tag.status ? 0 : 1;
    // 更新搬运的核销数量
    if (scannedTag.value.tag.status) {
      scannedTag.value.move.verified_tag_count++;
      scannedTag.value.move.unverified_tag_count--;
    } else {
      scannedTag.value.move.verified_tag_count--;
      scannedTag.value.move.unverified_tag_count++;
    }
  } catch (error) {
    console.error('更新核销状态失败:', error);
    showError('操作失败，请重试');
  }
};

// 重置扫描状态
const resetScan = () => {
  scannedTag.value = null;
};

// 返回上一页
const goBack = () => {
  router.go(-1);
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

// 组件卸载时停止扫描
onUnmounted(() => {
  stopScan();
});
</script>

<style scoped>
/* 保留必要样式，移除Element Plus已处理的自定义样式 */
.scan-container {
  min-height: 100vh;
  padding: 20px;
}
.scanner-stream-container {
  position: relative;
  width: 100%;
  height: 300px;
  margin-bottom: 20px;
}
.action-buttons {
  display: flex;
  gap: 10px;
  margin-top: 20px;
}
</style>