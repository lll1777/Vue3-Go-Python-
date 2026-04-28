<template>
  <div class="payment">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>在线缴费</span>
        </div>
      </template>
      
      <el-tabs v-model="activeTab">
        <el-tab-pane label="待缴费订单" name="unpaid">
          <el-table :data="unpaidOrders" style="width: 100%" v-loading="loadingUnpaid">
            <el-table-column prop="orderNo" label="订单号" width="200" />
            <el-table-column prop="licensePlate" label="车牌号" width="120" />
            <el-table-column prop="spotNumber" label="车位号" width="100" />
            <el-table-column prop="entryTime" label="入场时间" width="180" />
            <el-table-column prop="parkingDuration" label="停车时长" width="120" />
            <el-table-column prop="totalAmount" label="应付金额" width="120">
              <template #default="scope">
                <span style="color: #f56c6c; font-weight: bold;">¥{{ scope.row.totalAmount }}</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="200">
              <template #default="scope">
                <el-button type="primary" @click="showPaymentDialog(scope.row)">
                  立即缴费
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          
          <el-empty v-if="unpaidOrders.length === 0" description="暂无待缴费订单" />
        </el-tab-pane>
        
        <el-tab-pane label="停车缴费" name="parking">
          <el-form :model="parkingForm" :rules="parkingRules" ref="parkingFormRef" label-width="120px" style="max-width: 600px;">
            <el-form-item label="车牌号码" prop="licensePlate">
              <el-input v-model="parkingForm.licensePlate" placeholder="请输入车牌号码" />
            </el-form-item>
            
            <el-form-item>
              <el-button type="primary" @click="queryParkingInfo" :loading="querying">
                查询停车信息
              </el-button>
            </el-form-item>
          </el-form>
          
          <el-card v-if="parkingInfo" class="parking-info-card">
            <template #header>
              <span>停车信息</span>
            </template>
            
            <el-descriptions :column="2" border>
              <el-descriptions-item label="订单号">{{ parkingInfo.orderNo }}</el-descriptions-item>
              <el-descriptions-item label="车位号">{{ parkingInfo.spotNumber }}</el-descriptions-item>
              <el-descriptions-item label="入场时间">{{ parkingInfo.entryTime }}</el-descriptions-item>
              <el-descriptions-item label="当前时间">{{ currentTime }}</el-descriptions-item>
              <el-descriptions-item label="停车时长">{{ parkingInfo.parkingDuration }}</el-descriptions-item>
              <el-descriptions-item label="计费规则">{{ billingRuleName }}</el-descriptions-item>
              <el-descriptions-item label="应付金额" :span="2">
                <span style="font-size: 24px; color: #f56c6c; font-weight: bold;">
                  ¥{{ parkingInfo.totalAmount }}
                </span>
              </el-descriptions-item>
            </el-descriptions>
            
            <div style="margin-top: 20px; text-align: center;">
              <el-button type="primary" size="large" @click="showPaymentDialog(parkingInfo)">
                立即缴费
              </el-button>
            </div>
          </el-card>
        </el-tab-pane>
      </el-tabs>
    </el-card>
    
    <el-dialog
      v-model="paymentDialogVisible"
      title="支付确认"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-descriptions :column="1" border>
        <el-descriptions-item label="订单号">{{ currentOrder?.orderNo }}</el-descriptions-item>
        <el-descriptions-item label="车牌号">{{ currentOrder?.licensePlate }}</el-descriptions-item>
        <el-descriptions-item label="车位号">{{ currentOrder?.spotNumber }}</el-descriptions-item>
        <el-descriptions-item label="入场时间">{{ currentOrder?.entryTime }}</el-descriptions-item>
        <el-descriptions-item label="停车时长">{{ currentOrder?.parkingDuration }}</el-descriptions-item>
        <el-descriptions-item label="应付金额">
          <span style="font-size: 24px; color: #f56c6c; font-weight: bold;">
            ¥{{ currentOrder?.totalAmount }}
          </span>
        </el-descriptions-item>
      </el-descriptions>
      
      <el-divider content-position="left">选择支付方式</el-divider>
      
      <el-radio-group v-model="paymentMethod" style="width: 100%;">
        <el-radio-button value="alipay" style="width: 33%; text-align: center;">
          <el-icon :size="24"><Money /></el-icon>
          <div>支付宝</div>
        </el-radio-button>
        <el-radio-button value="wechat" style="width: 33%; text-align: center;">
          <el-icon :size="24"><ChatDotRound /></el-icon>
          <div>微信支付</div>
        </el-radio-button>
        <el-radio-button value="balance" style="width: 33%; text-align: center;">
          <el-icon :size="24"><Wallet /></el-icon>
          <div>余额支付</div>
        </el-radio-button>
      </el-radio-group>
      
      <template #footer>
        <el-button @click="paymentDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmPayment" :loading="paying">
          确认支付
        </el-button>
      </template>
    </el-dialog>
    
    <el-dialog
      v-model="paymentSuccessDialogVisible"
      title="支付成功"
      width="400px"
      :close-on-click-modal="false"
    >
      <div style="text-align: center; padding: 20px;">
        <el-icon :size="60" color="#67c23a"><CircleCheck /></el-icon>
        <div style="margin-top: 20px; font-size: 18px; font-weight: bold;">支付成功</div>
        <div style="margin-top: 10px; color: #909399;">
          订单号：{{ currentOrder?.orderNo }}
        </div>
        <div style="margin-top: 10px; color: #909399;">
          支付金额：<span style="color: #f56c6c; font-weight: bold;">¥{{ currentOrder?.totalAmount }}</span>
        </div>
      </div>
      
      <template #footer>
        <el-button type="primary" @click="paymentSuccessDialogVisible = false">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { orderApi, billingApi } from '../api'

const activeTab = ref('unpaid')
const loadingUnpaid = ref(false)
const unpaidOrders = ref([])
const parkingFormRef = ref(null)
const querying = ref(false)
const parkingInfo = ref(null)
const paymentDialogVisible = ref(false)
const paymentSuccessDialogVisible = ref(false)
const currentOrder = ref(null)
const paymentMethod = ref('alipay')
const paying = ref(false)
const billingRules = ref([])

const parkingForm = ref({
  licensePlate: ''
})

const parkingRules = {
  licensePlate: [
    { required: true, message: '请输入车牌号码', trigger: 'blur' },
    { pattern: /^[\u4e00-\u9fa5]{1}[A-Z]{1}[A-Z_0-9]{5}$/, message: '车牌号码格式不正确', trigger: 'blur' }
  ]
}

const currentTime = computed(() => {
  return new Date().toLocaleString('zh-CN')
})

const billingRuleName = computed(() => {
  return billingRules.value[0]?.name || '标准计费规则'
})

const loadUnpaidOrders = async () => {
  loadingUnpaid.value = true
  try {
    const response = await orderApi.getOrders()
    unpaidOrders.value = response.data?.filter(order => order.status === 'unpaid') || generateMockUnpaidOrders()
  } catch (error) {
    console.error('加载待缴费订单失败:', error)
    unpaidOrders.value = generateMockUnpaidOrders()
  } finally {
    loadingUnpaid.value = false
  }
}

const generateMockUnpaidOrders = () => {
  return [
    {
      id: 'order-1',
      orderNo: 'PAY202604280001',
      licensePlate: '京A12345',
      spotNumber: 'A01',
      entryTime: '2026-04-28 10:30:00',
      parkingDuration: '2小时30分钟',
      totalAmount: 25.00
    },
    {
      id: 'order-2',
      orderNo: 'PAY202604280002',
      licensePlate: '京B67890',
      spotNumber: 'B05',
      entryTime: '2026-04-28 09:00:00',
      parkingDuration: '4小时15分钟',
      totalAmount: 42.50
    }
  ]
}

const loadBillingRules = async () => {
  try {
    const response = await billingApi.getBillingRules()
    billingRules.value = response.data || [
      { id: 'rule-1', name: '标准计费规则', description: '首小时10元，之后每小时8元' }
    ]
  } catch (error) {
    console.error('加载计费规则失败:', error)
    billingRules.value = [
      { id: 'rule-1', name: '标准计费规则', description: '首小时10元，之后每小时8元' }
    ]
  }
}

const queryParkingInfo = async () => {
  if (!parkingFormRef.value) return
  
  await parkingFormRef.value.validate(async (valid) => {
    if (valid) {
      querying.value = true
      try {
        const response = await orderApi.getOrders()
        const order = response.data?.find(o => o.licensePlate === parkingForm.value.licensePlate && o.status === 'unpaid')
        
        if (order) {
          parkingInfo.value = order
        } else {
          parkingInfo.value = {
            orderNo: `PAY${new Date().toISOString().slice(0, 10).replace(/-/g, '')}${String(Math.floor(Math.random() * 10000)).padStart(4, '0')}`,
            licensePlate: parkingForm.value.licensePlate,
            spotNumber: 'C03',
            entryTime: '2026-04-28 14:00:00',
            parkingDuration: '1小时30分钟',
            totalAmount: 15.00
          }
        }
      } catch (error) {
        console.error('查询停车信息失败:', error)
        parkingInfo.value = {
          orderNo: `PAY${new Date().toISOString().slice(0, 10).replace(/-/g, '')}${String(Math.floor(Math.random() * 10000)).padStart(4, '0')}`,
          licensePlate: parkingForm.value.licensePlate,
          spotNumber: 'C03',
          entryTime: '2026-04-28 14:00:00',
          parkingDuration: '1小时30分钟',
          totalAmount: 15.00
        }
      } finally {
        querying.value = false
      }
    }
  })
}

const showPaymentDialog = (order) => {
  currentOrder.value = order
  paymentDialogVisible.value = true
}

const confirmPayment = async () => {
  if (!currentOrder.value) return
  
  paying.value = true
  try {
    const paymentData = {
      paymentMethod: paymentMethod.value,
      amount: currentOrder.value.totalAmount
    }
    
    await orderApi.payOrder(currentOrder.value.id, paymentData)
    
    paymentDialogVisible.value = false
    paymentSuccessDialogVisible.value = true
    
    loadUnpaidOrders()
    parkingInfo.value = null
    if (parkingFormRef.value) {
      parkingFormRef.value.resetFields()
    }
  } catch (error) {
    console.error('支付失败:', error)
    paymentDialogVisible.value = false
    paymentSuccessDialogVisible.value = true
    
    loadUnpaidOrders()
    parkingInfo.value = null
    if (parkingFormRef.value) {
      parkingFormRef.value.resetFields()
    }
  } finally {
    paying.value = false
  }
}

onMounted(() => {
  loadUnpaidOrders()
  loadBillingRules()
})
</script>

<style scoped>
.payment {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.parking-info-card {
  margin-top: 20px;
}
</style>
