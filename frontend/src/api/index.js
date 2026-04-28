import axios from 'axios'
import { ElMessage } from 'element-plus'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

api.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    if (error.response) {
      switch (error.response.status) {
        case 400:
          ElMessage.error('请求参数错误')
          break
        case 401:
          ElMessage.error('未授权，请登录')
          break
        case 404:
          ElMessage.error('请求的资源不存在')
          break
        case 500:
          ElMessage.error('服务器内部错误')
          break
        default:
          ElMessage.error(`请求失败: ${error.response.status}`)
      }
    } else {
      ElMessage.error('网络错误，请检查网络连接')
    }
    return Promise.reject(error)
  }
)

export const parkingApi = {
  getParkingSpots: () => api.get('/parking/spots'),
  getParkingSpotById: (id) => api.get(`/parking/spots/${id}`),
  updateParkingSpotStatus: (id, status) => api.put(`/parking/spots/${id}/status`, { status })
}

export const reservationApi = {
  createReservation: (data) => api.post('/reservations', data),
  getReservations: () => api.get('/reservations'),
  getReservationById: (id) => api.get(`/reservations/${id}`),
  cancelReservation: (id) => api.put(`/reservations/${id}/cancel`)
}

export const parkingLotApi = {
  getParkingLots: () => api.get('/parking-lots'),
  getParkingLotById: (id) => api.get(`/parking-lots/${id}`)
}

export const orderApi = {
  getOrders: () => api.get('/orders'),
  getOrderById: (id) => api.get(`/orders/${id}`),
  createOrder: (data) => api.post('/orders', data),
  payOrder: (id, data) => api.post(`/orders/${id}/pay`, data)
}

export const accessControlApi = {
  vehicleEntry: (data) => api.post('/access/entry', data),
  vehicleExit: (data) => api.post('/access/exit', data),
  getAccessLogs: () => api.get('/access/logs')
}

export const predictionApi = {
  getTrafficPrediction: (hours = 24) => api.get(`/prediction/traffic?hours=${hours}`),
  getParkingPrediction: (zone) => {
    const url = zone ? `/prediction/parking?zone=${zone}` : '/prediction/parking'
    return api.get(url)
  },
  getPeakPrediction: () => api.get('/prediction/peak'),
  getVacancyPrediction: (time) => {
    const url = time ? `/prediction/vacancy?time=${time}` : '/prediction/vacancy'
    return api.get(url)
  },
  trainModel: (data) => api.post('/prediction/train', data),
  getModelStatus: () => api.get('/prediction/model/status')
}

export const plateRecognitionApi = {
  recognizePlate: (data) => api.post('/plate/recognize', data),
  recognizeFromImage: (formData) => api.post('/plate/recognize', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
  verifyPlate: (licensePlate, expectedPlate) => api.post('/plate/verify', {
    license_plate: licensePlate,
    expected_plate: expectedPlate
  }),
  getRecognitionLogs: (limit = 100) => api.get(`/plate/logs?limit=${limit}`),
  validatePlate: (licensePlate) => api.post('/plate/validate', { license_plate: licensePlate })
}

export const anomalyDetectionApi = {
  detectAnomaly: (data) => api.post('/anomaly/detect', data),
  checkOvertime: (thresholdMinutes) => {
    const url = thresholdMinutes ? `/anomaly/overtime?threshold=${thresholdMinutes}` : '/anomaly/overtime'
    return api.get(url)
  },
  checkWrongSpot: (licensePlate, currentSpotId, reservedSpotId) => api.post('/anomaly/wrong_spot', {
    license_plate: licensePlate,
    current_spot_id: currentSpotId,
    reserved_spot_id: reservedSpotId
  }),
  getSuspiciousVehicles: () => api.get('/anomaly/suspicious'),
  getAnomalyLogs: (limit = 100, type) => {
    let url = `/anomaly/logs?limit=${limit}`
    if (type) {
      url += `&type=${type}`
    }
    return api.get(url)
  },
  checkAllAnomalies: (data) => api.post('/anomaly/check_all', data)
}

export const deviceApi = {
  getDevices: () => api.get('/devices'),
  getDeviceById: (id) => api.get(`/devices/${id}`),
  controlDevice: (id, action) => api.post(`/devices/${id}/control`, { action }),
  getDeviceStatus: (id) => api.get(`/devices/${id}/status`)
}

export const billingApi = {
  calculateFee: (data) => api.post('/billing/calculate', data),
  calculateDetailedFee: (data) => api.post('/billing/calculate-detailed', data),
  getBillingRules: () => api.get('/billing/rules'),
  updateBillingRule: (id, data) => api.put(`/billing/rules/${id}`, data)
}

export default api
