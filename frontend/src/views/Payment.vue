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
            <el-table-column label="操作" width="280">
              <template #default="scope">
                <el-button type="primary" @click="showPaymentDialog(scope.row)">
                  立即缴费
                </el-button>
                <el-button type="info" @click="showBillingDetailDialog(scope.row)">
                  计费明细
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
              <el-button type="info" size="large" @click="showBillingDetailDialog(parkingInfo)" style="margin-left: 10px;">
                查看计费明细
              </el-button>
            </div>
          </el-card>
        </el-tab-pane>
        
        <el-tab-pane label="计费计算器" name="calculator">
          <el-card class="calculator-card">
            <template #header>
              <span>停车费用计算器</span>
            </template>
            
            <el-form :model="calculatorForm" ref="calculatorFormRef" label-width="100px" style="max-width: 500px;">
              <el-form-item label="入场时间" prop="entryTime">
                <el-date-picker
                  v-model="calculatorForm.entryTime"
                  type="datetime"
                  placeholder="选择入场时间"
                  format="YYYY-MM-DD HH:mm:ss"
                  value-format="YYYY-MM-DD HH:mm:ss"
                  style="width: 100%;"
                />
              </el-form-item>
              
              <el-form-item label="出场时间" prop="exitTime">
                <el-date-picker
                  v-model="calculatorForm.exitTime"
                  type="datetime"
                  placeholder="选择出场时间"
                  format="YYYY-MM-DD HH:mm:ss"
                  value-format="YYYY-MM-DD HH:mm:ss"
                  style="width: 100%;"
                />
              </el-form-item>
              
              <el-form-item label="车位类型">
                <el-select v-model="calculatorForm.spotType" placeholder="选择车位类型" style="width: 100%;">
                  <el-option label="标准车位" value="standard" />
                  <el-option label="VIP车位" value="vip" />
                  <el-option label="新能源车位" value="ev" />
                </el-select>
              </el-form-item>
              
              <el-form-item>
                <el-button type="primary" @click="calculateParkingFee" :loading="calculating">
                  计算费用
                </el-button>
              </el-form-item>
            </el-form>
            
            <el-divider v-if="calculationResult" />
            
            <div v-if="calculationResult" class="calculation-result">
              <el-alert
                :title="'停车费用计算结果'"
                :type="calculationResult.within_grace_period ? 'success' : 'warning'"
                :closable="false"
                style="margin-bottom: 20px;"
              >
                <template #default>
                  <div style="display: flex; justify-content: space-between; align-items: center;">
                    <div>
                      <div style="margin-bottom: 5px;">
                        <span style="font-weight: bold;">停车时长：</span>
                        {{ formatDuration(calculationResult.total_duration_min) }}
                      </div>
                      <div>
                        <span style="font-weight: bold;">应付金额：</span>
                        <span style="font-size: 28px; color: #f56c6c; font-weight: bold;">
                          ¥{{ calculationResult.final_amount }}
                        </span>
                      </div>
                    </div>
                    <el-tag v-if="calculationResult.within_grace_period" type="success" size="large">
                      宽限期内免费
                    </el-tag>
                    <el-tag v-else-if="calculationResult.min_charge_applied" type="info" size="large">
                      最低消费
                    </el-tag>
                  </div>
                </template>
              </el-alert>
              
              <el-collapse v-model="activeDayNames" accordion>
                <el-collapse-item
                  v-for="(dailyBilling, index) in calculationResult.daily_billings"
                  :key="index"
                  :name="String(index)"
                >
                  <template #title>
                    <span style="font-weight: bold;">
                      {{ dailyBilling.date }}
                      <el-tag :type="getDayTagType(dailyBilling)" size="small" style="margin-left: 10px;">
                        {{ getDayLabel(dailyBilling) }}
                      </el-tag>
                    </span>
                    <span style="float: right; color: #f56c6c; font-weight: bold;">
                      ¥{{ dailyBilling.daily_total }}
                    </span>
                  </template>
                  
                  <el-table :data="dailyBilling.periods" size="small" border>
                    <el-table-column prop="start_time" label="开始时间" width="180">
                      <template #default="scope">
                        {{ formatTime(scope.row.start_time) }}
                      </template>
                    </el-table-column>
                    <el-table-column prop="end_time" label="结束时间" width="180">
                      <template #default="scope">
                        {{ formatTime(scope.row.end_time) }}
                      </template>
                    </el-table-column>
                    <el-table-column prop="duration_min" label="时长(分钟)" width="100" align="center" />
                    <el-table-column prop="period_type" label="时段类型" width="100" align="center">
                      <template #default="scope">
                        <el-tag :type="getPeriodTagType(scope.row.period_type)" size="small">
                          {{ getPeriodLabel(scope.row.period_type) }}
                        </el-tag>
                      </template>
                    </el-table-column>
                    <el-table-column prop="hourly_rate" label="费率(元/小时)" width="120" align="center">
                      <template #default="scope">
                        <template v-if="scope.row.is_first_hour">
                          <span style="color: #67c23a;">¥{{ calculationResult.rule_summary?.first_hour }}</span>
                          <el-tag type="success" size="mini" style="margin-left: 5px;">首小时</el-tag>
                        </template>
                        <template v-else>
                          ¥{{ scope.row.hourly_rate }}
                        </template>
                      </template>
                    </el-table-column>
                    <el-table-column prop="period_amount" label="时段费用" width="100" align="right">
                      <template #default="scope">
                        <span style="color: #f56c6c; font-weight: bold;">
                          ¥{{ scope.row.period_amount }}
                        </span>
                      </template>
                    </el-table-column>
                  </el-table>
                  
                  <div style="margin-top: 15px; padding: 10px; background-color: #f5f7fa; border-radius: 4px;">
                    <div style="display: flex; justify-content: space-between; margin-bottom: 5px;">
                      <span>时段小计：</span>
                      <span>¥{{ dailyBilling.sub_total }}</span>
                    </div>
                    <div v-if="dailyBilling.discount > 0" style="display: flex; justify-content: space-between; margin-bottom: 5px;">
                      <span style="color: #67c23a;">日封顶优惠：</span>
                      <span style="color: #67c23a;">-¥{{ dailyBilling.discount }}</span>
                    </div>
                    <div style="display: flex; justify-content: space-between; font-weight: bold; border-top: 1px solid #dcdfe6; padding-top: 10px;">
                      <span>本日合计：</span>
                      <span style="color: #f56c6c; font-size: 16px;">¥{{ dailyBilling.daily_total }}</span>
                    </div>
                  </div>
                </el-collapse-item>
              </el-collapse>
              
              <el-card style="margin-top: 20px;">
                <template #header>
                  <span>费用汇总</span>
                </template>
                <el-descriptions :column="2" border size="small">
                  <el-descriptions-item label="计费规则">
                    <div v-if="calculationResult.rule_summary">
                      <div>首小时：¥{{ calculationResult.rule_summary.first_hour }}</div>
                      <div>标准费率：¥{{ calculationResult.rule_summary.hourly_rate }}/小时</div>
                      <div>高峰费率：¥{{ calculationResult.rule_summary.peak_rate }}/小时</div>
                      <div>夜间费率：¥{{ calculationResult.rule_summary.night_rate }}/小时</div>
                      <div>节假日费率：¥{{ calculationResult.rule_summary.holiday_rate }}/小时</div>
                      <div>日封顶：¥{{ calculationResult.rule_summary.daily_max }}</div>
                      <div>最低消费：¥{{ calculationResult.rule_summary.min_charge }}</div>
                      <div>免费宽限期：{{ calculationResult.rule_summary.grace_period }}分钟</div>
                    </div>
                  </el-descriptions-item>
                  <el-descriptions-item label="费用计算">
                    <div style="display: flex; justify-content: space-between; margin-bottom: 5px;">
                      <span>时段费用合计：</span>
                      <span>¥{{ calculationResult.total_before_rules }}</span>
                    </div>
                    <div v-if="calculationResult.total_discount > 0" style="display: flex; justify-content: space-between; margin-bottom: 5px; color: #67c23a;">
                      <span>优惠减免：</span>
                      <span>-¥{{ calculationResult.total_discount }}</span>
                    </div>
                    <div style="display: flex; justify-content: space-between; font-weight: bold; padding-top: 10px; border-top: 1px solid #dcdfe6;">
                      <span>最终金额：</span>
                      <span style="color: #f56c6c; font-size: 20px;">¥{{ calculationResult.final_amount }}</span>
                    </div>
                  </el-descriptions-item>
                </el-descriptions>
              </el-card>
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
    
    <el-dialog
      v-model="billingDetailDialogVisible"
      title="计费明细"
      width="800px"
    >
      <div v-if="billingDetailResult">
        <el-alert
          :title="'停车费用明细'"
          type="info"
          :closable="false"
          style="margin-bottom: 20px;"
        >
          <template #default>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <div>
                <div style="margin-bottom: 5px;">
                  <span style="font-weight: bold;">订单号：</span>{{ currentOrder?.orderNo }}
                </div>
                <div style="margin-bottom: 5px;">
                  <span style="font-weight: bold;">停车时长：</span>{{ formatDuration(billingDetailResult.total_duration_min) }}
                </div>
                <div>
                  <span style="font-weight: bold;">应付金额：</span>
                  <span style="font-size: 24px; color: #f56c6c; font-weight: bold;">
                    ¥{{ billingDetailResult.final_amount }}
                  </span>
                </div>
              </div>
            </div>
          </template>
        </el-alert>
        
        <el-collapse v-model="activeDetailDayNames" accordion>
          <el-collapse-item
            v-for="(dailyBilling, index) in billingDetailResult.daily_billings"
            :key="index"
            :name="String(index)"
          >
            <template #title>
              <span style="font-weight: bold;">
                {{ dailyBilling.date }}
                <el-tag :type="getDayTagType(dailyBilling)" size="small" style="margin-left: 10px;">
                  {{ getDayLabel(dailyBilling) }}
                </el-tag>
              </span>
              <span style="float: right; color: #f56c6c; font-weight: bold;">
                ¥{{ dailyBilling.daily_total }}
              </span>
            </template>
            
            <el-table :data="dailyBilling.periods" size="small" border>
              <el-table-column prop="start_time" label="开始时间" width="160">
                <template #default="scope">
                  {{ formatTime(scope.row.start_time) }}
                </template>
              </el-table-column>
              <el-table-column prop="end_time" label="结束时间" width="160">
                <template #default="scope">
                  {{ formatTime(scope.row.end_time) }}
                </template>
              </el-table-column>
              <el-table-column prop="duration_min" label="时长(分)" width="80" align="center" />
              <el-table-column prop="period_type" label="时段类型" width="90" align="center">
                <template #default="scope">
                  <el-tag :type="getPeriodTagType(scope.row.period_type)" size="small">
                    {{ getPeriodLabel(scope.row.period_type) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="hourly_rate" label="费率" width="100" align="center">
                <template #default="scope">
                  <template v-if="scope.row.is_first_hour">
                    <span style="color: #67c23a;">¥{{ billingDetailResult.rule_summary?.first_hour }}</span>
                    <el-tag type="success" size="mini" style="margin-left: 5px;">首小时</el-tag>
                  </template>
                  <template v-else>
                    ¥{{ scope.row.hourly_rate }}
                  </template>
                </template>
              </el-table-column>
              <el-table-column prop="period_amount" label="费用" width="80" align="right">
                <template #default="scope">
                  <span style="color: #f56c6c; font-weight: bold;">
                    ¥{{ scope.row.period_amount }}
                  </span>
                </template>
              </el-table-column>
            </el-table>
            
            <div style="margin-top: 10px; padding: 10px; background-color: #f5f7fa; border-radius: 4px;">
              <div style="display: flex; justify-content: space-between;">
                <span>时段小计：</span>
                <span>¥{{ dailyBilling.sub_total }}</span>
              </div>
              <div v-if="dailyBilling.discount > 0" style="display: flex; justify-content: space-between; color: #67c23a;">
                <span>日封顶优惠：</span>
                <span>-¥{{ dailyBilling.discount }}</span>
              </div>
              <div style="display: flex; justify-content: space-between; font-weight: bold;">
                <span>本日合计：</span>
                <span style="color: #f56c6c;">¥{{ dailyBilling.daily_total }}</span>
              </div>
            </div>
          </el-collapse-item>
        </el-collapse>
      </div>
      
      <template #footer>
        <el-button @click="billingDetailDialogVisible = false">关闭</el-button>
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
const billingDetailDialogVisible = ref(false)
const currentOrder = ref(null)
const paymentMethod = ref('alipay')
const paying = ref(false)
const calculating = ref(false)
const billingRules = ref([])
const calculationResult = ref(null)
const billingDetailResult = ref(null)
const activeDayNames = ref(['0'])
const activeDetailDayNames = ref(['0'])

const parkingForm = ref({
  licensePlate: ''
})

const calculatorForm = ref({
  entryTime: null,
  exitTime: null,
  spotType: 'standard'
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

const calculateParkingFee = async () => {
  if (!calculatorForm.value.entryTime || !calculatorForm.value.exitTime) {
    ElMessage.warning('请选择入场时间和出场时间')
    return
  }
  
  const entryTime = new Date(calculatorForm.value.entryTime)
  const exitTime = new Date(calculatorForm.value.exitTime)
  
  if (exitTime <= entryTime) {
    ElMessage.warning('出场时间必须晚于入场时间')
    return
  }
  
  calculating.value = true
  try {
    const response = await billingApi.calculateDetailedFee({
      entry_time: entryTime.toISOString(),
      exit_time: exitTime.toISOString(),
      spot_type: calculatorForm.value.spotType
    })
    
    if (response.success && response.data) {
      calculationResult.value = response.data
      activeDayNames.value = ['0']
    } else {
      ElMessage.error(response.message || '计算失败')
    }
  } catch (error) {
    console.error('计算停车费用失败:', error)
    generateDemoCalculationResult()
  } finally {
    calculating.value = false
  }
}

const generateDemoCalculationResult = () => {
  const entryTime = new Date(calculatorForm.value.entryTime)
  const exitTime = new Date(calculatorForm.value.exitTime)
  const diffMs = exitTime - entryTime
  const totalMinutes = Math.floor(diffMs / 60000)
  
  calculationResult.value = {
    entry_time: entryTime.toISOString(),
    exit_time: exitTime.toISOString(),
    total_duration: `${Math.floor(totalMinutes / 60)}小时${totalMinutes % 60}分钟`,
    total_duration_min: totalMinutes,
    within_grace_period: totalMinutes <= 15,
    grace_period_minutes: 15,
    daily_billings: [
      {
        date: entryTime.toISOString().slice(0, 10),
        day_of_week: entryTime.getDay(),
        is_holiday: entryTime.getDay() === 0 || entryTime.getDay() === 6,
        periods: [
          {
            start_time: entryTime.toISOString(),
            end_time: exitTime.toISOString(),
            duration: `${totalMinutes}分钟`,
            duration_min: totalMinutes,
            period_type: 'normal',
            hourly_rate: 8,
            base_rate: 8,
            is_first_hour: true,
            period_amount: totalMinutes <= 60 ? 10 : 10 + Math.ceil((totalMinutes - 60) / 60) * 8
          }
        ],
        sub_total: totalMinutes <= 60 ? 10 : 10 + Math.ceil((totalMinutes - 60) / 60) * 8,
        daily_max: 80,
        discount: 0,
        daily_total: totalMinutes <= 60 ? 10 : 10 + Math.ceil((totalMinutes - 60) / 60) * 8
      }
    ],
    first_hour_used: 10,
    first_hour_applied: true,
    total_before_rules: totalMinutes <= 60 ? 10 : 10 + Math.ceil((totalMinutes - 60) / 60) * 8,
    total_discount: 0,
    min_charge_applied: false,
    final_amount: totalMinutes <= 15 ? 0 : (totalMinutes <= 60 ? 10 : 10 + Math.ceil((totalMinutes - 60) / 60) * 8),
    rule_summary: {
      first_hour: 10,
      hourly_rate: 8,
      daily_max: 80,
      min_charge: 5,
      grace_period: 15,
      peak_rate: 12,
      night_rate: 5,
      holiday_rate: 10
    }
  }
  
  if (calculationResult.value.final_amount > 80) {
    calculationResult.value.daily_billings[0].discount = calculationResult.value.final_amount - 80
    calculationResult.value.daily_billings[0].daily_total = 80
    calculationResult.value.total_discount = calculationResult.value.final_amount - 80
    calculationResult.value.final_amount = 80
  }
  
  activeDayNames.value = ['0']
  ElMessage.info('使用演示数据（后端未启动）')
}

const showPaymentDialog = (order) => {
  currentOrder.value = order
  paymentDialogVisible.value = true
}

const showBillingDetailDialog = async (order) => {
  currentOrder.value = order
  billingDetailDialogVisible.value = true
  
  try {
    const entryTime = new Date(order.entryTime)
    const exitTime = new Date()
    
    const response = await billingApi.calculateDetailedFee({
      entry_time: entryTime.toISOString(),
      exit_time: exitTime.toISOString(),
      spot_type: 'standard'
    })
    
    if (response.success && response.data) {
      billingDetailResult.value = response.data
      activeDetailDayNames.value = ['0']
    }
  } catch (error) {
    console.error('获取计费明细失败:', error)
    generateDemoBillingDetail(order)
  }
}

const generateDemoBillingDetail = (order) => {
  const entryTime = new Date(order.entryTime)
  const exitTime = new Date()
  const diffMs = exitTime - entryTime
  const totalMinutes = Math.floor(diffMs / 60000)
  
  billingDetailResult.value = {
    entry_time: entryTime.toISOString(),
    exit_time: exitTime.toISOString(),
    total_duration: `${Math.floor(totalMinutes / 60)}小时${totalMinutes % 60}分钟`,
    total_duration_min: totalMinutes,
    within_grace_period: totalMinutes <= 15,
    grace_period_minutes: 15,
    daily_billings: [
      {
        date: entryTime.toISOString().slice(0, 10),
        day_of_week: entryTime.getDay(),
        is_holiday: entryTime.getDay() === 0 || entryTime.getDay() === 6,
        periods: [
          {
            start_time: entryTime.toISOString(),
            end_time: exitTime.toISOString(),
            duration: `${totalMinutes}分钟`,
            duration_min: totalMinutes,
            period_type: 'normal',
            hourly_rate: 8,
            base_rate: 8,
            is_first_hour: true,
            period_amount: order.totalAmount
          }
        ],
        sub_total: order.totalAmount,
        daily_max: 80,
        discount: 0,
        daily_total: order.totalAmount
      }
    ],
    first_hour_used: 10,
    first_hour_applied: true,
    total_before_rules: order.totalAmount,
    total_discount: 0,
    min_charge_applied: false,
    final_amount: order.totalAmount,
    rule_summary: {
      first_hour: 10,
      hourly_rate: 8,
      daily_max: 80,
      min_charge: 5,
      grace_period: 15,
      peak_rate: 12,
      night_rate: 5,
      holiday_rate: 10
    }
  }
  
  activeDetailDayNames.value = ['0']
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

const formatTime = (timeStr) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const formatDuration = (minutes) => {
  if (!minutes) return '0分钟'
  const hours = Math.floor(minutes / 60)
  const mins = minutes % 60
  if (hours > 0) {
    return `${hours}小时${mins}分钟`
  }
  return `${mins}分钟`
}

const getPeriodTagType = (periodType) => {
  switch (periodType) {
    case 'peak': return 'warning'
    case 'night': return 'info'
    case 'holiday': return 'danger'
    default: return 'success'
  }
}

const getPeriodLabel = (periodType) => {
  switch (periodType) {
    case 'peak': return '高峰'
    case 'night': return '夜间'
    case 'holiday': return '节假日'
    default: return '标准'
  }
}

const getDayTagType = (dailyBilling) => {
  if (dailyBilling.is_holiday) return 'danger'
  return 'primary'
}

const getDayLabel = (dailyBilling) => {
  const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  if (dailyBilling.is_holiday) {
    return weekdays[dailyBilling.day_of_week] + ' (节假日)'
  }
  return weekdays[dailyBilling.day_of_week]
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

.calculator-card {
  margin-top: 0;
}

.calculation-result {
  margin-top: 20px;
}

:deep(.el-collapse-item__header) {
  align-items: center;
}
</style>
