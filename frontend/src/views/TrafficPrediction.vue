<template>
  <div class="traffic-prediction">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stat-card">
          <template #header>
            <div class="card-header">
              <span>当前车流量</span>
              <el-icon :size="24" color="#409EFF"><TrendCharts /></el-icon>
            </div>
          </template>
          <div class="stat-value">{{ currentTraffic }}</div>
          <div class="stat-change" :class="trafficChange > 0 ? 'up' : 'down'">
            <el-icon v-if="trafficChange > 0"><CaretTop /></el-icon>
            <el-icon v-else><CaretBottom /></el-icon>
            {{ Math.abs(trafficChange).toFixed(1) }}% 较昨日
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <template #header>
            <div class="card-header">
              <span>预测峰值</span>
              <el-icon :size="24" color="#e6a23c"><Warning /></el-icon>
            </div>
          </template>
          <div class="stat-value">{{ peakPrediction }}</div>
          <div class="stat-time">预计时间：{{ peakTime }}</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <template #header>
            <div class="card-header">
              <span>预计空闲率</span>
              <el-icon :size="24" color="#67c23a"><CircleCheck /></el-icon>
            </div>
          </template>
          <div class="stat-value">{{ vacancyRate }}%</div>
          <div class="stat-status">
            <el-tag :type="vacancyRate > 30 ? 'success' : vacancyRate > 10 ? 'warning' : 'danger'">
              {{ vacancyRate > 30 ? '充足' : vacancyRate > 10 ? '紧张' : '紧张' }}
            </el-tag>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <template #header>
            <div class="card-header">
              <span>预测准确率</span>
              <el-icon :size="24" color="#909399"><DataAnalysis /></el-icon>
            </div>
          </template>
          <div class="stat-value">{{ accuracyRate }}%</div>
          <el-progress :percentage="accuracyRate" :color="getAccuracyColor(accuracyRate)" />
        </el-card>
      </el-col>
    </el-row>
    
    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>24小时流量预测</span>
              <el-button-group>
                <el-button :type="chartPeriod === 'today' ? 'primary' : 'default'" @click="chartPeriod = 'today'">今日</el-button>
                <el-button :type="chartPeriod === 'tomorrow' ? 'primary' : 'default'" @click="chartPeriod = 'tomorrow'">明日</el-button>
                <el-button :type="chartPeriod === 'week' ? 'primary' : 'default'" @click="chartPeriod = 'week'">本周</el-button>
              </el-button-group>
            </div>
          </template>
          <div ref="trafficChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>车位占用预测</span>
              <el-button type="primary" @click="refreshPrediction">
                <el-icon><Refresh /></el-icon>刷新预测
              </el-button>
            </div>
          </template>
          <div ref="occupancyChartRef" class="chart-container"></div>
        </el-card>
      </el-col>
    </el-row>
    
    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="8">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>车牌识别记录</span>
            </div>
          </template>
          <el-table :data="licensePlateRecords" style="width: 100%" size="small">
            <el-table-column prop="licensePlate" label="车牌号" width="120" />
            <el-table-column prop="type" label="类型" width="80">
              <template #default="scope">
                <el-tag :type="scope.row.type === 'entry' ? 'success' : 'danger'" size="small">
                  {{ scope.row.type === 'entry' ? '入场' : '出场' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="time" label="时间" width="120">
              <template #default="scope">
                {{ formatTime(scope.row.time) }}
              </template>
            </el-table-column>
            <el-table-column prop="confidence" label="置信度" width="80">
              <template #default="scope">
                <el-progress :percentage="scope.row.confidence * 100" :stroke-width="10" />
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>异常停车检测</span>
              <el-button type="danger" size="small" @click="checkAnomalies">
                <el-icon><Search /></el-icon>检测
              </el-button>
            </div>
          </template>
          <el-table :data="anomalyRecords" style="width: 100%" size="small">
            <el-table-column prop="spotNumber" label="车位号" width="80" />
            <el-table-column prop="licensePlate" label="车牌号" width="100" />
            <el-table-column prop="type" label="异常类型" width="100">
              <template #default="scope">
                <el-tag :type="getAnomalyTagType(scope.row.type)" size="small">
                  {{ getAnomalyTypeText(scope.row.type) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="detectedTime" label="检测时间" width="120">
              <template #default="scope">
                {{ formatTime(scope.row.detectedTime) }}
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="anomalyRecords.length === 0" description="暂无异常停车记录" />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>预测模型状态</span>
            </div>
          </template>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="模型版本">
              {{ modelVersion }}
            </el-descriptions-item>
            <el-descriptions-item label="最后训练时间">
              {{ formatTime(lastTrainingTime) }}
            </el-descriptions-item>
            <el-descriptions-item label="训练数据量">
              {{ trainingDataCount }} 条
            </el-descriptions-item>
            <el-descriptions-item label="模型精度">
              <el-progress :percentage="modelPrecision" :color="getAccuracyColor(modelPrecision)" :stroke-width="15" />
            </el-descriptions-item>
            <el-descriptions-item label="模型状态">
              <el-tag type="success">运行中</el-tag>
            </el-descriptions-item>
          </el-descriptions>
          <div style="margin-top: 15px;">
            <el-button type="primary" style="width: 100%;" @click="retrainModel">
              <el-icon><RefreshRight /></el-icon>重新训练
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick, watch } from 'vue'
import * as echarts from 'echarts'
import { ElMessage } from 'element-plus'
import { predictionApi } from '../api'

const trafficChartRef = ref(null)
const occupancyChartRef = ref(null)
const chartPeriod = ref('today')

const currentTraffic = ref(128)
const trafficChange = ref(12.5)
const peakPrediction = ref(256)
const peakTime = ref('18:00')
const vacancyRate = ref(35)
const accuracyRate = ref(92.3)

const modelVersion = ref('v2.3.1')
const lastTrainingTime = ref(Date.now() - 24 * 60 * 60 * 1000)
const trainingDataCount = ref(125680)
const modelPrecision = ref(94.5)

const licensePlateRecords = ref([
  { licensePlate: '京A12345', type: 'entry', time: Date.now() - 5 * 60 * 1000, confidence: 0.98 },
  { licensePlate: '京B67890', type: 'exit', time: Date.now() - 10 * 60 * 1000, confidence: 0.96 },
  { licensePlate: '京C11111', type: 'entry', time: Date.now() - 15 * 60 * 1000, confidence: 0.99 },
  { licensePlate: '京D22222', type: 'entry', time: Date.now() - 20 * 60 * 1000, confidence: 0.95 },
  { licensePlate: '京E33333', type: 'exit', time: Date.now() - 25 * 60 * 1000, confidence: 0.97 }
])

const anomalyRecords = ref([
  { spotNumber: 'A05', licensePlate: '京F44444', type: 'overtime', detectedTime: Date.now() - 30 * 60 * 1000 },
  { spotNumber: 'B03', licensePlate: '京G55555', type: 'wrong_spot', detectedTime: Date.now() - 2 * 60 * 60 * 1000 }
])

let trafficChart = null
let occupancyChart = null

const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const getAccuracyColor = (value) => {
  if (value >= 90) return '#67c23a'
  if (value >= 80) return '#e6a23c'
  return '#f56c6c'
}

const getAnomalyTypeText = (type) => {
  const typeMap = {
    overtime: '超时停车',
    wrong_spot: '车位错误',
    no_plate: '无车牌',
    suspicious: '可疑车辆'
  }
  return typeMap[type] || type
}

const getAnomalyTagType = (type) => {
  const typeMap = {
    overtime: 'warning',
    wrong_spot: 'danger',
    no_plate: 'info',
    suspicious: 'warning'
  }
  return typeMap[type] || 'warning'
}

const generateTrafficChartData = () => {
  const hours = Array.from({ length: 24 }, (_, i) => `${String(i).padStart(2, '0')}:00`)
  const actualData = []
  const predictedData = []
  
  for (let i = 0; i < 24; i++) {
    if (i < 12) {
      actualData.push(Math.floor(50 + Math.sin(i / 4) * 30 + Math.random() * 10))
      predictedData.push(actualData[i] + Math.floor(Math.random() * 5 - 2))
    } else {
      actualData.push(null)
      predictedData.push(Math.floor(80 + Math.sin((i - 6) / 4) * 50 + Math.random() * 10))
    }
  }
  
  return { hours, actualData, predictedData }
}

const generateOccupancyChartData = () => {
  const zones = ['A区', 'B区', 'C区', 'D区', 'VIP区']
  const currentOccupancy = []
  const predictedOccupancy = []
  
  for (let i = 0; i < 5; i++) {
    currentOccupancy.push(Math.floor(50 + Math.random() * 30))
    predictedOccupancy.push(Math.floor(60 + Math.random() * 25))
  }
  
  return { zones, currentOccupancy, predictedOccupancy }
}

const initTrafficChart = async () => {
  await nextTick()
  if (!trafficChartRef.value) return
  
  trafficChart = echarts.init(trafficChartRef.value)
  
  const data = generateTrafficChartData()
  
  const option = {
    tooltip: {
      trigger: 'axis'
    },
    legend: {
      data: ['实际流量', '预测流量'],
      bottom: 0
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: data.hours
    },
    yAxis: {
      type: 'value',
      name: '车辆数'
    },
    series: [
      {
        name: '实际流量',
        type: 'line',
        smooth: true,
        data: data.actualData,
        itemStyle: {
          color: '#409EFF'
        },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(64, 158, 255, 0.3)' },
            { offset: 1, color: 'rgba(64, 158, 255, 0.05)' }
          ])
        }
      },
      {
        name: '预测流量',
        type: 'line',
        smooth: true,
        data: data.predictedData,
        itemStyle: {
          color: '#67c23a'
        },
        lineStyle: {
          type: 'dashed'
        }
      }
    ]
  }
  
  trafficChart.setOption(option)
  
  window.addEventListener('resize', () => {
    trafficChart?.resize()
  })
}

const initOccupancyChart = async () => {
  await nextTick()
  if (!occupancyChartRef.value) return
  
  occupancyChart = echarts.init(occupancyChartRef.value)
  
  const data = generateOccupancyChartData()
  
  const option = {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      }
    },
    legend: {
      data: ['当前占用率', '预测占用率'],
      bottom: 0
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: data.zones
    },
    yAxis: {
      type: 'value',
      name: '占用率(%)',
      max: 100
    },
    series: [
      {
        name: '当前占用率',
        type: 'bar',
        data: data.currentOccupancy,
        itemStyle: {
          color: '#409EFF'
        }
      },
      {
        name: '预测占用率',
        type: 'bar',
        data: data.predictedOccupancy,
        itemStyle: {
          color: '#e6a23c'
        }
      }
    ]
  }
  
  occupancyChart.setOption(option)
  
  window.addEventListener('resize', () => {
    occupancyChart?.resize()
  })
}

const refreshPrediction = () => {
  ElMessage.success('预测数据已刷新')
  initTrafficChart()
  initOccupancyChart()
}

const checkAnomalies = async () => {
  try {
    ElMessage.info('正在检测异常停车...')
    await new Promise(resolve => setTimeout(resolve, 1000))
    ElMessage.success('异常检测完成，发现 2 个异常')
  } catch (error) {
    console.error('异常检测失败:', error)
  }
}

const retrainModel = async () => {
  try {
    ElMessage.info('模型正在重新训练...')
    await new Promise(resolve => setTimeout(resolve, 2000))
    modelPrecision.value = Math.min(99, modelPrecision.value + Math.random() * 2)
    trainingDataCount.value += Math.floor(Math.random() * 1000)
    lastTrainingTime.value = Date.now()
    ElMessage.success('模型训练完成，精度已提升')
  } catch (error) {
    console.error('模型训练失败:', error)
  }
}

watch(chartPeriod, () => {
  initTrafficChart()
})

onMounted(() => {
  initTrafficChart()
  initOccupancyChart()
})
</script>

<style scoped>
.traffic-prediction {
  height: 100%;
}

.stat-card {
  height: 160px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stat-value {
  font-size: 36px;
  font-weight: bold;
  color: #303133;
  margin: 10px 0;
}

.stat-change {
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 5px;
}

.stat-change.up {
  color: #f56c6c;
}

.stat-change.down {
  color: #67c23a;
}

.stat-time {
  font-size: 14px;
  color: #909399;
}

.stat-status {
  margin-top: 10px;
}

.chart-container {
  height: 350px;
  width: 100%;
}
</style>
