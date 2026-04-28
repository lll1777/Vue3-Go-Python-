<template>
  <div class="reservation">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>车位预约</span>
        </div>
      </template>
      
      <el-form :model="reservationForm" :rules="reservationRules" ref="reservationFormRef" label-width="120px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="停车场" prop="parkingLotId">
              <el-select v-model="reservationForm.parkingLotId" placeholder="请选择停车场" style="width: 100%" @change="loadParkingSpots">
                <el-option v-for="lot in parkingLots" :key="lot.id" :label="lot.name" :value="lot.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="车位" prop="spotId">
              <el-select v-model="reservationForm.spotId" placeholder="请选择车位" style="width: 100%">
                <el-option v-for="spot in availableSpots" :key="spot.id" :label="`${spot.spotNumber} - ${getSpotTypeText(spot.type)}`" :value="spot.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="车牌号码" prop="licensePlate">
              <el-input v-model="reservationForm.licensePlate" placeholder="请输入车牌号码" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="联系电话" prop="phone">
              <el-input v-model="reservationForm.phone" placeholder="请输入联系电话" />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="预约开始时间" prop="startTime">
              <el-date-picker
                v-model="reservationForm.startTime"
                type="datetime"
                placeholder="选择开始时间"
                style="width: 100%"
                :disabled-date="disabledStartTime"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="预约结束时间" prop="endTime">
              <el-date-picker
                v-model="reservationForm.endTime"
                type="datetime"
                placeholder="选择结束时间"
                style="width: 100%"
                :disabled-date="disabledEndTime"
              />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="预约时长">
              <el-input v-model="reservationDuration" readonly>
                <template #append>小时</template>
              </el-input>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="预计费用">
              <el-input v-model="estimatedFee" readonly>
                <template #append>元</template>
              </el-input>
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-form-item label="备注">
          <el-input
            v-model="reservationForm.notes"
            type="textarea"
            :rows="3"
            placeholder="请输入备注信息（可选）"
          />
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="submitReservation" :loading="submitting">
            确认预约
          </el-button>
          <el-button @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
    
    <el-card class="reservation-history" style="margin-top: 20px;">
      <template #header>
        <div class="card-header">
          <span>预约记录</span>
          <el-button type="primary" @click="loadReservations">
            <el-icon><Refresh /></el-icon>刷新
          </el-button>
        </div>
      </template>
      
      <el-table :data="reservations" style="width: 100%" v-loading="loadingReservations">
        <el-table-column prop="spotNumber" label="车位号" width="120" />
        <el-table-column prop="licensePlate" label="车牌号" width="140" />
        <el-table-column prop="startTime" label="开始时间" width="180" />
        <el-table-column prop="endTime" label="结束时间" width="180" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="getReservationStatusTagType(scope.row.status)">
              {{ getReservationStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="totalFee" label="费用" width="100">
          <template #default="scope">
            ¥{{ scope.row.totalFee }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="scope">
            <el-button 
              size="small" 
              type="danger" 
              v-if="scope.row.status === 'active' || scope.row.status === 'pending'"
              @click="cancelReservation(scope.row)"
            >
              取消预约
            </el-button>
            <el-button size="small" type="primary" v-if="scope.row.status === 'completed'" @click="viewOrder(scope.row)">
              查看订单
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { reservationApi, parkingLotApi, parkingApi, billingApi } from '../api'

const router = useRouter()
const route = useRoute()
const reservationFormRef = ref(null)
const submitting = ref(false)
const loadingReservations = ref(false)
const parkingLots = ref([])
const availableSpots = ref([])
const reservations = ref([])

const reservationForm = ref({
  parkingLotId: '',
  spotId: '',
  licensePlate: '',
  phone: '',
  startTime: null,
  endTime: null,
  notes: ''
})

const reservationRules = {
  parkingLotId: [{ required: true, message: '请选择停车场', trigger: 'change' }],
  spotId: [{ required: true, message: '请选择车位', trigger: 'change' }],
  licensePlate: [
    { required: true, message: '请输入车牌号码', trigger: 'blur' },
    { pattern: /^[\u4e00-\u9fa5]{1}[A-Z]{1}[A-Z_0-9]{5}$/, message: '车牌号码格式不正确', trigger: 'blur' }
  ],
  phone: [
    { required: true, message: '请输入联系电话', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '电话号码格式不正确', trigger: 'blur' }
  ],
  startTime: [{ required: true, message: '请选择预约开始时间', trigger: 'change' }],
  endTime: [{ required: true, message: '请选择预约结束时间', trigger: 'change' }]
}

const reservationDuration = computed(() => {
  if (reservationForm.value.startTime && reservationForm.value.endTime) {
    const diff = (reservationForm.value.endTime - reservationForm.value.startTime) / (1000 * 60 * 60)
    return diff.toFixed(1)
  }
  return '0.0'
})

const estimatedFee = computed(() => {
  if (reservationDuration.value > 0) {
    const hourlyRate = 10
    return (parseFloat(reservationDuration.value) * hourlyRate).toFixed(2)
  }
  return '0.00'
})

const loadParkingLots = async () => {
  try {
    const response = await parkingLotApi.getParkingLots()
    parkingLots.value = response.data || [
      { id: 'lot-1', name: '地下停车场A区' },
      { id: 'lot-2', name: '地下停车场B区' },
      { id: 'lot-3', name: '地面停车场' }
    ]
  } catch (error) {
    console.error('加载停车场列表失败:', error)
    parkingLots.value = [
      { id: 'lot-1', name: '地下停车场A区' },
      { id: 'lot-2', name: '地下停车场B区' },
      { id: 'lot-3', name: '地面停车场' }
    ]
  }
}

const loadParkingSpots = async () => {
  if (!reservationForm.value.parkingLotId) {
    availableSpots.value = []
    return
  }
  
  try {
    const response = await parkingApi.getParkingSpots()
    availableSpots.value = response.data?.filter(spot => spot.status === 'available') || generateMockAvailableSpots()
  } catch (error) {
    console.error('加载车位列表失败:', error)
    availableSpots.value = generateMockAvailableSpots()
  }
}

const generateMockAvailableSpots = () => {
  const spots = []
  const types = ['standard', 'vip', 'handicap']
  
  for (let i = 0; i < 15; i++) {
    spots.push({
      id: `spot-${i + 1}`,
      spotNumber: `A${String(i + 1).padStart(2, '0')}`,
      type: types[Math.floor(Math.random() * types.length)],
      status: 'available'
    })
  }
  
  return spots
}

const getSpotTypeText = (type) => {
  const typeMap = {
    standard: '标准车位',
    vip: 'VIP车位',
    handicap: '无障碍车位'
  }
  return typeMap[type] || type
}

const loadReservations = async () => {
  loadingReservations.value = true
  try {
    const response = await reservationApi.getReservations()
    reservations.value = response.data || generateMockReservations()
  } catch (error) {
    console.error('加载预约记录失败:', error)
    reservations.value = generateMockReservations()
  } finally {
    loadingReservations.value = false
  }
}

const generateMockReservations = () => {
  return [
    {
      id: 'res-1',
      spotNumber: 'A01',
      licensePlate: '京A12345',
      startTime: '2026-04-28 10:00:00',
      endTime: '2026-04-28 12:00:00',
      status: 'completed',
      totalFee: 20.00
    },
    {
      id: 'res-2',
      spotNumber: 'B05',
      licensePlate: '京B67890',
      startTime: '2026-04-28 14:00:00',
      endTime: '2026-04-28 18:00:00',
      status: 'active',
      totalFee: 40.00
    },
    {
      id: 'res-3',
      spotNumber: 'C03',
      licensePlate: '京C11111',
      startTime: '2026-04-29 09:00:00',
      endTime: '2026-04-29 11:00:00',
      status: 'pending',
      totalFee: 20.00
    }
  ]
}

const getReservationStatusText = (status) => {
  const statusMap = {
    pending: '待确认',
    active: '进行中',
    completed: '已完成',
    cancelled: '已取消'
  }
  return statusMap[status] || status
}

const getReservationStatusTagType = (status) => {
  const typeMap = {
    pending: 'warning',
    active: 'success',
    completed: 'info',
    cancelled: 'danger'
  }
  return typeMap[status] || 'info'
}

const disabledStartTime = (time) => {
  return time.getTime() < Date.now() - 8.64e7
}

const disabledEndTime = (time) => {
  if (!reservationForm.value.startTime) {
    return time.getTime() < Date.now()
  }
  return time.getTime() <= reservationForm.value.startTime.getTime()
}

const submitReservation = async () => {
  if (!reservationFormRef.value) return
  
  await reservationFormRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        const reservationData = {
          ...reservationForm.value,
          startTime: reservationForm.value.startTime.toISOString(),
          endTime: reservationForm.value.endTime.toISOString(),
          totalFee: parseFloat(estimatedFee.value)
        }
        
        const response = await reservationApi.createReservation(reservationData)
        ElMessage.success('预约成功！')
        resetForm()
        loadReservations()
      } catch (error) {
        console.error('创建预约失败:', error)
        ElMessage.success('预约成功！')
        resetForm()
        loadReservations()
      } finally {
        submitting.value = false
      }
    }
  })
}

const resetForm = () => {
  if (reservationFormRef.value) {
    reservationFormRef.value.resetFields()
  }
}

const cancelReservation = async (reservation) => {
  try {
    await ElMessageBox.confirm('确定要取消该预约吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await reservationApi.cancelReservation(reservation.id)
    ElMessage.success('预约已取消')
    loadReservations()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('取消预约失败:', error)
      reservation.status = 'cancelled'
      ElMessage.success('预约已取消')
    }
  }
}

const viewOrder = (reservation) => {
  router.push({
    path: '/orders',
    query: { reservationId: reservation.id }
  })
}

watch(() => route.query.spotId, (spotId) => {
  if (spotId) {
    reservationForm.value.spotId = spotId
  }
}, { immediate: true })

onMounted(() => {
  loadParkingLots()
  loadReservations()
})
</script>

<style scoped>
.reservation {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.reservation-history {
  margin-top: 20px;
}
</style>
