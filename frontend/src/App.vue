<template>
  <div id="app">
    <el-container style="height: 100vh;">
      <el-header class="header">
        <div class="logo">
          <el-icon :size="30" color="#409EFF">
            <Car />
          </el-icon>
          <span class="title">智慧停车管理平台</span>
        </div>
        <el-menu 
          :default-active="activeMenu" 
          mode="horizontal" 
          background-color="#545c64" 
          text-color="#fff" 
          active-text-color="#ffd04b"
          router
        >
          <el-menu-item index="/parking-map">
            <template #title>
              <el-icon><MapLocation /></el-icon>
              车位地图
            </template>
          </el-menu-item>
          <el-menu-item index="/reservation">
            <template #title>
              <el-icon><Calendar /></el-icon>
              车位预约
            </template>
          </el-menu-item>
          <el-menu-item index="/payment">
            <template #title>
              <el-icon><CreditCard /></el-icon>
              在线缴费
            </template>
          </el-menu-item>
          <el-menu-item index="/orders">
            <template #title>
              <el-icon><List /></el-icon>
              订单管理
            </template>
          </el-menu-item>
          <el-menu-item index="/traffic-prediction">
            <template #title>
              <el-icon><TrendCharts /></el-icon>
              流量预测
            </template>
          </el-menu-item>
        </el-menu>
      </el-header>
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const activeMenu = computed(() => route.path)
</script>

<style>
#app {
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

.header {
  display: flex;
  align-items: center;
  background-color: #545c64;
  padding: 0 20px;
}

.logo {
  display: flex;
  align-items: center;
  margin-right: 20px;
}

.title {
  color: white;
  font-size: 20px;
  font-weight: bold;
  margin-left: 10px;
}

.main-content {
  background-color: #f0f2f5;
  padding: 20px;
}
</style>
