<template>
  <div class="parking-map">
    <el-card class="map-header">
      <template #header>
        <div class="card-header">
          <span>停车场地图</span>
          <el-button type="primary" @click="loadParkingSpots">
            <el-icon><Refresh /></el-icon>刷新
          </el-button>
        </div>
      </template>
      <div class="stats-container">
        <el-statistic title="总车位" :value="totalSpots" />
        <el-statistic title="空闲车位" :value="availableSpots" :value-style="{ color: '#67c23a' }" />
        <el-statistic title="已占用" :value="occupiedSpots" :value-style="{ color: '#f56c6c' }" />
        <el-statistic title="已预约" :value="reservedSpots" :value-style="{ color: '#e6a23c' }" />
      </div>
    </el-card>

    <el-row :gutter="20">
      <el-col :span="18">
        <el-card class="map-container">
          <template #header>
            <div class="card-header">
              <span>车位分布</span>
              <div class="legend">
                <span class="legend-item">
                  <span class="legend-color available"></span>空闲
                </span>
                <span class="legend-item">
                  <span class="legend-color occupied"></span>占用
                </span>
                <span class="legend-item">
                  <span class="legend-color reserved"></span>预约
                </span>
                <span class="legend-item">
                  <span class="legend-color maintenance"></span>维护
                </span>
              </div>
            </div>
          </template>
          <div class="parking-grid">
            <div v-for="spot in parkingSpots" :key="spot.id" class="parking-spot" :class="spot.status" @click="showSpotDetails(spot)">
              <div class="spot-number">{{ spot.spotNumber }}</div>
              <div class="spot-type">{{ spot.type === 'standard' ? '标准' : spot.type === 'vip' ? 'VIP' : '无障碍' }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="spot-details">
          <template #header>
            <span>车位详情</span>
          </template>
          <el-empty v-if="!selectedSpot" description="点击车位查看详情" />
          <div v-else class="spot-info">
            <el-descriptions :column="1" border>
              <el-descriptions-item label="车位编号">{{ selectedSpot.spotNumber }}</el-descriptions-item>
              <el-descriptions-item label="区域">{{ selectedSpot.zone }}</el-descriptions-item>
              <el-descriptions-item label="楼层">{{ selectedSpot.floor }} 层</el-descriptions-item>
              <el-descriptions-item label="类型">{{ selectedSpot.type === 'standard' ? '标准车位' : selectedSpot.type === 'vip' ? 'VIP车位' : '无障碍车位' }}</el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag :type="getStatusTagType(selectedSpot.status)">{{ getStatusText(selectedSpot.status) }}</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="当前车辆" v-if="selectedSpot.currentVehicle">
                {{ selectedSpot.currentVehicle.licensePlate }}
              </el-descriptions-item>
              <el-descriptions-item label="预计离开时间" v-if="selectedSpot.expectedLeaveTime">
                {{ selectedSpot.expectedLeaveTime }}
              </el-descriptions-item>
            </el-descriptions>
            <div class="spot-actions">
              <el-button 
                type="primary" 
                v-if="selectedSpot.status === 'available'" 
                @click="reserveSpot(selectedSpot)"
              >
                立即预约
              </el-button>
              <el-button 
                type="warning" 
                v-if="selectedSpot.status === 'occupied'" 
                @click="releaseSpot(selectedSpot)"
              >
                释放车位
              </el-button>
              <el-button 
                type="info" 
                v-if="selectedSpot.status === 'maintenance'" 
                @click="maintenanceComplete(selectedSpot)"
              >
                维护完成
              </el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { parkingApi } from '../api'

const router = useRouter()
const parkingSpots = ref([])
const selectedSpot = ref(null)

const totalSpots = computed(() => parkingSpots.value.length)
const availableSpots = computed(() => parkingSpots.value.filter(spot => spot.status === 'available').length)
const occupiedSpots = computed(() => parkingSpots.value.filter(spot => spot.status === 'occupied').length)
const reservedSpots = computed(() => parkingSpots.value.filter(spot => spot.status === 'reserved').length)

const loadParkingSpots = async () => {
  try {
    const response = await parkingApi.getParkingSpots()
    parkingSpots.value = response.data || []
  } catch (error) {
    console.error('加载车位信息失败:', error)
    parkingSpots.value = generateMockSpots()
  }
}

const generateMockSpots = () => {
  const zones = ['A区', 'B区', 'C区', 'D区']
  const types = ['standard', 'standard', 'standard', 'vip', 'handicap']
  const statuses = ['available', 'available', 'occupied', 'reserved', 'available']
  const spots = []
  
  for (let i = 0; i < 48; i++) {
    const zone = zones[Math.floor(i / 12)]
    const floor = Math.floor(Math.random() * 3) + 1
    const type = types[Math.floor(Math.random() * types.length)]
    const status = statuses[Math.floor(Math.random() * statuses.length)]
    
    spots.push({
      id: `spot-${i + 1}`,
      spotNumber: `${zone.charAt(0)}${String(i % 12 + 1).padStart(2, '0')}`,
      zone: zone,
      floor: floor,
      type: type,
      status: status,
      currentVehicle: status === 'occupied' ? {
        licensePlate: `京A${String(Math.floor(Math.random() * 100000)).padStart(5, '0')}`
      } : null,
      expectedLeaveTime: status === 'occupied' ? '2026-04-28 18:00' : null
    })
  }
  
  return spots
}

const showSpotDetails = (spot) => {
  selectedSpot.value = spot
}

const getStatusText = (status) => {
  const statusMap = {
    available: '空闲',
    occupied: '已占用',
    reserved: '已预约',
    maintenance: '维护中'
  }
  return statusMap[status] || status
}

const getStatusTagType = (status) => {
  const typeMap = {
    available: 'success',
    occupied: 'danger',
    reserved: 'warning',
    maintenance: 'info'
  }
  return typeMap[status] || 'info'
}

const reserveSpot = (spot) => {
  router.push({
    path: '/reservation',
    query: { spotId: spot.id }
  })
}

const releaseSpot = async (spot) => {
  try {
    await parkingApi.updateParkingSpotStatus(spot.id, 'available')
    spot.status = 'available'
    ElMessage.success('车位已释放')
  } catch (error) {
    console.error('释放车位失败:', error)
  }
}

const maintenanceComplete = async (spot) => {
  try {
    await parkingApi.updateParkingSpotStatus(spot.id, 'available')
    spot.status = 'available'
    ElMessage.success('维护已完成')
  } catch (error) {
    console.error('更新维护状态失败:', error)
  }
}

onMounted(() => {
  loadParkingSpots()
})
</script>

<style scoped>
.parking-map {
  height: 100%;
}

.map-header {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stats-container {
  display: flex;
  justify-content: space-around;
  margin-top: 10px;
}

.legend {
  display: flex;
  gap: 20px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 14px;
}

.legend-color {
  width: 16px;
  height: 16px;
  border-radius: 3px;
}

.legend-color.available {
  background-color: #67c23a;
}

.legend-color.occupied {
  background-color: #f56c6c;
}

.legend-color.reserved {
  background-color: #e6a23c;
}

.legend-color.maintenance {
  background-color: #909399;
}

.map-container {
  margin-bottom: 20px;
}

.parking-grid {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 10px;
}

.parking-spot {
  padding: 15px 10px;
  border-radius: 8px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
  border: 2px solid #dcdfe6;
}

.parking-spot:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.parking-spot.available {
  background-color: #f0f9eb;
  border-color: #67c23a;
}

.parking-spot.occupied {
  background-color: #fef0f0;
  border-color: #f56c6c;
}

.parking-spot.reserved {
  background-color: #fdf6ec;
  border-color: #e6a23c;
}

.parking-spot.maintenance {
  background-color: #f4f4f5;
  border-color: #909399;
}

.spot-number {
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 5px;
}

.spot-type {
  font-size: 12px;
  color: #909399;
}

.spot-details {
  height: 100%;
}

.spot-info {
  margin-top: 10px;
}

.spot-actions {
  margin-top: 20px;
  text-align: center;
}
</style>
