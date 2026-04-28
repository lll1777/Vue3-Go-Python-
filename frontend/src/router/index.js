import { createRouter, createWebHistory } from 'vue-router'
import ParkingMap from '../views/ParkingMap.vue'
import Reservation from '../views/Reservation.vue'
import Payment from '../views/Payment.vue'
import Orders from '../views/Orders.vue'
import TrafficPrediction from '../views/TrafficPrediction.vue'

const routes = [
  {
    path: '/',
    redirect: '/parking-map'
  },
  {
    path: '/parking-map',
    name: 'ParkingMap',
    component: ParkingMap
  },
  {
    path: '/reservation',
    name: 'Reservation',
    component: Reservation
  },
  {
    path: '/payment',
    name: 'Payment',
    component: Payment
  },
  {
    path: '/orders',
    name: 'Orders',
    component: Orders
  },
  {
    path: '/traffic-prediction',
    name: 'TrafficPrediction',
    component: TrafficPrediction
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
