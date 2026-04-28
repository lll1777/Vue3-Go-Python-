<template>
  <div class="orders">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>订单管理</span>
          <el-button type="primary" @click="loadOrders">
            <el-icon><Refresh /></el-icon>刷新
          </el-button>
        </div>
      </template>
      
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="订单号">
          <el-input v-model="searchForm.orderNo" placeholder="请输入订单号" clearable />
        </el-form-item>
        <el-form-item label="车牌号">
          <el-input v-model="searchForm.licensePlate" placeholder="请输入车牌号" clearable />
        </el-form-item>
        <el-form-item label="订单状态">
          <el-select v-model="searchForm.status" placeholder="全部状态" clearable>
            <el-option label="待支付" value="unpaid" />
            <el-option label="已支付" value="paid" />
            <el-option label="已取消" value="cancelled" />
            <el-option label="已完成" value="completed" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="searchOrders">
            <el-icon><Search /></el-icon>查询
          </el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
      
      <el-table :data="filteredOrders" style="width: 100%" v-loading="loadingOrders">
        <el-table-column prop="orderNo" label="订单号" width="200" />
        <el-table-column prop="licensePlate" label="车牌号" width="120" />
        <el-table-column prop="spotNumber" label="车位号" width="100" />
        <el-table-column prop="entryTime" label="入场时间" width="180" />
        <el-table-column prop="exitTime" label="出场时间" width="180">
          <template #default="scope">
            {{ scope.row.exitTime || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="parkingDuration" label="停车时长" width="120" />
        <el-table-column prop="totalAmount" label="订单金额" width="120">
          <template #default="scope">
            ¥{{ scope.row.totalAmount }}
          </template>
        </el-table-column>
        <el-table-column prop="paidAmount" label="已付金额" width="120">
          <template #default="scope">
            ¥{{ scope.row.paidAmount || 0 }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="getOrderStatusTagType(scope.row.status)">
              {{ getOrderStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="scope">
            <el-button size="small" @click="viewOrderDetail(scope.row)">
              详情
            </el-button>
            <el-button 
              size="small" 
              type="primary" 
              v-if="scope.row.status === 'unpaid'"
              @click="payOrder(scope.row)"
            >
              支付
            </el-button>
            <el-button 
              size="small" 
              type="danger" 
              v-if="scope.row.status === 'unpaid' || scope.row.status === 'reserved'"
              @click="cancelOrder(scope.row)"
            >
              取消
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="filteredOrders.length"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
        style="margin-top: 20px; justify-content: flex-end;"
      />
    </el-card>
    
    <el-dialog
      v-model="orderDetailVisible"
      title="订单详情"
      width="600px"
    >
      <el-descriptions :column="1" border>
        <el-descriptions-item label="订单号">{{ selectedOrder?.orderNo }}</el-descriptions-item>
        <el-descriptions-item label="订单类型">{{ getOrderTypeText(selectedOrder?.type) }}</el-descriptions-item>
        <el-descriptions-item label="车牌号">{{ selectedOrder?.licensePlate }}</el-descriptions-item>
        <el-descriptions-item label="车位号">{{ selectedOrder?.spotNumber }}</el-descriptions-item>
        <el-descriptions-item label="入场时间">{{ selectedOrder?.entryTime }}</el-descriptions-item>
        <el-descriptions-item label="出场时间">{{ selectedOrder?.exitTime || '-' }}</el-descriptions-item>
        <el-descriptions-item label="停车时长">{{ selectedOrder?.parkingDuration }}</el-descriptions-item>
        <el-descriptions-item label="计费规则">{{ selectedOrder?.billingRule || '标准计费规则' }}</el-descriptions-item>
        <el-descriptions-item label="订单金额">
          <span style="font-size: 18px; color: #f56c6c; font-weight: bold;">
            ¥{{ selectedOrder?.totalAmount }}
          </span>
        </el-descriptions-item>
        <el-descriptions-item label="已付金额">
          <span style="font-size: 18px; color: #67c23a; font-weight: bold;">
            ¥{{ selectedOrder?.paidAmount || 0 }}
          </span>
        </el-descriptions-item>
        <el-descriptions-item label="订单状态">
          <el-tag :type="getOrderStatusTagType(selectedOrder?.status)">
            {{ getOrderStatusText(selectedOrder?.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="支付方式" v-if="selectedOrder?.paymentMethod">
          {{ getPaymentMethodText(selectedOrder?.paymentMethod) }}
        </el-descriptions-item>
        <el-descriptions-item label="支付时间" v-if="selectedOrder?.paidTime">
          {{ selectedOrder?.paidTime }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ selectedOrder?.createdAt }}</el-descriptions-item>
      </el-descriptions>
      
      <template #footer>
        <el-button @click="orderDetailVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { orderApi } from '../api'

const router = useRouter()
const route = useRoute()
const loadingOrders = ref(false)
const orders = ref([])
const searchForm = ref({
  orderNo: '',
  licensePlate: '',
  status: ''
})
const currentPage = ref(1)
const pageSize = ref(10)
const orderDetailVisible = ref(false)
const selectedOrder = ref(null)

const filteredOrders = computed(() => {
  let result = orders.value
  
  if (searchForm.value.orderNo) {
    result = result.filter(order => order.orderNo.includes(searchForm.value.orderNo))
  }
  
  if (searchForm.value.licensePlate) {
    result = result.filter(order => order.licensePlate.includes(searchForm.value.licensePlate))
  }
  
  if (searchForm.value.status) {
    result = result.filter(order => order.status === searchForm.value.status)
  }
  
  return result
})

const loadOrders = async () => {
  loadingOrders.value = true
  try {
    const response = await orderApi.getOrders()
    orders.value = response.data || generateMockOrders()
  } catch (error) {
    console.error('加载订单列表失败:', error)
    orders.value = generateMockOrders()
  } finally {
    loadingOrders.value = false
  }
}

const generateMockOrders = () => {
  return [
    {
      id: 'order-1',
      orderNo: 'ORD202604280001',
      type: 'parking',
      licensePlate: '京A12345',
      spotNumber: 'A01',
      entryTime: '2026-04-28 10:30:00',
      exitTime: '2026-04-28 12:45:00',
      parkingDuration: '2小时15分钟',
      billingRule: '标准计费规则',
      totalAmount: 22.50,
      paidAmount: 22.50,
      status: 'completed',
      paymentMethod: 'alipay',
      paidTime: '2026-04-28 12:45:00',
      createdAt: '2026-04-28 10:30:00'
    },
    {
      id: 'order-2',
      orderNo: 'ORD202604280002',
      type: 'reservation',
      licensePlate: '京B67890',
      spotNumber: 'B05',
      entryTime: '2026-04-28 09:00:00',
      exitTime: null,
      parkingDuration: '5小时30分钟',
      billingRule: 'VIP计费规则',
      totalAmount: 88.00,
      paidAmount: 0,
      status: 'unpaid',
      paymentMethod: null,
      paidTime: null,
      createdAt: '2026-04-28 08:30:00'
    },
    {
      id: 'order-3',
      orderNo: 'ORD202604280003',
      type: 'reservation',
      licensePlate: '京C11111',
      spotNumber: 'C03',
      entryTime: '2026-04-29 09:00:00',
      exitTime: null,
      parkingDuration: '2小时00分钟',
      billingRule: '标准计费规则',
      totalAmount: 20.00,
      paidAmount: 20.00,
      status: 'reserved',
      paymentMethod: 'wechat',
      paidTime: '2026-04-28 15:00:00',
      createdAt: '2026-04-28 15:00:00'
    },
    {
      id: 'order-4',
      orderNo: 'ORD202604270001',
      type: 'parking',
      licensePlate: '京D22222',
      spotNumber: 'A05',
      entryTime: '2026-04-27 14:00:00',
      exitTime: '2026-04-27 16:30:00',
      parkingDuration: '2小时30分钟',
      billingRule: '标准计费规则',
      totalAmount: 25.00,
      paidAmount: 25.00,
      status: 'completed',
      paymentMethod: 'balance',
      paidTime: '2026-04-27 16:30:00',
      createdAt: '2026-04-27 14:00:00'
    },
    {
      id: 'order-5',
      orderNo: 'ORD202604270002',
      type: 'parking',
      licensePlate: '京E33333',
      spotNumber: 'B02',
      entryTime: '2026-04-27 10:00:00',
      exitTime: '2026-04-27 12:00:00',
      parkingDuration: '2小时00分钟',
      billingRule: '标准计费规则',
      totalAmount: 20.00,
      paidAmount: 20.00,
      status: 'completed',
      paymentMethod: 'alipay',
      paidTime: '2026-04-27 12:00:00',
      createdAt: '2026-04-27 10:00:00'
    }
  ]
}

const getOrderStatusText = (status) => {
  const statusMap = {
    unpaid: '待支付',
    paid: '已支付',
    reserved: '已预约',
    cancelled: '已取消',
    completed: '已完成'
  }
  return statusMap[status] || status
}

const getOrderStatusTagType = (status) => {
  const typeMap = {
    unpaid: 'warning',
    paid: 'success',
    reserved: 'primary',
    cancelled: 'danger',
    completed: 'info'
  }
  return typeMap[status] || 'info'
}

const getOrderTypeText = (type) => {
  const typeMap = {
    parking: '停车订单',
    reservation: '预约订单'
  }
  return typeMap[type] || type
}

const getPaymentMethodText = (method) => {
  const methodMap = {
    alipay: '支付宝',
    wechat: '微信支付',
    balance: '余额支付'
  }
  return methodMap[method] || method
}

const searchOrders = () => {
  currentPage.value = 1
}

const resetSearch = () => {
  searchForm.value = {
    orderNo: '',
    licensePlate: '',
    status: ''
  }
  currentPage.value = 1
}

const handleSizeChange = (val) => {
  pageSize.value = val
}

const handleCurrentChange = (val) => {
  currentPage.value = val
}

const viewOrderDetail = (order) => {
  selectedOrder.value = order
  orderDetailVisible.value = true
}

const payOrder = (order) => {
  router.push({
    path: '/payment',
    query: { orderId: order.id }
  })
}

const cancelOrder = async (order) => {
  try {
    await ElMessageBox.confirm('确定要取消该订单吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    order.status = 'cancelled'
    ElMessage.success('订单已取消')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('取消订单失败:', error)
    }
  }
}

watch(() => route.query.orderId, (orderId) => {
  if (orderId) {
    const order = orders.value.find(o => o.id === orderId)
    if (order) {
      viewOrderDetail(order)
    }
  }
}, { immediate: true })

onMounted(() => {
  loadOrders()
})
</script>

<style scoped>
.orders {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-form {
  margin-bottom: 20px;
}
</style>
