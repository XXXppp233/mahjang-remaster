<template>
    <div>
        <b v-for="tile in props.discarded">{{ tile }}</b>
    </div>
</template>

<script setup>

const props = defineProps({
  discarded: {
    type: Array,
    default: () => []
  }
})

</script>

<style scoped>
div {
    caret-color: transparent; /* 隐藏光标 */
    box-sizing: border-box;
    margin: 0;
    height: 100%;
    width: 100%;
    min-height: 8vh;
    font-size: 3vh;
    font-family: 'Segoe UI Symbol';
    display: block;
    word-break: break-word;
    line-height: 1.15;
    white-space: normal;
    overflow-y: auto;
}

b {
    font-weight: normal;
}

b:hover {
    cursor: pointer;
    color: #f44336;
}
</style>
