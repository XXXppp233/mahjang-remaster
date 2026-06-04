<template>
  <div class="fail">
    <img class="fail-img" :src="`images/input.gif`" alt="" />
    <h1>无法连接到游戏伺服器</h1>
    <p>可能是伺服器未启动，可以尝试通过以下方式联系伺服器管理员或刷新页面</p>
    <ul>
      <li><a href="mailto:chenmld233@gmail.com">chenmld233@gmail.com</a></li>
    </ul>
  </div>
</template>
<script setup></script>

<style scoped>
.fail {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  text-align: center;
}
.fail-img {
  width: 20vw;
}
a {
  text-decoration: none;
  color: inherit;
}
</style>
