<template>
  <div class="org-select">
    <div class="org-menu">
      <button
        v-for="org in orgs"
        :key="org"
        type="button"
        class="org-option"
        :class="{ active: org === status.tempOrg }"
        @click="selectOrg(org)"
      >
        {{ org }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, watch } from 'vue'
import { statusStore } from '@/stores/status'

const status = statusStore()
const orgs = computed(() => status.getCharacterGroups())

function selectOrg(org) {
  status.setTempOrg(org)
}

function ensureOrg(list) {
  if (list.length && (!status.tempOrg || !list.includes(status.tempOrg))) {
    status.setTempOrg(list[0])
  }
}

onMounted(() => ensureOrg(orgs.value))
watch(orgs, ensureOrg)
</script>

<style scoped>
.org-select {
  width: 100%;
}

.org-menu {
  display: flex;
  flex-wrap: wrap;
  align-content: flex-start;
  align-items: flex-start;
  justify-content: flex-start;
  gap: 0.8vh 0.8vw;
  padding: 1vh;
  background: #fafafa;
  max-height: 12vh;
  overflow-y: auto;
}

.org-option {
  flex: 0 0 auto;
  padding: 0.8vh 0.6vw;
  width: auto;
  max-width: 100%;
  border: 0.3vh solid #111;
  border-radius: 0.3vh;
  background: #fff;
  box-shadow: 0.3vh 0.3vh 0 #111;
  color: #111;
  font-size: 1.8vh;
  font-weight: 800;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  cursor: pointer;
}

.org-option:hover,
.org-option.active {
  background: #ffeb3b;
}
</style>
